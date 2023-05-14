package routine

import (
	"fmt"
	"os"
	"spacetraders/entity"
)

type DeliverContractItem struct {
	item     string
	returnTo entity.Waypoint
}

func (r DeliverContractItem) Run(state *State) RoutineResult {
	_ = state.Ship.EnsureNavState(entity.NavDocked)

	slot := state.Ship.Cargo.GetSlotWithItem(r.item)
	state.Log(fmt.Sprintf("Deliver %dx %s", slot.Units, r.item))
	deliverResult, err := state.Contract.Deliver(state.Ship.Symbol, r.item, slot.Units)
	if err != nil {
		state.Log(fmt.Sprintf("Error delivering contract: %s", err))
		os.Exit(1)
	}

	deliverable := deliverResult.Contract.Terms.GetDeliverable(r.item)

	if deliverable.UnitsFulfilled >= deliverable.UnitsRequired {
		state.Log("Contract completed")
		err := state.Contract.Fulfill()
		if err == nil {
			state.FireEvent("contractComplete", nil)
		}
		state.Log(fmt.Sprintf("Contract fulfill err: %s", err))
		os.Exit(1)
	}

	return RoutineResult{
		SetRoutine: NavigateTo{r.returnTo, GetSurvey{}},
	}
}

func (r DeliverContractItem) Name() string {
	return fmt.Sprintf("Deliver Contract Item (%s)", r.item)
}
