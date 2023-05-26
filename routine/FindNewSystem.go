package routine

import (
    "fmt"
    "sort"
    "spacetraders/database"
    "spacetraders/entity"
)

type FindNewSystem struct {
    isAtJumpGate  bool
    systems       *[]entity.System
    startFromPage int
}

func (f FindNewSystem) Run(state *State) RoutineResult {
    if f.startFromPage == 0 {
        f.startFromPage = 1
    }
    state.Log(fmt.Sprintf("Starting on page %d", f.startFromPage))

    state.WaitingForHttp = true
    systemsPtr, _ := state.Agent.Systems(f.startFromPage)
    currentSystem, _ := state.Ship.Nav.WaypointSymbol.GetSystem()
    state.WaitingForHttp = false

    systems := *systemsPtr

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
    // We've already explored this system
    if database.GetSystem(toSystem.Symbol) != nil {
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
