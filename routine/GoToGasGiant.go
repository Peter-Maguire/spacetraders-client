package routine

import (
	"sort"
	"spacetraders/constant"
	"spacetraders/database"
	"spacetraders/entity"
)

type GoToGasGiant struct {
	blacklist []entity.Waypoint
}

func (f GoToGasGiant) Run(state *State) RoutineResult {

	waypoints := database.GetWaypoints()

	gasGiants := make([]database.ScannedWaypoint, 0)
	var currentWaypoint database.ScannedWaypoint
	for _, waypoint := range waypoints {
		if waypoint.Waypoint == state.Ship.Nav.WaypointSymbol {
			currentWaypoint = waypoint
		}
		if f.isBlacklisted(waypoint.Waypoint) {
			continue
		}
		if waypoint.WaypointData.Type == constant.WaypointTypeGasGiant {
			gasGiants = append(gasGiants, waypoint)
		}
	}

	if len(gasGiants) > 0 {
		// TODO: find other systems with gas giants
		return RoutineResult{
			//SetRoutine: Explore{},
			Stop:       true,
			StopReason: "No gas giants available",
		}
	}

	sort.Slice(gasGiants, func(i, j int) bool {
		return gasGiants[i].WaypointData.GetDistanceFrom(currentWaypoint.WaypointData.LimitedWaypointData) < gasGiants[j].WaypointData.GetDistanceFrom(currentWaypoint.WaypointData.LimitedWaypointData)
	})

	return RoutineResult{
		SetRoutine: NavigateTo{
			waypoint: gasGiants[0].Waypoint,
			next:     SiphonGas{},
		},
	}
}

func (f GoToGasGiant) isBlacklisted(waypoint entity.Waypoint) bool {
	for _, bl := range f.blacklist {
		if bl == waypoint {
			return true
		}
	}
	return false
}

func (f GoToGasGiant) Name() string {
	return "Go To Gas Giant"
}
