package routine

import (
	"fmt"
	"sort"
	"spacetraders/database"
	"spacetraders/entity"
)

type FindNewWaypoint struct {
}

func (f FindNewWaypoint) Run(state *State) RoutineResult {
	// find new place
	waypoints, _ := state.Ship.Nav.WaypointSymbol.GetSystemWaypoints()
	database.LogWaypoints(waypoints)
	goodWaypoints := make([]entity.WaypointData, 0)
	for _, waypoint := range *waypoints {
		if f.hasGoodTraits(waypoint.Traits) {
			visited := database.GetWaypoint(waypoint.Symbol)
			if visited == nil || visited.FirstVisited.Unix() < 0 {
				state.Log(fmt.Sprintf("Found interesting waypoint at %s", waypoint.Symbol))
				goodWaypoints = append(goodWaypoints, waypoint)
			}
		}
	}

	currentWaypoint := database.GetWaypoint(state.Ship.Nav.WaypointSymbol)
	currentWaypointData := currentWaypoint.GetData()
	if len(goodWaypoints) > 0 {
		state.Log(fmt.Sprintf("Found %d good waypoints", len(goodWaypoints)))
		sort.Slice(goodWaypoints, func(i, j int) bool {
			d1 := goodWaypoints[i].GetDistanceFrom(currentWaypointData.LimitedWaypointData)
			d2 := goodWaypoints[i].GetDistanceFrom(currentWaypointData.LimitedWaypointData)
			return d1 < d2
		})
		distance := goodWaypoints[0].GetDistanceFrom(currentWaypointData.LimitedWaypointData)
		state.Log(fmt.Sprintf("Nearest waypoint is %s which is %d away", goodWaypoints[0].Symbol, distance))
		return RoutineResult{
			SetRoutine: NavigateTo{
				waypoint:     goodWaypoints[0].Symbol,
				next:         Explore{},
				nextIfNoFuel: Explore{},
			},
		}
	}

	system, _ := state.Ship.Nav.WaypointSymbol.GetSystem()
	database.VisitSystem(system, waypoints)
	// TODO: Fix system jumping
	state.Log("No more good waypoints left in this system")
	return RoutineResult{
		Stop:       true,
		StopReason: "No more good waypoints left in this system",
	}
	for _, waypoint := range *waypoints {
		if waypoint.Type == "JUMP_GATE" {
			if waypoint.Symbol == state.Ship.Nav.WaypointSymbol {
				state.Log("We're at a jump gate, time to go find a new place")
				return RoutineResult{
					SetRoutine: FindNewSystem{isAtJumpGate: true},
					//WaitUntil:  &cooldownUntil,
				}
			}
			state.Log("Going to jump gate")
			return RoutineResult{
				SetRoutine: NavigateTo{waypoint: waypoint.Symbol, next: Explore{}},
				//WaitUntil:  &cooldownUntil,
			}
		}
	}
	if state.Ship.Cargo.GetSlotWithItem("ANTIMATTER").Units > 2 {
		return RoutineResult{
			SetRoutine: FindNewSystem{},
			//WaitUntil:  &cooldownUntil,
		}
	}
	state.Log("No jump gate either, not sure how we got here. May as well go mining.")
	return RoutineResult{
		Stop:       true,
		StopReason: "Unable to leave system",
	}
}

var desiredTraits = []string{"MARKETPLACE", "SHIPYARD", "UNCHARTED", "TRADING_HUB", "BLACK_MARKET", "COMMON_METAL_DEPOSITS", "RARE_METAL_DEPOSITS", "PRECIOUS_METAL_DEPOSITS", "MINERAL_DEPOSITS"}

func (f FindNewWaypoint) hasGoodTraits(traits []entity.Trait) bool {
	for _, trait := range traits {
		for _, desiredTrait := range desiredTraits {
			if trait.Symbol == desiredTrait {
				return true
			}
		}
	}
	return false
}

func (f FindNewWaypoint) Name() string {
	return "Find New Waypoint"
}
