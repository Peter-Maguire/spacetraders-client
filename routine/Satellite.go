package routine

import (
	"fmt"
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

	shipyard, _ := state.Ship.Nav.WaypointSymbol.GetShipyard(state.Context)
	database.StoreShipCosts(shipyard)

	shipStock := shipyard.GetStockOf(shipToBuy)

	if shipStock == nil {
		state.Log("Somehow the ship isn't in stock here?")
		return RoutineResult{WaitSeconds: 30}
	}

	if state.Agent.Credits >= shipStock.PurchasePrice {
		state.Log("We can buy a ship")
		shipyard, _ := state.Ship.Nav.WaypointSymbol.GetShipyard(state.Context)
		database.StoreShipCosts(shipyard)
		// TODO: check the prices again?
		result, err := state.Agent.BuyShip(state.Context, state.Ship.Nav.WaypointSymbol, shipToBuy)

		if err != nil {
			state.Log(fmt.Sprintf("Error buying ship: %s", err.Error()))
			return RoutineResult{WaitSeconds: 30}
		}

		state.EventBus <- OrchestratorEvent{Name: "newShip", Data: result.Ship}
	}

	// At this point we should be on the shipyard and waiting, so let's get the next sellComplete event and check there

	state.Log("Waiting for a sell to complete")
	for {
		event := <-state.EventBus
		switch event.Name {
		case "sellComplete":
			//agent := event.Data.(*entity.Agent)
			state.Log("Detected a sell complete")
			return RoutineResult{}
		}
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
