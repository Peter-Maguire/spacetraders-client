package routine

func DetermineObjective(state *State) RoutineResult {
	if state.Ship.HasMount("MOUNT_MINING_LASER_I") {
		if state.Ship.Cargo.Units >= state.Ship.Cargo.Capacity-5 {
			state.Log("We're full up here")
			return RoutineResult{
				SetRoutine: SellExcessInventory,
			}
		}
		return RoutineResult{
			SetRoutine: GoToAsteroidField,
		}
	}

	state.Log("This type of ship isn't supported yet")
	return RoutineResult{
		WaitSeconds: 3600,
	}
}
