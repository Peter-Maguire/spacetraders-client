package routine

import (
	"spacetraders/constant"
	"spacetraders/database"
	"spacetraders/entity"
)

type Scrap struct {
}

func (s Scrap) Run(state *State) RoutineResult {
	wps := database.GetWaypoints()

	shipyards := make([]entity.WaypointData, 0)
	for _, wp := range wps {
		if wp.WaypointData.HasTrait(constant.TraitShipyard) {

			shipyards = append(shipyards, wp.WaypointData)
		}
	}

	return RoutineResult{}

}

func (s Scrap) Name() string {
	return "Scrap Ship"
}
