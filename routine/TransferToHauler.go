package routine

import (
    "fmt"
    "sort"
    "spacetraders/entity"
)

type TransferToHauler struct {
    next   Routine
    hauler *entity.Ship
}

func (t TransferToHauler) Run(state *State) RoutineResult {
    if t.hauler.Cargo.Units == t.hauler.Cargo.Capacity {
        state.Log("We have no hauler or hauler is full")
        return RoutineResult{SetRoutine: SellExcessInventory{}}
    }

    sort.Slice(state.Ship.Cargo.Inventory, func(i, j int) bool {
        return state.Ship.Cargo.Inventory[i].Units < state.Ship.Cargo.Inventory[j].Units
    })

    for _, slot := range state.Ship.Cargo.Inventory {
        if slot.Symbol == "ANTIMATTER" || slot.Units < 10 {
            continue
        }
        state.Log(fmt.Sprintf("Transferring %dx %s to hauler", slot.Units, slot.Symbol))
        err := state.Ship.TransferCargo(t.hauler.Symbol, slot.Symbol, slot.Units)
        if err != nil {
            state.Log("Transfer failed: " + err.Error())
            break
        }
    }

    state.FireEvent("transferToHauler", t.hauler)

    return RoutineResult{
        SetRoutine: t.next,
    }
}

func (t TransferToHauler) Name() string {
    return "Transfer To Hauler"
}
