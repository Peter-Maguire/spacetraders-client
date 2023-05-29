package routine

import (
	"fmt"
	"sort"
	"spacetraders/entity"
)

type GoToMiningArea struct {
	next Routine
}

func (g GoToMiningArea) Run(state *State) RoutineResult {
	_ = state.Ship.EnsureNavState(entity.NavOrbit)

	waypointsPtr, _ := state.Ship.Nav.WaypointSymbol.GetSystemWaypoints()
	waypoints := *waypointsPtr
	waypointScores := make(map[entity.Waypoint]int)

	for _, waypoint := range waypoints {
		waypointScores[waypoint.Symbol] = g.ScoreWaypoint(waypoint)
	}

	sort.Slice(waypoints, func(i, j int) bool {
		return waypointScores[waypoints[i].Symbol] > waypointScores[waypoints[j].Symbol]
	})

	bestWaypoint := waypoints[0]

	state.Log(fmt.Sprintf("Waypoint %s has score of %d", bestWaypoint.Symbol, waypointScores[bestWaypoint.Symbol]))

	if waypointScores[bestWaypoint.Symbol] > 0 {
		return RoutineResult{SetRoutine: NavigateTo{
			waypoint: bestWaypoint.Symbol,
			next:     MineOres{},
		}}
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

func (g GoToMiningArea) ScoreWaypoint(waypoint entity.WaypointData) int {
	score := 0
	if waypoint.HasTrait("PRECIOUS_METAL_DEPOSITS") {
		score += 25
	}

	if waypoint.HasTrait("RARE_METAL_DEPOSITS") {
		score += 20
	}

	if waypoint.HasTrait("COMMON_METAL_DEPOSITS") {
		score += 20
	}

	if waypoint.HasTrait("MINERAL_DEPOSITS") {
		score += 5
	}

	if waypoint.HasTrait("MARKETPLACE") {
		score += 2
	}

	if waypoint.HasTrait("OVERCROWDED") {
		score -= 5
	}

	if waypoint.HasTrait("BARREN") {
		score -= 10
	}

	if waypoint.HasTrait("STRIPPED") {
		score -= 20
	}

	return score
}

func (g GoToMiningArea) Name() string {
	return "Go To Mining Area"
}
