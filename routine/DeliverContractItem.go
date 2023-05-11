package routine

import (
	"fmt"
	"os"
	"spacetraders/entity"
)

func DeliverContractItem(item string, returnTo entity.Waypoint) Routine {
	return func(state *entity.State) RoutineResult {

		_ = state.Ship.EnsureNavState(entity.NavDocked)

		slot := state.Ship.Cargo.GetSlotWithItem(item)
		fmt.Printf("Deliver %dx %s\n", slot.Units, item)
		deliverResult, err := state.Contract.Deliver(state.Ship.Symbol, item, slot.Units)
		if err != nil {
			fmt.Println(err)
		}

		deliverable := deliverResult.Contract.Terms.GetDeliverable(item)

		if deliverable.UnitsFulfilled >= deliverable.UnitsRequired {
			fmt.Println("Contract completed")
			err := state.Contract.Fulfill()
			fmt.Println(err)
			// TODO: new contract
			os.Exit(1)
		}

		return RoutineResult{
			SetRoutine: NavigateTo(returnTo, GetSurvey),
		}
	}
}
