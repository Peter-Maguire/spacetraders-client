package routine

import (
	"math/rand"
	"spacetraders/entity"
)

type GoToRandomFactionWaypoint struct {
	next Routine
}

func (g GoToRandomFactionWaypoint) Run(state *State) RoutineResult {
	waypoints, _ := state.Ship.Nav.WaypointSymbol.GetSystemWaypoints()

	factionWaypoints := make([]*entity.WaypointData, 0)

	for _, waypoint := range *waypoints {
		if waypoint.Faction.Symbol != "" {
			factionWaypoints = append(factionWaypoints, &waypoint)
		}
	}
	waypoint := int(rand.Int63n(int64(len(factionWaypoints))))

	return RoutineResult{
		SetRoutine: NavigateTo{
			waypoint: factionWaypoints[waypoint].Symbol,
			next:     g.next,
		},
	}
}

func (g GoToRandomFactionWaypoint) Name() string {
	return "Go To Random Faction Waypoint"
}
