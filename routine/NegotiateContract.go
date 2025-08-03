package routine

import (
	"fmt"
	"spacetraders/entity"
	"spacetraders/http"
)

type NegotiateContract struct {
}

func (n NegotiateContract) Run(state *State) RoutineResult {

	contracts, _ := state.Agent.Contracts(state.Context)

	for _, c := range *contracts {
		if !c.Fulfilled {
			state.Contract = &c
			state.Log("We haven't finished this contract yet")
			return RoutineResult{SetRoutine: DetermineObjective{}}
		}
	}

	_ = state.Ship.EnsureNavState(state.Context, entity.NavDocked)
	newContract, err := state.Ship.NegotiateContract(state.Context)

	// TODO: contract logic

	if err == nil {
		state.Log("New contract get")
		for _, term := range newContract.Terms.Deliver {
			state.Log(fmt.Sprintf("We are delivering %dx %s to %s for %d credits", term.UnitsRequired, term.TradeSymbol, term.DestinationSymbol, newContract.Terms.Payment.GetTotalPayment()))
		}
		_ = newContract.Accept(state.Context)
		state.Contract = newContract
		state.FireEvent("newContract", newContract)
	} else {

		switch err.Code {
		case http.ErrWaypointNoFaction:
			return RoutineResult{SetRoutine: GoToRandomFactionWaypoint{next: n}}
		}
		state.Log(err.Error())
	}

	return RoutineResult{SetRoutine: DetermineObjective{}, WaitSeconds: 5}
}

func (n NegotiateContract) Name() string {
	return "Negotiate Contract"
}
