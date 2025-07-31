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
			leastUnits := 1000
			leastUnitsSymbol := ""
			for _, i := range state.Ship.Cargo.Inventory {
				if i.Units < leastUnits {
					leastUnits = i.Units
					leastUnitsSymbol = i.Symbol
				}
			}
			state.Ship.JettisonCargo(state.Context, leastUnitsSymbol, 1)
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
