package routine

import (
	"fmt"
	"spacetraders/entity"
)

func Refuel(nextState Routine) func(state *entity.State, ship *entity.Ship) RoutineResult {
	return func(state *entity.State, ship *entity.Ship) RoutineResult {
		fmt.Println("Refuelling...")
		_ = ship.EnsureNavState(entity.NavDocked)
		_ = ship.Refuel()
		return RoutineResult{SetRoutine: nextState}
	}
}
