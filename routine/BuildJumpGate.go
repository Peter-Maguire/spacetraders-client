package routine

import (
	"fmt"
	"spacetraders/constant"
	"spacetraders/entity"
	"spacetraders/util"
)

type BuildJumpGate struct {
	next Routine
}

func (b BuildJumpGate) Run(state *State) RoutineResult {

	jumpGatesUnderConstruction := make([]*entity.WaypointData, 0)

	waypoints, _ := state.Ship.Nav.WaypointSymbol.GetSystemWaypoints(state.Context)
	for _, waypoint := range *waypoints {
		if waypoint.SystemSymbol == state.Ship.Nav.SystemSymbol && waypoint.Type == constant.WaypointTypeJumpGate {
			fullWp, _ := waypoint.GetFullWaypoint(state.Context)
			if fullWp.IsUnderConstruction {
				jumpGatesUnderConstruction = append(jumpGatesUnderConstruction, fullWp)
			}
		}
	}

	if len(jumpGatesUnderConstruction) == 0 {
		state.Log("No jump gates in this system are under construction")
		return RoutineResult{
			SetRoutine: b.next,
		}
	}

	// TODO: get from database
	wpData, _ := state.Ship.Nav.WaypointSymbol.GetWaypointData(state.Context)

	util.SortWaypointsClosestTo(jumpGatesUnderConstruction, wpData.LimitedWaypointData)

	closestJumpGate := jumpGatesUnderConstruction[0]
	if !state.Ship.IsAtWaypoint(closestJumpGate.Symbol) {
		state.Log("Going to under construction jump gate")
		return RoutineResult{
			SetRoutine: NavigateTo{waypoint: closestJumpGate.Symbol, next: b},
		}
	}

	constructionSite, _ := closestJumpGate.Symbol.GetConstructionSite(state.Context)

	for _, material := range constructionSite.Materials {
		if material.IsComplete() {
			continue
		}

		state.Log(fmt.Sprintf("Construction site has %d/%d %s", material.Fulfilled, material.Required, material.TradeSymbol))
	}

	return RoutineResult{
		SetRoutine: MineOres{},
	}
}

func (b BuildJumpGate) Name() string {
	return fmt.Sprintf("Build Jump Gate -> %s", b.next.Name())
}
