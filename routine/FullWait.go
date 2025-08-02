package routine

import "spacetraders/constant"

type FullWait struct {
}

func (f FullWait) Run(state *State) RoutineResult {

	// TODO: wait for hauler to come back if going to the market and back would take longer than the hauler is going to take to come back
	haulers := state.GetShipsWithRoleAtOrGoingToWaypoint(constant.ShipRoleHauler, state.Ship.Nav.WaypointSymbol)

	hasHauler := false

	for _, h := range haulers {
		if h.Fuel.Current > 0 {
			hasHauler = true
			break
		}
	}

	if !hasHauler {
		state.Log("We have no hauler here")
		return RoutineResult{
			SetRoutine: SellExcessInventory{next: GoToMiningArea{}},
		}
	}

	if state.Ship.Cargo.Units == state.Ship.Cargo.Capacity {
		state.Log("Still full..")
		return RoutineResult{
			WaitSeconds: 30,
		}
	}

	return RoutineResult{
		SetRoutine: DetermineObjective{},
	}
}

func (f FullWait) Name() string {
	return "Wait for Hauler"
}
