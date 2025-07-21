package routine

import (
	"fmt"
	"sort"
	"spacetraders/entity"
	"spacetraders/util"
)

type GoToMiningArea struct {
	next Routine
}

func (g GoToMiningArea) Run(state *State) RoutineResult {
	_ = state.Ship.EnsureNavState(state.Context, entity.NavOrbit)

	waypointsPtr, _ := state.Ship.Nav.WaypointSymbol.GetSystemWaypoints(state.Context)
	waypoints := *waypointsPtr
	waypointScores := make(map[entity.Waypoint]int)

	currentWaypoint, _ := state.Ship.Nav.WaypointSymbol.GetWaypointData(state.Context)

	for _, waypoint := range waypoints {
		distance := waypoint.GetDistanceFrom(currentWaypoint.LimitedWaypointData)
		fuelCost := util.GetFuelCost(distance, state.Ship.Nav.FlightMode)
		if fuelCost > state.Ship.Fuel.Current {
			// We can't do this because this one is too far away
			continue
		}
		waypointScores[waypoint.Symbol] = g.ScoreWaypoint(waypoint)
	}

	sort.Slice(waypoints, func(i, j int) bool {
		return waypointScores[waypoints[i].Symbol] > waypointScores[waypoints[j].Symbol]
	})

	if len(waypointScores) == 0 || waypointScores[waypoints[0].Symbol] <= 0 {
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

	if waypoint.HasTrait("MARKETPLACE") {
		score += 1
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
