package routine

import (
	"fmt"
	"spacetraders/database"
)

type Jettison struct {
	nextIfSuccessful Routine
	nextIfFailed     Routine
}

func (j Jettison) Run(state *State) RoutineResult {
	hasJettisoned := false
	state.Log("Cargo is full")
	hasUnvisited := len(database.GetUnvisitedWaypointsInSystem(string(state.Ship.Nav.SystemSymbol))) > 0
	for _, slot := range state.Ship.Cargo.Inventory {
		marketsSelling := database.GetMarketsSellingInSystem([]string{slot.Symbol}, string(state.Ship.Nav.SystemSymbol))

		if j.IsUseless(slot.Symbol) || (state.Contract != nil && state.Contract.Terms.GetDeliverable(slot.Symbol) != nil && !hasUnvisited && len(marketsSelling) == 0) {
			state.Log(fmt.Sprintf("Jettison %dx %s", slot.Units, slot.Symbol))
			err := state.Ship.JettisonCargo(state.Context, slot.Symbol, slot.Units)
			if err != nil {
				fmt.Println(err)
			}
			hasJettisoned = hasJettisoned || err == nil
		}
	}
	//return RoutineResult{WaitSeconds: 10}
	if hasJettisoned {
		return RoutineResult{
			SetRoutine: j.nextIfSuccessful,
		}
	}
	state.Log("Had nothing to jettison")
	return RoutineResult{
		SetRoutine: j.nextIfFailed,
	}
}

// TODO: deduplicate
func (j Jettison) IsUseless(item string) bool {
	for _, uselessItem := range uselessItems {
		if uselessItem == item {
			return true
		}
	}
	return false
}

func (j Jettison) Name() string {
	return "Jettison Waste"
}
