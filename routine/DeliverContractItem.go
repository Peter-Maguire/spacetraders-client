package routine

import (
	"fmt"
	"os"
	"spacetraders/entity"
)

func DeliverContractItem(item string, returnTo entity.Waypoint) Routine {
	return func(state *State) RoutineResult {

		_ = state.Ship.EnsureNavState(entity.NavDocked)

		slot := state.Ship.Cargo.GetSlotWithItem(item)
		state.Log(fmt.Sprintf("Deliver %dx %s", slot.Units, item))
		deliverResult, err := state.Contract.Deliver(state.Ship.Symbol, item, slot.Units)
		if err != nil {
			state.Log(fmt.Sprintf("Error delivering contract: %s", err))
			os.Exit(1)
		}

		deliverable := deliverResult.Contract.Terms.GetDeliverable(item)

		if deliverable.UnitsFulfilled >= deliverable.UnitsRequired {
			state.Log("Contract completed")
			err := state.Contract.Fulfill()
			state.Log(fmt.Sprintf("Contract fulfill err: %s", err))
			// TODO: new contract
			os.Exit(1)
		}

		return RoutineResult{
			SetRoutine: NavigateTo(returnTo, GetSurvey),
		}
	}
}
