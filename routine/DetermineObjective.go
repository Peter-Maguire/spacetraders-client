package routine

type DetermineObjective struct {
}

func (d DetermineObjective) Run(state *State) RoutineResult {
	if state.Ship.IsMiningShip() {
		if state.Ship.Cargo.Units >= state.Ship.Cargo.Capacity-5 {
			state.Log("We're full up here")
			return RoutineResult{
				SetRoutine: SellExcessInventory{},
			}
		}
		return RoutineResult{
			SetRoutine: GoToAsteroidField{},
		}
	}

	state.Log("This type of ship isn't supported yet")
	return RoutineResult{
		Stop: true,
	}
}

func (d DetermineObjective) Name() string {
	return "Determine Objective"
}
