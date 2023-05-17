package routine

import (
    "spacetraders/entity"
    "spacetraders/http"
)

type Refuel struct {
    next           Routine
    hasTriedMarket bool
}

func (r Refuel) Run(state *State) RoutineResult {
    if state.Ship.Fuel.Current > 20 && !r.hasTriedMarket {

        market, err := state.Ship.Nav.WaypointSymbol.GetMarket()

        if err != nil || market.GetTradeGood("FUEL") == nil {
            waypoints, _ := state.Ship.Nav.WaypointSymbol.GetSystemWaypoints()
            for _, waypoint := range *waypoints {
                if waypoint.HasTrait("MARKETPLACE") && waypoint.Symbol != state.Ship.Nav.WaypointSymbol {
                    state.Log("Trying a different market")
                    r.hasTriedMarket = true
                    return RoutineResult{
                        SetRoutine: NavigateTo{
                            waypoint: waypoint.Symbol,
                            next:     r,
                        },
                    }
                }
            }
        } else {

            _ = state.Ship.EnsureNavState(entity.NavDocked)
            refuelErr := state.Ship.Refuel()

            if refuelErr == nil {
                return RoutineResult{
                    SetRoutine: r.next,
                }
            }

            switch refuelErr.Code {
            case http.ErrShipInTransit, http.ErrNavigateInTransit:
                state.Log("Ship in transit")
                return RoutineResult{
                    WaitSeconds: 30,
                }
            }
        }
    }

    state.Log("Cannot refuel")

    if state.Ship.Nav.FlightMode == "DRIFT" {
        state.Log("We're boned")
        return RoutineResult{
            Stop: true,
        }
    }

    state.Log("Setting flight mode to drift")
    _ = state.Ship.SetFlightMode("DRIFT")

    return RoutineResult{
        SetRoutine: r.next,
    }

}

func (r Refuel) Name() string {
    return "Refuel"
}
