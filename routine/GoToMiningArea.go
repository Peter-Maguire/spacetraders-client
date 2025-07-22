package routine

import (
	"fmt"
	"sort"
	"spacetraders/database"
	"spacetraders/entity"
)

type GoToMiningArea struct {
}

func (g GoToMiningArea) Run(state *State) RoutineResult {
	_ = state.Ship.EnsureNavState(state.Context, entity.NavOrbit)

	waypointsPtr, _ := state.Ship.Nav.WaypointSymbol.GetSystemWaypoints(state.Context)
	waypoints := *waypointsPtr
	waypointScores := make(map[entity.Waypoint]int)

	//currentWaypoint, _ := state.Ship.Nav.WaypointSymbol.GetWaypointData(state.Context)

	waypointData := make([]*database.Waypoint, len(waypoints))
	for i, waypoint := range waypoints {
		waypointData[i] = database.GetWaypoint(waypoint.Symbol)
	}

	eligibleWaypoints := make([]entity.WaypointData, 0)
	for _, waypoint := range waypoints {
		eligible, score := g.ScoreWaypoint(waypoint, state, waypointData)
		if !eligible {
			continue
		}
		eligibleWaypoints = append(eligibleWaypoints, waypoint)
		waypointScores[waypoint.Symbol] = score
	}

	sort.Slice(eligibleWaypoints, func(i, j int) bool {
		return waypointScores[waypoints[i].Symbol] > waypointScores[waypoints[j].Symbol]
	})

	if len(waypointScores) == 0 {
		state.Log("No good waypoints found within reach")
		if state.Ship.Fuel.IsFull() {
			if state.Ship.Nav.FlightMode == "DRIFT" {
				return RoutineResult{
					Stop:       true,
					StopReason: "Unable to find anywhere to mine in range",
				}
			}
			state.Log("Trying again in drift mode")
			state.Ship.SetFlightMode(state.Context, "DRIFT")
			return RoutineResult{}
		}

		state.Log("Attempting to refuel and trying again")
		return RoutineResult{
			SetRoutine: Refuel{
				next: g,
			},
		}
	}

	bestWaypoint := eligibleWaypoints[0]

	state.Log(fmt.Sprintf("Waypoint %s has score of %d", bestWaypoint.Symbol, waypointScores[bestWaypoint.Symbol]))

	return RoutineResult{SetRoutine: NavigateTo{
		waypoint: bestWaypoint.Symbol,
		next:     MineOres{},
	}}

	//if state.Ship.Nav.SystemSymbol != state.Agent.Headquarters.GetSystemName() {
	//	return RoutineResult{
	//		SetRoutine: GoToJumpGate{next: GoToSystem{
	//			system: state.Agent.Headquarters.GetSystemName(),
	//			next:   g,
	//		}},
	//	}
	//}
	//
	//state.Log("Couldn't find a waypoint pointing to an asteroid field")
	//return RoutineResult{
	//	WaitSeconds: 60,
	//}
}

func (g GoToMiningArea) ScoreWaypoint(waypoint entity.WaypointData, state *State, waypoints []*database.Waypoint) (bool, int) {
	score := -1
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

	if score <= 0 {
		return false, score
	}

	if waypoint.HasTrait("MARKETPLACE") {
		score += 10
	}

	if waypoint.HasTrait("OVERCROWDED") {
		return false, 0
	}

	if waypoint.HasTrait("BARREN") {
		return false, 0
	}

	if waypoint.HasTrait("STRIPPED") {
		return false, 0
	}

	closestDistance := 2000
	for _, dbWaypoint := range waypoints {
		if dbWaypoint.MarketData == nil || string(dbWaypoint.MarketData) == "null" {
			continue
		}
		if dbWaypoint.Waypoint == string(waypoint.Symbol) {
			continue
		}
		wpData := dbWaypoint.GetData()
		distance := waypoint.GetDistanceFrom(wpData.LimitedWaypointData)
		if distance < closestDistance {
			closestDistance = distance
		}
	}

	score -= closestDistance

	return true, score
}

func (g GoToMiningArea) Name() string {
	return "Go To Mining Area"
}
