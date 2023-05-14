package routine

import (
	"fmt"
	"spacetraders/entity"
	"spacetraders/http"
)

type NavigateTo struct {
	waypoint entity.Waypoint
	next     Routine
}

func (n NavigateTo) Run(state *State) RoutineResult {
	state.Log(fmt.Sprint("Navigating to ", n.waypoint))

	if state.Ship.Nav.WaypointSymbol == n.waypoint && state.Ship.Nav.Status != "IN_TRANSIT" {
		state.Log("We're already at our destination")
		return RoutineResult{
			SetRoutine: n.next,
		}
	}

	if state.Ship.Nav.Status == "IN_TRANSIT" && state.Ship.Nav.Route.Destination.Symbol == n.waypoint {
		state.Log("We're already on our way there")
		return RoutineResult{
			WaitUntil:  &state.Ship.Nav.Route.Arrival,
			SetRoutine: n.next,
		}
	}

	_ = state.Ship.EnsureNavState(entity.NavOrbit)

	_, err := state.Ship.Navigate(n.waypoint)
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
				SetRoutine: n.next,
			}
		}
		state.Log(fmt.Sprintf("Unknown error: %s", err))
		return RoutineResult{
			WaitSeconds: 10,
		}
	}

	return RoutineResult{
		WaitUntil:  &state.Ship.Nav.Route.Arrival,
		SetRoutine: n.next,
	}
}

func (n NavigateTo) Name() string {
	return fmt.Sprintf("Navigate to %s", n.waypoint)
}
