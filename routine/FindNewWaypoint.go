package routine

import (
	"fmt"
	"spacetraders/database"
	"spacetraders/entity"
	"time"
)

type FindNewWaypoint struct {
}

func (f FindNewWaypoint) Run(state *State) RoutineResult {
	// find new place
	var waypoints *[]entity.WaypointData
	var cooldownUntil = time.Now()
	result, err := state.Ship.ScanWaypoints()
	if err != nil {
		state.Log(err.Error())
		waypoints, _ = state.Ship.Nav.WaypointSymbol.GetSystemWaypoints()
	} else {
		waypoints = result.Waypoints
		cooldownUntil = result.Cooldown.Expiration
	}
	for _, waypoint := range *waypoints {
		if f.hasGoodTraits(waypoint.Traits) {
			visited := database.GetWaypoint(waypoint.Symbol)
			if visited == nil {
				state.Log(fmt.Sprintf("Found interesting waypoint at %s", waypoint.Symbol))
				return RoutineResult{
					SetRoutine: NavigateTo{waypoint: waypoint.Symbol, next: Explore{}},
					WaitUntil:  &cooldownUntil,
				}
			}
		}
	}

	system, _ := state.Ship.Nav.WaypointSymbol.GetSystem()
	database.VisitSystem(system, waypoints)
	state.Log("No more good waypoints left in this system")
	for _, waypoint := range *waypoints {
		if waypoint.Type == "JUMP_GATE" {
			if waypoint.Symbol == state.Ship.Nav.WaypointSymbol {
				state.Log("We're at a jump gate, time to go find a new place")
				return RoutineResult{
					SetRoutine: FindNewSystem{isAtJumpGate: true},
					WaitUntil:  &cooldownUntil,
				}
			}
			state.Log("Going to jump gate")
			return RoutineResult{
				SetRoutine: NavigateTo{waypoint: waypoint.Symbol, next: Explore{}},
				WaitUntil:  &cooldownUntil,
			}
		}
	}
	if state.Ship.Cargo.GetSlotWithItem("ANTIMATTER").Units > 2 {
		return RoutineResult{
			SetRoutine: FindNewSystem{},
			WaitUntil:  &cooldownUntil,
		}
	}
	state.Log("No jump gate either, not sure how we got here. May as well go mining.")
	return RoutineResult{
		Stop: true,
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
