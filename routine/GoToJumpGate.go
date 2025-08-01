package routine

import (
	"fmt"
	"spacetraders/constant"
)

type GoToJumpGate struct {
	next Routine
}

func (g GoToJumpGate) Run(state *State) RoutineResult {

	waypoints, _ := state.Ship.Nav.WaypointSymbol.GetSystemName().GetWaypointsOfType(state.Context, constant.WaypointTypeJumpGate)
	for _, waypoint := range *waypoints {
		if waypoint.Type == constant.WaypointTypeJumpGate {
			fullWp, _ := waypoint.GetFullWaypoint(state.Context)
			fmt.Println(fullWp)
			if fullWp.IsUnderConstruction {
				continue
			}
			if waypoint.Symbol == state.Ship.Nav.WaypointSymbol {
				state.Log("We're at a jump gate already")
				return RoutineResult{
					SetRoutine: g.next,
				}
			}
			state.Log("Going to jump gate")
			return RoutineResult{
				SetRoutine: NavigateTo{waypoint: waypoint.Symbol, next: g.next},
			}
		}
	}
	state.Log("Jump gate is under construction")

	return RoutineResult{SetRoutine: BuildJumpGate{next: g.next}}
}

func (g GoToJumpGate) Name() string {
	return fmt.Sprintf("Go To Jump Gate -> %s", g.next.Name())
}
