package routine

import (
	"fmt"
	"spacetraders/entity"
	"spacetraders/http"
	"time"
)

func NavigateTo(waypoint entity.Waypoint, nextState Routine) Routine {
	return func(state *State) RoutineResult {
		fmt.Println("Navigating to ", waypoint)

		_ = state.Ship.EnsureNavState(entity.NavOrbit)

		nav, err := state.Ship.Navigate(waypoint)
		if err != nil {
			switch err.Code {
			case http.ErrInsufficientFuelForNav:
				state.Log("Refuelling and trying again")
				_ = state.Ship.EnsureNavState(entity.NavDocked)
				_ = state.Ship.Refuel()
				return RoutineResult{}
			}
			fmt.Println("Unknown error ", err)
		}

		waitingTime := nav.Route.Arrival.Sub(time.Now())
		fmt.Printf("Arriving at %s in %.f seconds", waypoint, waitingTime.Seconds())

		return RoutineResult{
			WaitSeconds: int(waitingTime.Seconds()),
			SetRoutine:  nextState,
		}
	}
}
