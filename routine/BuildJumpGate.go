package routine

import (
	"fmt"
	"spacetraders/constant"
	"spacetraders/database"
	"spacetraders/entity"
	"spacetraders/util"
)

type BuildJumpGate struct {
	next Routine
}

func (b BuildJumpGate) Run(state *State) RoutineResult {

	jumpGatesUnderConstruction := make([]*entity.WaypointData, 0)

	waypoints, _ := state.Ship.Nav.WaypointSymbol.GetSystemName().GetWaypointsOfType(state.Context, constant.WaypointTypeJumpGate)
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

	wpData := database.GetWaypoint(state.Ship.Nav.WaypointSymbol).GetData()

	util.SortWaypointsClosestTo(jumpGatesUnderConstruction, wpData.LimitedWaypointData)

	closestJumpGate := jumpGatesUnderConstruction[0]
	constructionSite, _ := closestJumpGate.Symbol.GetConstructionSite(state.Context)

	for _, material := range constructionSite.Materials {
		if material.IsComplete() {
			continue
		}

		state.Log(fmt.Sprintf("Construction site has %d/%d %s", material.Fulfilled, material.Required, material.TradeSymbol))
	}

	for _, state := range *state.States {
		state.ConstructionSite = constructionSite
	}

	return RoutineResult{
		SetRoutine: ProcureConstructionSiteItem{next: b.next},
	}
}

func (b BuildJumpGate) Name() string {
	return fmt.Sprintf("Build Jump Gate -> %s", b.next.Name())
}
