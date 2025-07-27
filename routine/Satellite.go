package routine

import (
	"fmt"
	"spacetraders/constant"
	"spacetraders/database"
	"spacetraders/entity"
	"spacetraders/util"
)

type Satellite struct{}

func (s Satellite) Run(state *State) RoutineResult {

	/* TODO:
	*    If there is more than one satellite, each should go to a different shipyard or market
	*    to periodically update the prices and go to the next one
	*    e.g if there are 10 markets+shipyards and 2 satellites, they should each take the 5 closest ones and go between them one by one
	*    EDIT: waypoint data should probably be edited to have "times visited" so that they always go to the closest waypoint with
	*    less visits than the one they're currently on. It's also worth checking if there's any point to me going to waypoints that have no market or shipyard.
	 */

	waypointData, _ := state.Ship.Nav.WaypointSymbol.GetWaypointData(state.Context)

	mkt, _ := waypointData.Symbol.GetMarket(state.Context)
	syd, _ := waypointData.Symbol.GetShipyard(state.Context)

	database.VisitWaypoint(waypointData, mkt, syd)

	if state.Contract == nil {
		return RoutineResult{
			SetRoutine: NegotiateContract{},
		}
	}

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

	sats := state.GetShipsWithRole(constant.ShipRoleSatellite)
	isSatZero := len(sats) > 1 && sats[0].Symbol == state.Ship.Symbol

	if !isSatZero {
		dbWaypoint := database.GetWaypoint(waypointData.Symbol)
		fmt.Printf("Looking for waypoitns visited less than %d times\n", dbWaypoint.TimesVisited)
		wps := database.GetLeastVisitedWaypointsInSystem(state.Ship.Nav.SystemSymbol, dbWaypoint.TimesVisited)
		if len(wps) == 0 {
			fmt.Printf("Looking for waypoitns visited less than %d times\n", dbWaypoint.TimesVisited+1)
			wps = database.GetLeastVisitedWaypointsInSystem(state.Ship.Nav.SystemSymbol, dbWaypoint.TimesVisited+1)
		}
		wpDatas := make([]*entity.WaypointData, len(wps))
		for i, wp := range wps {
			wpData := wp.GetData()
			wpDatas[i] = &wpData
		}

		goodWps := make([]*entity.WaypointData, 0)
		for _, wpData := range wpDatas {
			if wpData.HasTrait("MARKETPLACE") || wpData.HasTrait("SHIPYARD") {
				goodWps = append(goodWps, wpData)
			}
		}
		if len(goodWps) > 0 {
			wpDatas = goodWps
		} else if state.Ship.Registration.Role != constant.ShipRoleSatellite {
			return RoutineResult{
				SetRoutine: DetermineObjective{},
			}
		}

		fmt.Println(wpDatas, waypointData)
		util.SortWaypointsClosestTo(wpDatas, waypointData.LimitedWaypointData)
		for _, wp := range wpDatas {
			shipsAtWaypoint := state.GetShipsWithRoleAtOrGoingToWaypoint(constant.ShipRoleSatellite, wp.Symbol)
			if len(shipsAtWaypoint) == 0 {
				return RoutineResult{SetRoutine: NavigateTo{waypoint: wp.Symbol, next: s}}
			}
			fmt.Println("Ship already going to ", wp.Symbol)
		}
		state.Log("!!! Found no waypoints to go to")
	}

	shipToBuy := s.GetShipToBuy(state)

	state.Log(fmt.Sprintf("We want to buy a %s", shipToBuy))

	shipCost := database.GetShipCost(shipToBuy, state.Ship.Nav.SystemSymbol)

	if shipCost == nil {
		state.Log("We aren't aware of any shipyards selling this ship type yet")
		waypoints, _ := state.Ship.Nav.WaypointSymbol.GetSystemWaypoints(state.Context)

		for _, w := range *waypoints {
			if w.HasTrait("SHIPYARD") {
				shipyard, _ := w.Symbol.GetShipyard(state.Context)
				if shipyard.SellsShipType(shipToBuy) {
					state.Log("Found shipyard selling the desired ship")
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

	if shipCost.Waypoint == "" {
		fmt.Println(shipCost)
		return RoutineResult{
			Stop:       true,
			StopReason: "ship cost fuck",
		}
	}
	if shipCost.Waypoint != string(state.Ship.Nav.WaypointSymbol) {
		state.Log("Going to where this ship is cheapest")
		return RoutineResult{
			SetRoutine: NavigateTo{waypoint: entity.Waypoint(shipCost.Waypoint), next: s},
		}
	}

	shipyard, _ := state.Ship.Nav.WaypointSymbol.GetShipyard(state.Context)
	database.StoreShipCosts(shipyard)

	shipStock := shipyard.GetStockOf(shipToBuy)

	if shipStock == nil {
		state.Log("Somehow the ship isn't in stock here?")
		return RoutineResult{WaitSeconds: 30}
	}

	state.Log(fmt.Sprintf("We need %d credits, we currently have %d credits", shipStock.PurchasePrice, state.Agent.Credits))

	if state.Agent.Credits >= shipStock.PurchasePrice {
		state.Log("We can buy a ship")
		shipyard, _ := state.Ship.Nav.WaypointSymbol.GetShipyard(state.Context)
		database.StoreShipCosts(shipyard)
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
		case "sellComplete", "contractComplete":
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

func (s Satellite) GetShipToBuy(state *State) string {
	shipsOfEachType := make(map[constant.ShipRole]int)

	for _, st := range *state.States {
		shipsOfEachType[st.Ship.Registration.Role]++
	}

	state.Log("We currently have:")
	for t, a := range shipsOfEachType {
		state.Log(fmt.Sprintf("%dx of type %s", a, t))
	}

	if shipsOfEachType[constant.ShipRoleExcavator] < 5 {
		return "SHIP_MINING_DRONE"
	}

	if shipsOfEachType[constant.ShipRoleHauler] == 0 {
		return "SHIP_LIGHT_HAULER"
	}

	if shipsOfEachType[constant.ShipRoleSurveyor] == 0 {
		return "SHIP_SURVEYOR"
	}

	return "SHIP_MINING_DRONE"
}
