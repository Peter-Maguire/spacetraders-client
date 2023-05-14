package routine

import "spacetraders/entity"

type GoToAsteroidField struct {
}

func (g GoToAsteroidField) Run(state *State) RoutineResult {
	_ = state.Ship.EnsureNavState(entity.NavOrbit)
	waypointData, _ := state.Ship.Nav.WaypointSymbol.GetWaypointData()

	// We are currently in an asteroid field
	if waypointData.Type == "ASTEROID_FIELD" {
		state.Log("We are already in an asteroid field, convenient!")
		return RoutineResult{
			SetRoutine: GetSurvey{},
		}
	}

	system, _ := state.Ship.Nav.WaypointSymbol.GetSystem()

	for _, waypoint := range system.Waypoints {
		if waypoint.Type == "ASTEROID_FIELD" {
			return RoutineResult{
				SetRoutine: NavigateTo{waypoint.Symbol, GetSurvey{}},
			}
		}
	}

	state.Log("Couldn't find a waypoint pointing to an asteroid field")
	return RoutineResult{
		WaitSeconds: 60,
	}
}

func (g GoToAsteroidField) Name() string {
	return "Go To Asteroid Field"
}
