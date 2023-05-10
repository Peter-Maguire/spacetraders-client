package routine

import (
	"fmt"
	"os"
	"spacetraders/entity"
)

func DeliverContractItem(item string, returnTo entity.Waypoint) func(state *entity.State, ship *entity.Ship) RoutineResult {
	return func(state *entity.State, ship *entity.Ship) RoutineResult {

		_ = ship.EnsureNavState(entity.NavDocked)

		slot := ship.Cargo.GetSlotWithItem(item)
		fmt.Printf("Deliver %dx %s\n", slot.Units, item)
		deliverResult, err := state.Contract.Deliver(ship.Symbol, item, slot.Units)
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
