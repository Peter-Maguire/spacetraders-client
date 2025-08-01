package routine

import (
	"fmt"
	"time"
)

type AwaitRescue struct {
	next           Routine
	startedWaiting *time.Time
}

func (a AwaitRescue) Run(state *State) RoutineResult {
	if a.startedWaiting == nil {
		now := time.Now()
		a.startedWaiting = &now
	}
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

	if a.startedWaiting.Sub(time.Now()) > time.Hour {
		state.Log("We've been waiting for over an hour...")
		return RoutineResult{
			SetRoutine: DetermineObjective{},
		}
	}

	return RoutineResult{
		WaitSeconds: 60,
	}
}

func (a AwaitRescue) Name() string {

	return fmt.Sprintf("Await Rescue -> %s", a.next.Name())
}
