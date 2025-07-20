package routine

import (
	"spacetraders/database"
	"spacetraders/entity"
)

type Satellite struct{}

func (s Satellite) Run(state *State) RoutineResult {

	/* TODO: If there is more than one satellite, each should go to a different shipyard or market
	*  to periodically update the prices and go to the next one
	* e.g if there are 10 markets+shipyards and 2 satellites, they should each take the 5 closest ones and go between them one by one
	 */

	marketRates := database.GetMarkets()
	if len(marketRates) == 0 {
		return RoutineResult{
			SetRoutine: Explore{
				oneShot:      true,
				desiredTrait: "MARKETPLACE",
				next:         s,
			},
		}
	}

	waypointData, _ := state.Ship.Nav.WaypointSymbol.GetWaypointData(state.Context)

	// TODO: this from orchestrator
	shipToBuy := "SHIP_MINING_DRONE"

	if !s.onShipyard(state, waypointData, shipToBuy) {

		// TODO this should use the database data
		waypoints, _ := state.Ship.Nav.WaypointSymbol.GetSystemWaypoints(state.Context)

		for _, w := range *waypoints {
			if w.HasTrait("SHIPYARD") {
				shipyard, _ := w.Symbol.GetShipyard(state.Context)
				if shipyard.SellsShipType(shipToBuy) {
					return RoutineResult{
						SetRoutine: NavigateTo{
							waypoint: w.Symbol,
							next:     s,
						},
					}
				}
				// TODO: find the closest
			}
		}

		return RoutineResult{
			SetRoutine: Explore{
				oneShot:      true,
				desiredTrait: "SHIPYARD",
				next:         s,
			},
		}
	}

	//shipyard, _ := state.Ship.Nav.WaypointSymbol.GetShipyard(state.Context)
	//database.StoreShipCosts(shipyard)
	//
	//// TODO: Should use the eventbus to create the new ship here instead of this
	//result, err := state.Agent.BuyShip(state.Context, state.Ship.Nav.WaypointSymbol, shipToBuy)
	//if err != nil && result != nil {
	//	//newState := State{
	//	//	Agent:    state.Agent,
	//	//	Contract: state.Contract,
	//	//	Ship:     result.Ship,
	//	//	Haulers:  state.Haulers,
	//	//	EventBus: state.EventBus,
	//	//	States:   state.States,
	//	//}
	//	ui.MainLog(fmt.Sprintln("New ship", result.Ship.Symbol))
	//	//o.States = append(o.States, &state)
	//	//go .routineLoop(&state)
	//}

	return RoutineResult{
		WaitSeconds: 120,
	}

}

func (s Satellite) onShipyard(state *State, wpd *entity.WaypointData, targetShip string) bool {
	if !wpd.HasTrait("SHIPYARD") {
		return false
	}

	shipyard, _ := wpd.Symbol.GetShipyard(state.Context)
	database.StoreShipCosts(shipyard)
	return shipyard.SellsShipType(targetShip)
}

func (s Satellite) Name() string {
	return "Satellite"
}
