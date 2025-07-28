package routine

import (
	"fmt"
)

type AwaitRescue struct {
	next Routine
}

func (a AwaitRescue) Run(state *State) RoutineResult {
	state.Ship.EnsureNavState(state.Context, "ORBIT")

	cargo, _ := state.Ship.GetCargo(state.Context)

	fuelSlot := cargo.GetSlotWithItem("FUEL")
	if fuelSlot == nil {
		if state.Ship.Cargo.IsFull() {
			return RoutineResult{
				SetRoutine: Jettison{
					nextIfSuccessful: a,
					nextIfFailed:     a,
				},
			}
		}

		return RoutineResult{
			WaitSeconds: 60,
		}
	}

	err := state.Ship.RefuelFromCargo(state.Context, fuelSlot.Units)
	if err != nil {
		state.Log("Refuel failed: " + err.Message)
		return RoutineResult{
			WaitSeconds: 60,
		}
	}

	return RoutineResult{
		WaitSeconds: 60,
	}
}

func (a AwaitRescue) Name() string {

	return fmt.Sprintf("Await Rescue -> %s", a.next.Name())
}
