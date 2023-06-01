package routine

import (
	"fmt"
	"sort"
	"spacetraders/database"
	"spacetraders/entity"
	"spacetraders/http"
	"spacetraders/util"
)

type GoToSystem struct {
	system string
	next   Routine
}

func (g GoToSystem) Run(state *State) RoutineResult {

	if state.Ship.Nav.SystemSymbol == g.system {
		return RoutineResult{SetRoutine: g.next}
	}

	currentSystem := database.GetSystemData(state.Ship.Nav.SystemSymbol)
	if currentSystem != nil {
		currentSystem, _ = state.Ship.Nav.WaypointSymbol.GetSystem()
		state.Log("TODO currentSystem isn't stored in the database!")
	}

	targetSystem := database.GetSystemData(g.system)

	if targetSystem != nil {
		state.Log("TODO targetSystem isn't stored in the database!")
		targetSystem, _ = state.Agent.GetSystem(g.system)
		database.AddUnvisitedSystems([]entity.System{*targetSystem}, 0)
	}

	distance := currentSystem.GetDistanceFrom(targetSystem)

	antimatterCargo := state.Ship.Cargo.GetSlotWithItem("ANTIMATTER")
	if antimatterCargo == nil || antimatterCargo.Units == 0 || distance > 500 {
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

	switch err.Code {
	case http.ErrInsufficientAntimatter:
		state.Log("Cargo must be out of date, retrying")
		_, _ = state.Ship.GetCargo()
		return RoutineResult{}
	case http.ErrAlreadyInSystem:
		return RoutineResult{SetRoutine: g.next}
	}

	state.Log("Unable to jump")
	state.Log(err.Error())
	return RoutineResult{Stop: true, StopReason: err.Error()}
}

func (g GoToSystem) Name() string {
	return fmt.Sprintf("Go To System %s", g.system)
}
