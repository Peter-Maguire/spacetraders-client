package routine

import (
	"spacetraders/entity"
)

func Refuel(nextRoutine Routine) Routine {
	return func(state *State) RoutineResult {
		state.Log("Refuelling...")
		_ = state.Ship.EnsureNavState(entity.NavDocked)
		_ = state.Ship.Refuel()
		return RoutineResult{SetRoutine: nextRoutine}
	}
}
