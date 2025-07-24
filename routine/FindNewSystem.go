package routine

import (
	"encoding/json"
	"fmt"
	"sort"
	"spacetraders/constant"
	"spacetraders/database"
	"spacetraders/entity"
	"spacetraders/http"
	"spacetraders/util"
)

type FindNewSystem struct {
	isAtJumpGate  bool
	systems       *[]entity.System
	startFromPage int
	next          Routine
	skipVisited   bool
}

func (f FindNewSystem) Run(state *State) RoutineResult {
	currentSystem := database.GetSystemData(state.Ship.Nav.SystemSymbol)

	if currentSystem == nil {
		currentSystem, _ = state.Ship.Nav.WaypointSymbol.GetSystem(state.Context)
		database.AddUnvisitedSystems([]entity.System{*currentSystem}, 0)
	}

	return RoutineResult{Stop: true, StopReason: "System jumping not supported"}

	if f.startFromPage == 0 {
		unvisitedSystems := database.GetUnvisitedSystems()

		if !f.skipVisited {
			visitedSystems := database.GetVisitedSystems()
			unvisitedSystems = append(visitedSystems, unvisitedSystems...)
		}

		if len(unvisitedSystems) == 0 {
			f.startFromPage = 1
			return RoutineResult{SetRoutine: f}
		}

		sort.Slice(unvisitedSystems, func(i, j int) bool {
			sys1 := unvisitedSystems[i]
			sys2 := unvisitedSystems[j]
			return util.CalcDistance(currentSystem.X, currentSystem.Y, sys1.X, sys1.Y) < util.CalcDistance(currentSystem.X, currentSystem.Y, sys2.X, sys2.Y)
		})

		for _, candidate := range unvisitedSystems {
			// Because we are sorted by distance, we can stop at 2000 since no other systems will be reachable
			//if util.CalcDistance(currentSystem.X, currentSystem.Y, candidate.X, candidate.Y) > 2000 {
			//	state.Log(fmt.Sprintf("System %s is over 2000 units away", candidate.System))
			//	continue
			//}

			var systemEntity entity.System
			err := json.Unmarshal(candidate.Data, &systemEntity)
			if err != nil {
				state.Log(fmt.Sprintf("Failed to unmarshal existing system: %s", err))
				continue
			}

			if f.CanJumpTo(&systemEntity, currentSystem) {
				state.Log(fmt.Sprintf("Found good known but unexplored system %s", systemEntity.Symbol))
				jumpGate := systemEntity.GetJumpGate(state.Context)
				if jumpGate != nil {
					jumpResult, err := state.Ship.Jump(state.Context, systemEntity.Waypoints[0].Symbol)
					if err != nil {
						state.Log("Error jumping")
						fmt.Println(err)
					} else {
						cooldownTime := jumpResult.Cooldown.Expiration
						return RoutineResult{
							WaitUntil:  &cooldownTime,
							SetRoutine: f.next,
						}
					}
				} else {
					state.Log("System doesn't have a jump gate")
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

	systemsPtr, err := state.Agent.Systems(state.Context, f.startFromPage)

	if err != nil {
		state.Log(err.Error())
		return RoutineResult{
			Stop:       true,
			StopReason: err.Error(),
		}
	}

	systems := *systemsPtr

	if len(systems) == 0 {
		state.Log("Out of systems")
		return RoutineResult{
			SetRoutine: f.next,
		}
	}
	database.AddUnvisitedSystems(systems, f.startFromPage)

	sort.Slice(systems, func(i, j int) bool {
		return currentSystem.GetDistanceFrom(&systems[i]) < currentSystem.GetDistanceFrom(&systems[j])
	})

	for _, system := range systems {
		if f.CanJumpTo(&system, currentSystem) {
			if f.isAtJumpGate || state.Ship.Cargo.GetSlotWithItem("ANTIMATTER").Units > 0 {
				state.Log(fmt.Sprintf("Jumping to %s", system.Symbol))
				jumpGate := system.GetJumpGate(state.Context)
				if jumpGate == nil {
					state.Log("No jump gate in this system")
					continue
				}
				jumpResult, err := state.Ship.Jump(state.Context, jumpGate.Symbol)
				if err != nil {
					if err.Code == http.ErrJumpGateUnderConstruction {
						state.Log("Gate is under construction")
						continue
					}
					state.Log("Error jumping")
					fmt.Println(err)
				} else {
					cooldownTime := jumpResult.Cooldown.Expiration
					return RoutineResult{
						WaitUntil:  &cooldownTime,
						SetRoutine: f.next,
					}
				}
			}
			state.Log("We don't have the antimatter and we're not at the jump gate")
			return RoutineResult{
				Stop:       true,
				StopReason: "Not at jump gate or no antimatter",
			}
		} else if util.GetFuelCost(system.GetDistanceFrom(currentSystem), "DRIFT") < state.Ship.Fuel.Current && state.Ship.CanWarp() {
			_ = state.Ship.SetFlightMode(state.Context, "DRIFT")
			jumpGate := system.GetJumpGate(state.Context)
			if jumpGate == nil {
				state.Log("No jump gate in this system")
				continue
			}
			res, err := state.Ship.Warp(state.Context, jumpGate.Symbol)
			if err != nil {
				state.Log(err.Error())
				continue
			}
			arrival := res.Nav.Route.Arrival
			return RoutineResult{
				WaitUntil:  &arrival,
				SetRoutine: f.next,
			}
		}
	}
	state.Log("No eligible systems on this page")
	f.startFromPage++
	return RoutineResult{
		SetRoutine: f,
	}
}

func (f FindNewSystem) CanJumpTo(toSystem *entity.System, fromSystem *entity.System) bool {
	// Can't jump to the system we are currently in
	if toSystem.Symbol == fromSystem.Symbol {
		fmt.Println("Can't jump to current system")
		return false
	}
	// There is no jump gate at the destination
	if !f.HasJumpGate(toSystem.Waypoints) {
		fmt.Println("System has no jump gate")
		return false
	}

	// Can't jump further than 2000 units
	if toSystem.GetDistanceFrom(fromSystem) >= 2000 {
		fmt.Println("System is too far away")
		return false
	}

	return true
}

func (f FindNewSystem) HasJumpGate(waypoints []entity.LimitedWaypointData) bool {
	for _, waypoint := range waypoints {
		if waypoint.Type == constant.WaypointTypeJumpGate {
			return true
		}
	}
	return false
}

func (f FindNewSystem) Name() string {
	return fmt.Sprintf("Find New System (Page %d) -> %s", f.startFromPage, f.next.Name())
}
