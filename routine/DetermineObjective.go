package routine

import (
	"fmt"
	"spacetraders/util"
	"time"
)

type DetermineObjective struct {
}

func (d DetermineObjective) Run(state *State) RoutineResult {
	if state.Ship.Nav.FlightMode != "CRUISE" {
		_ = state.Ship.SetFlightMode("CRUISE")
	}
	if state.Ship.Nav.Status == "IN_TRANSIT" && state.Ship.Nav.Route.Arrival.After(time.Now()) {
		state.Log("We are currently going somewhere")
		arrivalTime := state.Ship.Nav.Route.Arrival
		return RoutineResult{
			WaitUntil: &arrivalTime,
		}
	}

	if state.Ship.Registration.Role == "COMMAND" {
		return RoutineResult{
			SetRoutine: Explore{},
		}
	}

	if state.Ship.Registration.Role == "HAULER" {
		haulerNumber := 0
		for i, hauler := range state.Haulers {
			if hauler.Symbol == state.Ship.Symbol {
				haulerNumber = i
				break
			}
		}

		if haulerNumber == 0 {
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
			Stop:       true,
			StopReason: "Hauling not supported",
			//SetRoutine:  Explore{},
			//WaitSeconds: int(time.Now().UnixMilli()%100) * 10,
		}
	}

	if state.Ship.IsMiningShip() {
		if state.Ship.Cargo.Units >= state.Ship.Cargo.Capacity {
			state.Log("We're full up here")
			return RoutineResult{
				SetRoutine: FullWait{},
				//SetRoutine: SellExcessInventory{MineOres{}},
			}
		}

		if state.Contract != nil {
			return RoutineResult{
				SetRoutine: GoToMiningArea{GetSurvey{}},
			}
		} else {
			return RoutineResult{
				SetRoutine: GoToMiningArea{MineOres{}},
			}
		}
	}

	state.Log("This type of ship isn't supported yet")
	return RoutineResult{
		Stop:       true,
		StopReason: "Unknown Ship Type",
	}
}

func (d DetermineObjective) Name() string {
	return "Determine Objective"
}
