package routine

import "fmt"

type Haul struct {
}

func (h Haul) Run(state *State) RoutineResult {
    state.WaitingForHttp = true
    cargo, _ := state.Ship.GetCargo()
    state.WaitingForHttp = false

    if cargo.Units == 0 {
        state.Log("Nothing to haul yet")
        return RoutineResult{
            WaitSeconds: 30,
        }
    }

    for _, slot := range cargo.Inventory {
        sellResult, err := state.Ship.SellCargo(slot.Symbol, slot.Units)
        if err == nil {
            state.Log(fmt.Sprintf("Sold %dx %s for %d credits", slot.Units, slot.Symbol, sellResult.Transaction.TotalPrice))
        }
    }

    return RoutineResult{
        WaitSeconds: 20,
    }
}

func (h Haul) Name() string {
    return "Haul"
}
