package routine

import (
	"fmt"
	"spacetraders/entity"
	"spacetraders/http"
	"time"
)

func NavigateTo(waypoint entity.Waypoint, nextState Routine) Routine {
	return func(state *State) RoutineResult {
		state.Log(fmt.Sprint("Navigating to ", waypoint))

		if state.Ship.Nav.WaypointSymbol == waypoint && state.Ship.Nav.Status != "IN_TRANSIT" {
			state.Log("We're already at our destination")
			return RoutineResult{
				SetRoutine: nextState,
			}
		}

		if state.Ship.Nav.Status == "IN_TRANSIT" && state.Ship.Nav.Route.Destination.Symbol == waypoint {
			state.Log("We're already on our way there")
			waitingTime := state.Ship.Nav.Route.Arrival.Sub(time.Now())
			return RoutineResult{
				WaitSeconds: int(waitingTime.Seconds()),
				SetRoutine:  nextState,
			}
		}

		_ = state.Ship.EnsureNavState(entity.NavOrbit)

		nav, err := state.Ship.Navigate(waypoint)
		if err != nil {
			switch err.Code {
			case http.ErrInsufficientFuelForNav:
				state.Log("Refuelling and trying again")
				_ = state.Ship.EnsureNavState(entity.NavDocked)
				_ = state.Ship.Refuel()
				return RoutineResult{}
			case http.ErrShipInTransit:
				state.Log("Ship in transit")
				return RoutineResult{
					WaitSeconds: 30,
				}
			case http.ErrShipAtDestination:
				state.Log("Oh we're already there")
				return RoutineResult{
					SetRoutine: nextState,
				}
			}
			state.Log(fmt.Sprintf("Unknown error: %s", err))
			return RoutineResult{
				WaitSeconds: 10,
			}
		}

		waitingTime := nav.Route.Arrival.Sub(time.Now())
		state.Log(fmt.Sprintf("Arriving at %s in %.f seconds", waypoint, waitingTime.Seconds()))

		return RoutineResult{
			WaitSeconds: int(waitingTime.Seconds()),
			SetRoutine:  nextState,
		}
	}
}
