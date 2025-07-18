package routine

import (
	"spacetraders/entity"
	"spacetraders/http"
)

type NegotiateContract struct {
}

func (n NegotiateContract) Run(state *State) RoutineResult {
	_ = state.Ship.EnsureNavState(state.Context, entity.NavDocked)
	newContract, err := state.Ship.NegotiateContract(state.Context)

	// TODO: contract logic

	if err == nil {
		state.Log("New contract get")
		_ = newContract.Accept(state.Context)
		state.Contract = newContract
		state.FireEvent("newContract", newContract)
	} else {

		switch err.Code {
		case http.ErrNoFactionPresence:
			return RoutineResult{SetRoutine: GoToRandomFactionWaypoint{next: n}}
		}
		state.Log(err.Error())
	}

	return RoutineResult{SetRoutine: DetermineObjective{}, WaitSeconds: 5}
}

func (n NegotiateContract) Name() string {
	return "Negotiate Contract"
}
