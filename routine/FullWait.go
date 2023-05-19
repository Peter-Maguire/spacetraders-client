package routine

type FullWait struct {
}

func (f FullWait) Run(state *State) RoutineResult {

    if len(state.Haulers) == 0 {
        return RoutineResult{
            SetRoutine: SellExcessInventory{next: MineOres{}},
        }
    }

    if state.Ship.Cargo.Units == state.Ship.Cargo.Capacity {
        //state.Log("Still full..")
        return RoutineResult{
            WaitSeconds: 30,
        }
    }

    return RoutineResult{
        SetRoutine: MineOres{},
    }
}

func (f FullWait) Name() string {
    return "Full Inv Wait"
}
