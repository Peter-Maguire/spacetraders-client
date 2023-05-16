package routine

import (
	"fmt"
	"spacetraders/database"
	"spacetraders/entity"
)

type Explore struct {
}

func (e Explore) Run(state *State) RoutineResult {
	_ = state.Ship.EnsureNavState(entity.NavOrbit)
	visited := database.GetWaypoint(state.Ship.Nav.WaypointSymbol)
	if visited != nil {
		return RoutineResult{
			SetRoutine: FindNewWaypoint{},
		}
	}

	state.Log(fmt.Sprintf("Checking out %s", state.Ship.Nav.WaypointSymbol))
	waypointData, _ := state.Ship.Nav.WaypointSymbol.GetWaypointData()

	var shipyardData *entity.ShipyardStock
	var marketData *entity.Market

	if waypointData.HasTrait("UNCHARTED") {
		data, err := state.Ship.Chart()
		if err == nil {
			state.Log("Charted waypoint")
			waypointData = data.Waypoint
		}
	}

	if waypointData.HasTrait("SHIPYARD") {
		state.Log("There's a shipyard here")
		shipyardData, _ = waypointData.Symbol.GetShipyard()
	}

	if waypointData.HasTrait("MARKETPLACE") {
		state.Log("There's a marketplace here")
		marketData, _ = waypointData.Symbol.GetMarket()
		fuelTrader := marketData.GetTradeGood("FUEL")
		if fuelTrader != nil && state.Ship.Fuel.Current < state.Ship.Fuel.Capacity/2 {
			state.Log("Refuelling here")
			_ = state.Ship.EnsureNavState(entity.NavDocked)
			_ = state.Ship.Refuel()
		}

		//if state.Ship.Cargo.GetSlotWithItem("ANTIMATTER").Units < 2 && marketData.GetTradeGood("ANTIMATTER") != nil {
		//
		//}
	}

	database.VisitWaypoint(waypointData, marketData, shipyardData)

	return RoutineResult{}
}

func (e Explore) Name() string {
	return "Explore"
}
