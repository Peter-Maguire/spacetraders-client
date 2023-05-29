package routine

import "spacetraders/entity"

type NegotiateContract struct {
}

func (n NegotiateContract) Run(state *State) RoutineResult {
	_ = state.Ship.EnsureNavState(entity.NavDocked)
	newContract, err := state.Ship.NegotiateContract()

	// TODO: contract logic

	if err == nil {
		state.Log("New contract get")
		_ = newContract.Accept()
		state.Contract = newContract
		state.FireEvent("newContract", newContract)
	} else {
		state.Log(err.Error())
	}

	return RoutineResult{SetRoutine: DetermineObjective{}, WaitSeconds: 5}
}

func (n NegotiateContract) Name() string {
	return "Negotiate Contract"
}
