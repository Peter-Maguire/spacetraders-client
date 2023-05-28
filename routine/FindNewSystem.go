package routine

import (
	"encoding/json"
	"fmt"
	"sort"
	"spacetraders/database"
	"spacetraders/entity"
	"spacetraders/util"
)

type FindNewSystem struct {
	isAtJumpGate  bool
	systems       *[]entity.System
	startFromPage int
}

func (f FindNewSystem) Run(state *State) RoutineResult {

	currentSystem := database.GetSystemData(state.Ship.Nav.SystemSymbol)

	if f.startFromPage == 0 {
		unvisitedSystems := database.GetUnvisitedSystems()

		sort.Slice(unvisitedSystems, func(i, j int) bool {
			sys1 := unvisitedSystems[i]
			sys2 := unvisitedSystems[j]
			return util.CalcDistance(currentSystem.X, currentSystem.Y, sys1.X, sys1.Y) < util.CalcDistance(currentSystem.X, currentSystem.Y, sys2.X, sys2.Y)
		})

		for _, candidate := range unvisitedSystems {
			// Because we are sorted by distance, we can stop at 2000 since no other systems will be reachable
			if util.CalcDistance(currentSystem.X, currentSystem.Y, candidate.X, candidate.Y) > 2000 {
				break
			}

			var systemEntity entity.System
			err := json.Unmarshal(candidate.Data, &systemEntity)
			if err != nil {
				state.Log(fmt.Sprintf("Failed to unmarshal existing system: %s", err))
				continue
			}

			if f.CanJumpTo(&systemEntity, currentSystem) {
				state.Log(fmt.Sprintf("Found good known but unexplored system %s", systemEntity.Symbol))
				state.WaitingForHttp = true
				jumpResult, err := state.Ship.Jump(systemEntity.Symbol)
				state.WaitingForHttp = false
				if err != nil {
					state.Log("Error jumping")
					fmt.Println(err)
				} else {
					cooldownTime := jumpResult.Cooldown.Expiration
					return RoutineResult{
						WaitUntil:  &cooldownTime,
						SetRoutine: Explore{},
					}
				}
			}
		}

		startPage := 0
		for _, candidate := range unvisitedSystems {
			if candidate.Page > startPage {
				startPage = candidate.Page
			}
		}

		f.startFromPage = startPage
	}

	state.Log(fmt.Sprintf("Starting on page %d", f.startFromPage))

	state.WaitingForHttp = true
	systemsPtr, _ := state.Agent.Systems(f.startFromPage)
	state.WaitingForHttp = false

	systems := *systemsPtr

	if len(systems) == 0 {
		return RoutineResult{Stop: true, StopReason: "Ran out of systems"}
	}
	database.AddUnvisitedSystems(systems, f.startFromPage)

	sort.Slice(systems, func(i, j int) bool {
		return currentSystem.GetDistanceFrom(&systems[i]) < currentSystem.GetDistanceFrom(&systems[j])
	})

	for _, system := range systems {
		if f.CanJumpTo(&system, currentSystem) {
			if f.isAtJumpGate || state.Ship.Cargo.GetSlotWithItem("ANTIMATTER").Units > 0 {
				state.Log(fmt.Sprintf("Jumping to %s", system.Symbol))
				state.WaitingForHttp = true
				jumpResult, err := state.Ship.Jump(system.Symbol)
				state.WaitingForHttp = false
				if err != nil {
					state.Log("Error jumping")
					fmt.Println(err)
				} else {
					cooldownTime := jumpResult.Cooldown.Expiration
					return RoutineResult{
						WaitUntil:  &cooldownTime,
						SetRoutine: Explore{},
					}
				}
			}
			state.Log("We don't have the antimatter and we're not at the jump gate")
			return RoutineResult{
				Stop:       true,
				StopReason: "Not at jump gate or no antimatter",
			}
		}
	}
	state.Log("No new systems to check on this page")
	f.startFromPage++
	return RoutineResult{
		SetRoutine: f,
	}
}

func (f FindNewSystem) CanJumpTo(toSystem *entity.System, fromSystem *entity.System) bool {
	// Can't jump to the system we are currently in
	if toSystem.Symbol == fromSystem.Symbol {
		return false
	}
	// There is no jump gate at the destination
	if !f.HasJumpGate(toSystem.Waypoints) {
		return false
	}

	// Can't jump further than 2000 units
	if toSystem.GetDistanceFrom(fromSystem) >= 2000 {
		return false
	}

	return true
}

func (f FindNewSystem) HasJumpGate(waypoints []entity.LimitedWaypointData) bool {
	for _, waypoint := range waypoints {
		if waypoint.Type == "JUMP_GATE" {
			return true
		}
	}
	return false
}

func (f FindNewSystem) Name() string {
	return fmt.Sprintf("Find New System - Page %d", f.startFromPage)
}
