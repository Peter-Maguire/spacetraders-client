package routine

import (
	"fmt"
	"sort"
	"spacetraders/constant"
	"spacetraders/database"
	"spacetraders/entity"
	"spacetraders/http"
	"spacetraders/util"
)

type GoToSystem struct {
	system entity.SystemSymbol
	next   Routine
}

func (g GoToSystem) Run(state *State) RoutineResult {

	if state.Ship.Nav.SystemSymbol == g.system {
		state.Log("We're already in this system")
		return RoutineResult{SetRoutine: g.next}
	}

	currentSystem := database.GetSystemData(string(state.Ship.Nav.SystemSymbol))
	if currentSystem == nil {
		currentSystem, _ = state.Ship.Nav.WaypointSymbol.GetSystem(state.Context)
		database.StoreSystem(currentSystem)
	}

	targetSystem := database.GetSystemData(string(g.system))

	if targetSystem == nil {
		//state.Log("TODO targetSystem isn't stored in the database!")
		//return RoutineResult{SetRoutine: FindNewSystem{}}
		targetSystem, _ = state.Agent.GetSystem(state.Context, string(g.system))
		database.StoreSystem(targetSystem)
		//database.AddUnvisitedSystems([]entity.System{*targetSystem}, 0)
	}

	distance := targetSystem.GetDistanceFrom(currentSystem)

	antimatterCargo := state.Ship.Cargo.GetSlotWithItem("ANTIMATTER")
	if antimatterCargo == nil || antimatterCargo.Units == 0 || distance > 500 {
		wpd, _ := state.Ship.Nav.WaypointSymbol.GetWaypointData(state.Context)
		if wpd.Type != constant.WaypointTypeJumpGate || wpd.IsUnderConstruction {
			state.Log("Going to jump gate")
			return RoutineResult{
				SetRoutine: GoToJumpGate{next: g},
			}
		}
	}

	if distance > 2000 {
		// Find an intermediary system that's within 2000 units
		intermediaries := database.GetVisitedSystems()
		jumpableIntermediaries := make([]database.System, 0)
		for _, intermediary := range intermediaries {
			if intermediary.System == string(g.system) || intermediary.System == string(state.Ship.Nav.SystemSymbol) {
				continue
			}
			if util.CalcDistance(currentSystem.X, currentSystem.Y, intermediary.X, intermediary.Y) < 2000 {
				jumpableIntermediaries = append(jumpableIntermediaries, intermediary)
			}
		}

		if len(jumpableIntermediaries) == 0 {
			state.Log("Cannot escape System - going exploring")
			return RoutineResult{
				SetRoutine: Explore{next: DetermineObjective{}},
			}
		}

		sort.Slice(jumpableIntermediaries, func(i, j int) bool {
			iDistance := util.CalcDistance(jumpableIntermediaries[i].X, jumpableIntermediaries[i].Y, targetSystem.X, targetSystem.Y)
			jDistance := util.CalcDistance(jumpableIntermediaries[j].X, jumpableIntermediaries[j].Y, targetSystem.X, targetSystem.Y)
			return iDistance > jDistance
		})

		state.Log(fmt.Sprintf("Jumping to intermediary system %s with distance %d", jumpableIntermediaries[0].System, util.CalcDistance(jumpableIntermediaries[0].X, jumpableIntermediaries[0].Y, targetSystem.X, targetSystem.Y)))
		return RoutineResult{
			SetRoutine: GoToSystem{
				system: entity.SystemSymbol(jumpableIntermediaries[0].System),
				next:   g,
			},
		}

	}

	_ = state.Ship.EnsureNavState(state.Context, entity.NavOrbit)
	systemData, _ := entity.GetSystem(state.Context, g.system)
	jumpGate := systemData.GetJumpGate(state.Context)
	if jumpGate == nil {
		return RoutineResult{
			Stop:       true,
			StopReason: "No Jump Gate found in system",
		}
	}
	jumpResult, err := state.Ship.Jump(state.Context, jumpGate.Symbol)

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
		_, _ = state.Ship.GetCargo(state.Context)
		return RoutineResult{}
		// TODO: figure out how this ship has gotten stuck here
		//case http.ErrAlreadyInSystem:
		//return RoutineResult{SetRoutine: g.next}
	}

	state.Log("Unable to jump")
	state.Log(err.Error())
	return RoutineResult{Stop: true, StopReason: err.Error()}
}

func (g GoToSystem) Name() string {
	return fmt.Sprintf("Go To System %s -> %s", g.system, g.next.Name())
}
