package routine

import (
	"fmt"
	"spacetraders/entity"
)

func Refuel(nextRoutine Routine) Routine {
	return func(state *entity.State) RoutineResult {
		fmt.Println("Refuelling...")
		_ = state.Ship.EnsureNavState(entity.NavDocked)
		_ = state.Ship.Refuel()
		return RoutineResult{SetRoutine: nextRoutine}
	}
}
