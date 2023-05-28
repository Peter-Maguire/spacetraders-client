package routine

import (
	"fmt"
	"sort"
	"spacetraders/database"
	"spacetraders/entity"
	"spacetraders/util"
)

type GoToSystem struct {
	system string
	next   Routine
}

func (g GoToSystem) Run(state *State) RoutineResult {
	currentSystem, _ := state.Ship.Nav.WaypointSymbol.GetSystem()

	targetSystem := database.GetSystemData(g.system)

	if targetSystem != nil {
		state.Log("TODO this system isn't stored in the database!")
		targetSystem, _ = state.Agent.GetSystem(g.system)
	}

	distance := currentSystem.GetDistanceFrom(targetSystem)

	if state.Ship.Cargo.GetSlotWithItem("ANTIMATTER").Units == 0 || distance > 500 {
		wpd, _ := state.Ship.Nav.WaypointSymbol.GetWaypointData()
		if wpd.Type != "JUMP_GATE" {
			state.Log("Going to jump gate")
			return RoutineResult{
				SetRoutine: GoToJumpGate{next: g},
			}
		}
	}

	if distance > 2000 {
		// Find an intermediary system thats within 2000 units
		intermediaries := database.GetVisitedSystems()
		jumpableIntermediaries := make([]database.System, 0)
		for _, intermediary := range intermediaries {
			if util.CalcDistance(currentSystem.X, currentSystem.Y, intermediary.X, intermediary.Y) < 2000 {
				jumpableIntermediaries = append(jumpableIntermediaries, intermediary)
			}
		}

		if len(jumpableIntermediaries) == 0 {
			return RoutineResult{
				Stop:       true,
				StopReason: "Cannot escape system",
			}
		}

		sort.Slice(jumpableIntermediaries, func(i, j int) bool {
			iDistance := util.CalcDistance(jumpableIntermediaries[i].X, jumpableIntermediaries[i].Y, targetSystem.X, targetSystem.Y)
			jDistance := util.CalcDistance(jumpableIntermediaries[j].X, jumpableIntermediaries[j].Y, targetSystem.X, targetSystem.Y)
			return iDistance < jDistance
		})

		state.Log(fmt.Sprintf("Jumping to intermediary system %s", jumpableIntermediaries[0].System))
		return RoutineResult{
			SetRoutine: GoToSystem{
				system: jumpableIntermediaries[0].System,
				next:   g,
			},
		}

	}

	_ = state.Ship.EnsureNavState(entity.NavOrbit)
	jumpResult, err := state.Ship.Jump(g.system)

	if err == nil {
		waitUntil := jumpResult.Cooldown.Expiration
		return RoutineResult{
			WaitUntil:  &waitUntil,
			SetRoutine: g.next,
		}
	}

	state.Log("Unable to jump")
	state.Log(err.Error())
	return RoutineResult{Stop: true, StopReason: err.Error()}
}

func (g GoToSystem) Name() string {
	return fmt.Sprintf("Go To System %s", g.system)
}
