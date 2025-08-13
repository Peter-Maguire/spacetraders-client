package routine

import (
	"fmt"
	"math/rand"
	"spacetraders/constant"
	"spacetraders/database"
	"time"
)

type DetermineObjective struct {
}

func (d DetermineObjective) Run(state *State) RoutineResult {
	if state.Ship.Nav.Status == "IN_TRANSIT" && state.Ship.Nav.Route.Arrival.After(time.Now()) {
		state.Log("We are currently going somewhere")
		arrivalTime := state.Ship.Nav.Route.Arrival
		return RoutineResult{
			WaitUntil: &arrivalTime,
		}
	}

	dbWaypoint := database.GetWaypoint(state.Ship.Nav.WaypointSymbol)
	if dbWaypoint == nil || dbWaypoint.FirstVisited.Unix() < 0 {
		state.Log("Visiting waypoint")
		return RoutineResult{
			SetRoutine: Explore{
				oneShot: true,
				next:    d,
			},
		}
	}

	if state.Ship.Fuel.Current == 1 {
		return RoutineResult{
			SetRoutine: Refuel{next: d},
		}
	}

	//phase := state.
	state.Ship.EnsureFlightMode(state.Context, constant.FlightModeCruise)

	// TODO: satellite should explore until it's explored the entire system next go to Refresh Markets (rotate through all the markets refreshing each)
	unvisitedWaypoints := database.GetUnvisitedWaypointsInSystem(string(state.Ship.Nav.SystemSymbol))

	goodUnvisitedWaypoints := false
	for _, uw := range unvisitedWaypoints {
		data := uw.GetData()
		if data.HasTrait(constant.TraitMarketplace) || data.HasTrait(constant.TraitShipyard) {
			goodUnvisitedWaypoints = true
			break
		}
	}

	if (state.Ship.Registration.Role == constant.ShipRoleCommand && goodUnvisitedWaypoints) || state.Ship.Registration.Role == constant.ShipRoleSatellite {
		return RoutineResult{
			SetRoutine: Satellite{},
		}
	}

	if state.Ship.Registration.Role == constant.ShipRoleCommand {
		if state.Config.GetBool("commandShipTrades", false) {
			return RoutineResult{
				WaitSeconds: rand.Intn(10),
				SetRoutine:  Trade{},
			}
		}

		if state.Contract == nil || state.Contract.Fulfilled {
			return RoutineResult{SetRoutine: GoToRandomFactionWaypoint{next: NegotiateContract{}}}
		}

		for _, deliverable := range state.Contract.Terms.Deliver {
			if !deliverable.IsFulfilled() {
				state.Log(fmt.Sprintf("We have to find some %s to deliver", deliverable.TradeSymbol))
				return RoutineResult{SetRoutine: ProcureContractItem{deliverable: &deliverable}}
			}
		}
	}

	if state.Ship.Registration.Role == constant.ShipRoleTransport {

		transporters := state.GetShipsWithRole(constant.ShipRoleTransport)

		if len(transporters) > 10 && transporters[0].Symbol == state.Ship.Symbol {
			waypoints, _ := state.Ship.Nav.WaypointSymbol.GetSystemName().GetWaypointsOfType(state.Context, constant.WaypointTypeJumpGate)
			for _, waypoint := range *waypoints {
				if waypoint.SystemSymbol == state.Ship.Nav.SystemSymbol && waypoint.Type == constant.WaypointTypeJumpGate {
					fullWp, _ := waypoint.GetFullWaypoint(state.Context)
					if fullWp.IsUnderConstruction {
						return RoutineResult{
							SetRoutine: BuildJumpGate{next: d},
						}
					}
				}
			}
		}

		return RoutineResult{
			WaitSeconds: rand.Intn(90),
			SetRoutine:  Trade{},
		}
	}

	if state.Ship.Registration.Role == constant.ShipRoleHauler {
		//state.Log(fmt.Sprintf("Hauler number: %d", haulerNumber))
		//if haulerNumber == 0 && len(state.Haulers) > 1 {
		//	if state.Contract != nil && state.Contract.Fulfilled == false {
		//		for _, deliverable := range state.Contract.Terms.Deliver {
		//			if !deliverable.IsFulfilled() && !util.IsMineable(deliverable.TradeSymbol) {
		//				state.Log(fmt.Sprintf("We have to find some %s to deliver", deliverable.TradeSymbol))
		//				return RoutineResult{SetRoutine: ProcureContractItem{deliverable: &deliverable}}
		//			}
		//		}
		//	} else {
		//		return RoutineResult{
		//			SetRoutine: NegotiateContract{},
		//		}
		//	}
		//}

		// TODO: Fix hauling
		return RoutineResult{
			SetRoutine: Haul{},
			//Stop:       true,
			//StopReason: "Hauling not supported",
			//SetRoutine:  Explore{},
			//WaitSeconds: int(time.Now().UnixMilli()%100) * 10,
		}
	}

	if state.Ship.Registration.Role == constant.ShipRoleRefinery {
		return RoutineResult{
			SetRoutine: Refine{},
		}
	}

	if state.Ship.Registration.Role == constant.ShipRoleSurveyor {
		return RoutineResult{
			SetRoutine: GoToMiningArea{next: GetSurvey{}},
		}
	}

	if state.Ship.Registration.Role == constant.ShipRoleExcavator {
		if state.Ship.Cargo.Units >= state.Ship.Cargo.Capacity {
			state.Log("We're full up here")
			return RoutineResult{
				SetRoutine: Jettison{
					nextIfSuccessful: d,
					nextIfFailed:     FullWait{},
				},
			}
		}
	}

	if state.Ship.IsMiningShip() {
		return RoutineResult{
			SetRoutine: GoToMiningArea{},
		}
	}

	if state.Ship.IsSiphonShip() {
		return RoutineResult{
			SetRoutine: GoToGasGiant{},
		}
	}

	state.Log(fmt.Sprintf("This type of ship (%s) isn't supported yet", state.Ship.Registration.Role))
	return RoutineResult{
		Stop:       true,
		StopReason: fmt.Sprintf("Unknown Ship Type %s", state.Ship.Registration.Role),
	}
}

func (d DetermineObjective) Name() string {
	return "Determine Objective"
}
