package routine

import (
	"fmt"
	"spacetraders/entity"
)

type GoToSystem struct {
	system string
	next   Routine
}

func (g GoToSystem) Run(state *State) RoutineResult {
	_ = state.Ship.EnsureNavState(entity.NavOrbit)
	jumpResult, err := state.Ship.Jump(g.system)

	if err == nil {
		waitUntil := jumpResult.Cooldown.Expiration
		return RoutineResult{
			WaitUntil:  &waitUntil,
			SetRoutine: g.next,
		}
	}

	state.Log("Unable to jump")
	state.Log(err.Error())
	return RoutineResult{Stop: true}
}

func (g GoToSystem) Name() string {
	return fmt.Sprintf("Go To System %s", g.system)
}
