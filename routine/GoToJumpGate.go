package routine

import (
	"fmt"
	"spacetraders/constant"
)

type GoToJumpGate struct {
	next Routine
}

func (g GoToJumpGate) Run(state *State) RoutineResult {
	waypoints, _ := state.Ship.Nav.WaypointSymbol.GetSystemWaypoints(state.Context)
	for _, waypoint := range *waypoints {
		if waypoint.Type == constant.WaypointTypeJumpGate {
			if waypoint.Symbol == state.Ship.Nav.WaypointSymbol {
				state.Log("We're at a jump gate, already")
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
	state.Log("Unable to find jump-gate")
	return RoutineResult{Stop: true, StopReason: "Unable to find Jump Gate"}
}

func (g GoToJumpGate) Name() string {
	return fmt.Sprintf("Go To Jump Gate -> %s", g.next.Name())
}
