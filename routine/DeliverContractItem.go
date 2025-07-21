package routine

import (
	"fmt"
	"spacetraders/entity"
	"spacetraders/metrics"
)

type DeliverContractItem struct {
	item string
	next Routine
}

func (r DeliverContractItem) Run(state *State) RoutineResult {

	deliverable := state.Contract.Terms.GetDeliverable(r.item)
	if deliverable == nil {
		state.Log("Item specified is not in the contract as a deliverable")
		return RoutineResult{
			SetRoutine: r.next,
		}
	}

	if state.Ship.Nav.WaypointSymbol != deliverable.DestinationSymbol {
		state.Log("Going to contract location")
		return RoutineResult{
			SetRoutine: NavigateTo{
				waypoint: deliverable.DestinationSymbol,
				next:     r,
			},
		}
	}

	_ = state.Ship.EnsureNavState(state.Context, entity.NavDocked)

	slot := state.Ship.Cargo.GetSlotWithItem(r.item)
	state.Log(fmt.Sprintf("Deliver %dx %s", slot.Units, r.item))
	deliverResult, err := state.Contract.Deliver(state.Context, state.Ship.Symbol, r.item, slot.Units)

	// Update the cargo
	_, _ = state.Ship.GetCargo(state.Context)

	if err != nil {
		state.Log(fmt.Sprintf("Error delivering contract: %s", err))
		return RoutineResult{SetRoutine: r.next}
	} else {
		metrics.ContractProgress.Set(float64(deliverResult.Contract.Terms.Deliver[0].UnitsFulfilled))
		metrics.ContractRequirement.Set(float64(deliverResult.Contract.Terms.Deliver[0].UnitsRequired))
	}
	
	if deliverable.UnitsFulfilled >= deliverable.UnitsRequired {
		state.Log("Contract completed")
		err := state.Contract.Fulfill(state.Context)
		if err == nil {
			state.FireEvent("contractComplete", nil)
		}
		state.Log(fmt.Sprintf("Contract fulfill err: %s", err))
	}

	return RoutineResult{
		SetRoutine: r.next,
	}
}

func (r DeliverContractItem) Name() string {
	return fmt.Sprintf("Deliver Contract Item (%s)", r.item)
}
