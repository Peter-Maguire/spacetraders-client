package routine

import "spacetraders/entity"

type GoToMiningArea struct {
    next Routine
}

func (g GoToMiningArea) Run(state *State) RoutineResult {
    waypointData, _ := state.Ship.Nav.WaypointSymbol.GetWaypointData()

    // We are currently in an asteroid field
    if waypointData.Type == "ASTEROID_FIELD" {
        state.Log("We are already in an asteroid field, convenient!")
        return RoutineResult{
            SetRoutine: g.next,
        }
    }

    _ = state.Ship.EnsureNavState(entity.NavOrbit)

    system, _ := state.Ship.Nav.WaypointSymbol.GetSystem()

    for _, waypoint := range system.Waypoints {
        if waypoint.Type == "ASTEROID_FIELD" {
            return RoutineResult{
                SetRoutine: NavigateTo{waypoint.Symbol, g.next},
            }
        }
    }

    if state.Ship.Nav.SystemSymbol != state.Agent.Headquarters.GetSystemName() {
        return RoutineResult{
            SetRoutine: GoToJumpGate{next: GoToSystem{
                system: state.Agent.Headquarters.GetSystemName(),
                next:   g,
            }},
        }
    }

    state.Log("Couldn't find a waypoint pointing to an asteroid field")
    return RoutineResult{
        WaitSeconds: 60,
    }
}

func (g GoToMiningArea) Name() string {
    return "Go To Asteroid Field"
}
