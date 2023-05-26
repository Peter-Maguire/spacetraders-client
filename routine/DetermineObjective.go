package routine

import (
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

    if state.Ship.Registration.Role == "COMMAND" {
        state.Log("Command ship can go exploring")
        return RoutineResult{
            SetRoutine: Explore{},
        }
    }

    if state.Ship.Registration.Role == "HAULER" {
        // TODO: Fix hauling
        return RoutineResult{Stop: true, StopReason: "Hauler not supported"}
        return RoutineResult{
            SetRoutine: GoToMiningArea{Haul{}},
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
