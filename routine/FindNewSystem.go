package routine

import (
	"fmt"
	"spacetraders/database"
	"spacetraders/entity"
)

type FindNewSystem struct {
	isAtJumpGate  bool
	systems       *[]entity.System
	startFromPage int
}

func (f FindNewSystem) Run(state *State) RoutineResult {
	//if f.systems == nil {
	//if f.startFromPage == 0 {
	//    state.Log("Scanning systems...")
	//    scanResult, err := state.Ship.ScanSystems()
	//    if err == nil {
	//        f.systems = scanResult.Systems
	//        waitUntil := scanResult.Cooldown.Expiration
	//        return RoutineResult{
	//            WaitUntil:  &waitUntil,
	//            SetRoutine: f,
	//        }
	//    }
	//    state.Log(err.Error())
	//}
	if f.startFromPage == 0 {
		f.startFromPage = 1
	}
	state.Log(fmt.Sprintf("Starting on page %d", f.startFromPage))

	systems, _ := state.Agent.Systems(f.startFromPage)
	//f.systems = systems
	//if err != nil {
	//    state.Log(err.Error())
	//    return RoutineResult{Stop: true}
	//}
	//}

	for _, system := range *systems {
		if system.Symbol != state.Ship.Nav.SystemSymbol && f.HasJumpGate(system.Waypoints) && database.GetSystem(system.Symbol) == nil {
			if f.isAtJumpGate {
				state.Log(fmt.Sprintf("Jumping to %s", system.Symbol))
				jumpResult, err := state.Ship.Jump(system.Symbol)
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
			if state.Ship.Cargo.GetSlotWithItem("ANTIMATTER").Units > 2 {
				state.Log(fmt.Sprintf("Warping to %s", system.Symbol))
				warpResult, err := state.Ship.Warp(system.Waypoints[0].Symbol)
				if err != nil {
					state.Log("Error warping" + err.Error())
					fmt.Println(err)
				} else {
					arrivalTime := warpResult.Nav.Route.Arrival
					return RoutineResult{
						WaitUntil:  &arrivalTime,
						SetRoutine: Explore{},
					}
				}
			}
			state.Log("We don't have the antimatter and we're not at the jump gate")
			return RoutineResult{
				Stop: true,
			}
		}
	}
	state.Log("No new systems to check on this page")
	f.startFromPage++
	return RoutineResult{
		SetRoutine: f,
	}
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
