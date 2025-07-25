package routine

import (
	"fmt"
	"spacetraders/constant"
	"spacetraders/database"
	"spacetraders/util"
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
		return RoutineResult{
			SetRoutine: Explore{
				oneShot: true,
				next:    d,
			},
		}
	}

	if state.Ship.Nav.FlightMode != "CRUISE" {
		state.Log("Changing flight mode to CRUISE")
		_ = state.Ship.SetFlightMode(state.Context, "CRUISE")
	}

	// TODO: satellite should explore until it's explored the entire system next go to Refresh Markets (rotate through all the markets refreshing each)
	if /*state.Ship.Registration.Role == "COMMAND" ||*/ state.Ship.Registration.Role == constant.ShipRoleSatellite {
		return RoutineResult{
			SetRoutine: Satellite{},
		}
	}

	if state.Ship.Registration.Role == constant.ShipRoleCommand && len(state.Haulers) < 1 {
		if state.Contract == nil || state.Contract.Fulfilled {
			return RoutineResult{SetRoutine: GoToRandomFactionWaypoint{next: NegotiateContract{}}}
		}
		for _, deliverable := range state.Contract.Terms.Deliver {
			if !deliverable.IsFulfilled() && !util.IsMineable(deliverable.TradeSymbol) {
				state.Log(fmt.Sprintf("We have to find some %s to deliver", deliverable.TradeSymbol))
				return RoutineResult{SetRoutine: ProcureContractItem{deliverable: &deliverable}}
			}
		}
	}

	if state.Ship.Registration.Role == constant.ShipRoleHauler {
		haulerNumber := 0
		for i, hauler := range state.Haulers {
			if hauler.Symbol == state.Ship.Symbol {
				haulerNumber = i
				break
			}
		}

		state.Log(fmt.Sprintf("Hauler number: %d", haulerNumber))
		if haulerNumber == 0 && len(state.Haulers) > 1 {
			if state.Contract != nil && state.Contract.Fulfilled == false {
				for _, deliverable := range state.Contract.Terms.Deliver {
					if !deliverable.IsFulfilled() && !util.IsMineable(deliverable.TradeSymbol) {
						state.Log(fmt.Sprintf("We have to find some %s to deliver", deliverable.TradeSymbol))
						return RoutineResult{SetRoutine: ProcureContractItem{deliverable: &deliverable}}
					}
				}
			} else {
				return RoutineResult{
					SetRoutine: NegotiateContract{},
				}
			}
		}

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
		rout := GoToMiningArea{}
		rout.next = GetSurvey{next: Explore{next: GoToRandomFactionWaypoint{next: rout}}}
		return RoutineResult{
			SetRoutine: rout,
		}
	}

	if state.Ship.IsMiningShip() {
		if state.Ship.Cargo.Units >= state.Ship.Cargo.Capacity {
			state.Log("We're full up here")
			return RoutineResult{
				SetRoutine: Jettison{
					nextIfSuccessful: d,
					nextIfFailed:     FullWait{},
				},
				//SetRoutine: SellExcessInventory{MineOres{}},
			}
		}

		return RoutineResult{
			SetRoutine: GoToMiningArea{},
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
