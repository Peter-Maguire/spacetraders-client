package routine

import (
	"fmt"
	"spacetraders/constant"
	"spacetraders/database"
	"spacetraders/entity"
)

type Explore struct {
	desiredTrait  constant.WaypointTrait
	marketTargets []string
	oneShot       bool
	visitVisited  bool
	next          Routine
}

func (e Explore) Run(state *State) RoutineResult {
	_ = state.Ship.EnsureNavState(state.Context, entity.NavOrbit)
	dbWaypoint := database.GetWaypoint(state.Ship.Nav.WaypointSymbol)
	if !e.visitVisited && dbWaypoint != nil && dbWaypoint.FirstVisited.Unix() > 0 {
		state.Log("We've already explored this waypoint")
		return RoutineResult{
			SetRoutine: FindNewWaypoint{
				visitVisited: e.visitVisited,
				desiredTrait: e.desiredTrait,
				next:         e,
			},
		}
	}

	state.Log(fmt.Sprintf("Checking out %s", state.Ship.Nav.WaypointSymbol))
	system, _ := state.Ship.Nav.WaypointSymbol.GetSystem(state.Context)
	waypointData, _ := state.Ship.Nav.WaypointSymbol.GetWaypointData(state.Context)

	var shipyardData *entity.ShipyardStock
	var marketData *entity.Market

	defer func() {
		state.Log("Logging visit")
		database.VisitWaypoint(waypointData, marketData, shipyardData)
	}()

	if waypointData.HasTrait(constant.TraitUncharted) {
		data, err := state.Ship.Chart(state.Context)
		if err == nil {
			state.Log("Charted waypoint")
			waypointData = data.Waypoint
		}
	}

	if waypointData.HasTrait(constant.TraitShipyard) {
		state.Log("There's a shipyard here")
		var err error
		shipyardData, err = waypointData.Symbol.GetShipyard(state.Context)
		if err != nil {
			fmt.Println("shipyard error", err)
		} else {
			fmt.Println("Storing shipyard data")
			fmt.Println(shipyardData)
			database.StoreShipCosts(shipyardData)
		}

	}

	if waypointData.HasTrait(constant.TraitMarketplace) {
		state.Log("There's a marketplace here")
		marketData, _ = waypointData.Symbol.GetMarket(state.Context)
		// TODO: Market rates should include IMPORT, EXPORT and EXCHANGE, not just whatever is going on here
		database.StoreMarketRates(system, waypointData, marketData.TradeGoods)
		database.StoreMarketExchange(system, waypointData, "export", marketData.Exports)
		database.StoreMarketExchange(system, waypointData, "import", marketData.Imports)
		database.StoreMarketExchange(system, waypointData, "exchange", marketData.Exchange)
		fuelTrader := marketData.GetTradeGood("FUEL")
		if fuelTrader != nil && state.Ship.Fuel.Current < state.Ship.Fuel.Capacity/2 {
			state.Log("Refuelling here")
			_ = state.Ship.EnsureNavState(state.Context, entity.NavDocked)
			rr, _ := state.Ship.Refuel(state.Context)
			if rr != nil {
				database.LogTransaction("explore_refuel", rr.Transaction)
			}
		}

		antiMatterTrader := marketData.GetTradeGood("ANTIMATTER")
		antiMatterSlot := state.Ship.Cargo.GetSlotWithItem("ANTIMATTER")
		fmt.Println("antimatter", antiMatterSlot, antiMatterTrader)
		//if antiMatterTrader != nil && (antiMatterSlot == nil || antiMatterSlot.Units < 2) {
		//	state.Log("Buying some antimatter")
		//	_ = state.Ship.EnsureNavState(state.Context, entity.NavDocked)
		//	res, _ := state.Ship.Purchase(state.Context, "ANTIMATTER", 5)
		//	if res != nil {
		//		state.Log("Success")
		//		metrics.NumCredits.WithLabelValues(state.Agent.Symbol).Set(float64(res.Agent.Credits))
		//	}
		//}

		if e.marketTargets != nil {
			for _, good := range marketData.TradeGoods {
				if good.SellPrice > 0 {
					for _, target := range e.marketTargets {
						if good.Symbol == target {
							state.Log(fmt.Sprintf("Found market target %s", good.Symbol))
							return RoutineResult{
								SetRoutine: e.next,
							}
						}
					}
				}
			}

		}

	}

	if e.oneShot {
		return RoutineResult{
			SetRoutine: e.next,
		}
	}
	return RoutineResult{}
}

func (e Explore) Name() string {
	name := "Explore"

	if e.desiredTrait != "" {
		name += fmt.Sprintf(" (Find %s)", e.desiredTrait)
	}
	if e.marketTargets != nil {
		name += " (Find Market)"
	}

	if e.next != nil {
		return fmt.Sprintf("%s -> %s", name, e.next.Name())
	}

	return name
}
