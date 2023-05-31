package routine

import (
	"fmt"
	"spacetraders/entity"
	"spacetraders/http"
	"time"
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

	if state.Ship.Nav.Status == "IN_TRANSIT" && state.Ship.Nav.Route.Arrival.After(time.Now()) {
		state.Log("We're on our way somewhere")
		return RoutineResult{
			WaitUntil: &state.Ship.Nav.Route.Arrival,
		}
	}

	if state.Ship.Nav.SystemSymbol != n.waypoint.GetSystemName() {
		state.Log("Jumping to system first")
		return RoutineResult{
			SetRoutine: GoToSystem{next: n, system: n.waypoint.GetSystemName()},
		}
	}

	_ = state.Ship.EnsureNavState(entity.NavOrbit)

	_, err := state.Ship.Navigate(n.waypoint)
	if err != nil {
		switch err.Code {
		case http.ErrCooldown:
			return RoutineResult{
				WaitSeconds: int(err.Data["cooldown"].(map[string]any)["remainingSeconds"].(float64)),
			}
		case http.ErrInsufficientFuelForNav:
			state.Log(err.Message)
			state.Log("Refuelling and trying again")
			return RoutineResult{
				SetRoutine: Refuel{next: n},
			}
		case http.ErrShipInTransit, http.ErrNavigateInTransit:
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
