package routine

import (
	"fmt"
	"sort"
	"spacetraders/constant"
	"spacetraders/database"
	"spacetraders/entity"
)

type FindNewWaypoint struct {
	desiredTrait constant.WaypointTrait
	visitVisited bool
	next         Routine
}

func (f FindNewWaypoint) Run(state *State) RoutineResult {
	if f.next == nil {
		f.next = Explore{
			desiredTrait: f.desiredTrait,
			visitVisited: f.visitVisited,
		}
	}
	// find new place
	waypoints, _ := state.Ship.Nav.WaypointSymbol.GetSystemWaypoints(state.Context)
	database.LogWaypoints(waypoints)
	goodWaypoints := make([]entity.WaypointData, 0)
	for _, waypoint := range *waypoints {
		if f.hasGoodTraits(waypoint.Traits) {
			visited := database.GetWaypoint(waypoint.Symbol)
			if waypoint.Symbol != state.Ship.Nav.WaypointSymbol && f.visitVisited || visited == nil || visited.FirstVisited.Unix() < 0 {
				//state.Log(fmt.Sprintf("Found interesting waypoint at %s", waypoint.Symbol))
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
			d2 := goodWaypoints[j].GetDistanceFrom(currentWaypointData.LimitedWaypointData)
			return d1 < d2
		})
		distance := goodWaypoints[0].GetDistanceFrom(currentWaypointData.LimitedWaypointData)
		state.Log(fmt.Sprintf("Nearest waypoint is %s which is %d away", goodWaypoints[0].Symbol, distance))
		return RoutineResult{
			SetRoutine: NavigateTo{
				waypoint:     goodWaypoints[0].Symbol,
				next:         f.next,
				nextIfNoFuel: f.next,
			},
		}
	}

	system, _ := state.Ship.Nav.WaypointSymbol.GetSystem(state.Context)
	database.VisitSystem(system, waypoints)
	//// TODO: Fix system jumping
	//state.Log("No more good waypoints left in this system")
	//return RoutineResult{
	//	Stop:       true,
	//	StopReason: "No more good waypoints left in this system",
	//}

	for _, waypoint := range *waypoints {
		if waypoint.Type == constant.WaypointTypeJumpGate {
			if waypoint.Symbol == state.Ship.Nav.WaypointSymbol {
				state.Log("We're at a jump gate, time to go find a new place")
				return RoutineResult{
					SetRoutine: FindNewSystem{isAtJumpGate: true, next: f},
					//WaitUntil:  &cooldownUntil,
				}
			}
			state.Log("Going to jump gate")
		}
	}

	return RoutineResult{
		SetRoutine: GoToJumpGate{next: f},
	}

	//if state.Ship.Cargo.GetSlotWithItem("ANTIMATTER").Units > 2 {
	//	return RoutineResult{
	//		SetRoutine: FindNewSystem{next: f},
	//		//WaitUntil:  &cooldownUntil,
	//	}
	//}
	//state.Log("No jump gate either, not sure how we got here.")
	//return RoutineResult{
	//	Stop:       true,
	//	StopReason: "Unable to leave system",
	//}
}

var desiredTraits = []constant.WaypointTrait{"MARKETPLACE", "SHIPYARD", "UNCHARTED", "TRADING_HUB", "BLACK_MARKET", "COMMON_METAL_DEPOSITS", "RARE_METAL_DEPOSITS", "PRECIOUS_METAL_DEPOSITS", "MINERAL_DEPOSITS"}

func (f FindNewWaypoint) hasGoodTraits(traits []entity.Trait) bool {
	for _, trait := range traits {
		if f.desiredTrait != "" {
			if trait.Symbol == f.desiredTrait {
				return true
			}
		} else {
			for _, desiredTrait := range desiredTraits {
				if trait.Symbol == desiredTrait {
					return true
				}
			}
		}
	}
	return false
}

func (f FindNewWaypoint) Name() string {
	return fmt.Sprintf("Find New Waypoint -> %s", f.next.Name())
}
