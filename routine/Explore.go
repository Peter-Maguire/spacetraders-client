package routine

import (
    "fmt"
    "spacetraders/database"
    "spacetraders/entity"
    "spacetraders/metrics"
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
        database.StoreShipCosts(shipyardData)
    }

    if waypointData.HasTrait("MARKETPLACE") {
        state.Log("There's a marketplace here")
        marketData, _ = waypointData.Symbol.GetMarket()
        database.StoreMarketRates(waypointData, marketData.TradeGoods)
        fuelTrader := marketData.GetTradeGood("FUEL")
        if fuelTrader != nil && state.Ship.Fuel.Current < state.Ship.Fuel.Capacity/2 {
            state.Log("Refuelling here")
            _ = state.Ship.EnsureNavState(entity.NavDocked)
            _ = state.Ship.Refuel()
        }

        antiMatterTrader := marketData.GetTradeGood("ANTIMATTER")
        if antiMatterTrader != nil && state.Ship.Cargo.GetSlotWithItem("ANTIMATTER").Units < 2 {
            state.Log("Buying some antimatter")
            _ = state.Ship.EnsureNavState(entity.NavDocked)
            res, _ := state.Ship.Purchase("ANTIMATTER", 5)
            if res != nil {
                state.Log("Success")
                fmt.Printf("Credits set to %d", res.Agent.Credits)
                metrics.NumCredits.Set(float64(res.Agent.Credits))
            }
        }
    }

    database.VisitWaypoint(waypointData)

    return RoutineResult{}
}

func (e Explore) Name() string {
    return "Explore"
}
