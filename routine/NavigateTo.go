package routine

import (
	"fmt"
	"spacetraders/entity"
	"spacetraders/http"
	"time"
)

func NavigateTo(waypoint entity.Waypoint, nextState Routine) func(state *entity.State, ship *entity.Ship) RoutineResult {
	return func(state *entity.State, ship *entity.Ship) RoutineResult {
		fmt.Println("Navigating to ", waypoint)

		_ = ship.EnsureNavState(entity.NavOrbit)

		nav, err := ship.Navigate(waypoint)
		if err != nil {
			switch err.Code {
			case http.ErrInsufficientFuelForNav:
				fmt.Println("Refuelling and trying again")
				_ = ship.EnsureNavState(entity.NavDocked)
				_ = ship.Refuel()
				return RoutineResult{}
			}
			fmt.Println("Unknown error ", err.Data)
		}

		waitingTime := nav.Route.Arrival.Sub(time.Now())
		fmt.Printf("Arriving at %s in %.f seconds", waypoint, waitingTime.Seconds())

		return RoutineResult{
			WaitSeconds: int(waitingTime.Seconds()),
			SetRoutine:  nextState,
		}
	}
}
