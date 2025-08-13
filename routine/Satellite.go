package routine

import (
	"fmt"
	"spacetraders/constant"
	"spacetraders/database"
	"spacetraders/entity"
	"spacetraders/metrics"
	"spacetraders/util"
	"strings"
)

type Satellite struct {
	shipsToBuy []string
}

func (s Satellite) Run(state *State) RoutineResult {

	/* TODO:
	*    If there is more than one satellite, each should go to a different shipyard or market
	*    to periodically update the prices and go to the next one
	*    e.g if there are 10 markets+shipyards and 2 satellites, they should each take the 5 closest ones and go between them one by one
	*    EDIT: waypoint data should probably be edited to have "times visited" so that they always go to the closest waypoint with
	*    less visits than the one they're currently on. It's also worth checking if there's any point to me going to waypoints that have no market or shipyard.
	 */

	state.Ship.EnsureNavState(state.Context, entity.NavDocked)

	waypointData, _ := state.Ship.Nav.WaypointSymbol.GetWaypointData(state.Context)

	var mkt *entity.Market
	var syd *entity.ShipyardStock

	if waypointData.HasTrait(constant.TraitMarketplace) {
		mkt, _ = waypointData.Symbol.GetMarket(state.Context)
	}

	if waypointData.HasTrait(constant.TraitShipyard) {
		syd, _ = waypointData.Symbol.GetShipyard(state.Context)
	}

	database.VisitWaypoint(waypointData, mkt, syd)

	agent, _ := entity.GetAgent(state.Context)
	metrics.NumCredits.WithLabelValues(state.Agent.Symbol).Set(float64(agent.Credits))

	if mkt != nil {
		system := database.GetSystemData(string(state.Ship.Nav.SystemSymbol))

		if len(mkt.TradeGoods) == 0 {
			state.Log("No trade goods... time desync?")
			return RoutineResult{
				WaitSeconds: 2,
			}
		}
		database.StoreMarketRates(system, waypointData, mkt.TradeGoods)
		if len(mkt.Exports) > 0 {
			database.StoreMarketExchange(system, waypointData, "export", mkt.Exports)
		}
		if len(mkt.Imports) > 0 {
			database.StoreMarketExchange(system, waypointData, "import", mkt.Imports)
		}
		if len(mkt.Exchange) > 0 {
			database.StoreMarketExchange(system, waypointData, "exchange", mkt.Exchange)
		}
	}

	if syd != nil {
		database.StoreShipCosts(syd)
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
	fmt.Println(len(sats))
	isSatZero := sats[0].Symbol == state.Ship.Symbol

	if state.Contract == nil && isSatZero {
		return RoutineResult{
			SetRoutine: NegotiateContract{},
		}
	}

	unvisitedWaypoints := database.GetUnvisitedWaypointsInSystem(string(state.Ship.Nav.SystemSymbol))

	goodUnvisitedWaypoints := false
	for _, uw := range unvisitedWaypoints {
		data := uw.GetData()
		if data.HasTrait(constant.TraitMarketplace) || data.HasTrait(constant.TraitShipyard) {
			goodUnvisitedWaypoints = true
			break
		}
	}

	if !goodUnvisitedWaypoints && state.Ship.Registration.Role != constant.ShipRoleSatellite {
		return RoutineResult{
			SetRoutine: DetermineObjective{},
		}
	}

	if goodUnvisitedWaypoints || !isSatZero {
		dbWaypoint := database.GetWaypoint(waypointData.Symbol)
		wps := database.GetLeastVisitedWaypointsInSystem(string(state.Ship.Nav.SystemSymbol), dbWaypoint.TimesVisited)
		if len(wps) == 0 {
			wps = database.GetLeastVisitedWaypointsInSystem(string(state.Ship.Nav.SystemSymbol), dbWaypoint.TimesVisited+1)
		}
		wpDatas := make([]*entity.WaypointData, len(wps))
		for i, wp := range wps {
			wpData := wp.GetData()
			wpDatas[i] = &wpData
		}

		goodWps := make([]*entity.WaypointData, 0)
		for _, wpData := range wpDatas {
			if wpData.HasTrait(constant.TraitMarketplace) || wpData.HasTrait(constant.TraitShipyard) {
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

		util.SortWaypointsClosestTo(wpDatas, waypointData.LimitedWaypointData)
		for _, wp := range wpDatas {
			shipsAtWaypoint := state.GetShipsWithRoleAtOrGoingToWaypoint(constant.ShipRoleSatellite, wp.Symbol)
			if len(shipsAtWaypoint) == 0 {
				return RoutineResult{SetRoutine: NavigateTo{waypoint: wp.Symbol, next: s}}
			}
		}
		state.Log("!!! Found no waypoints to go to")
	}

	shipsToBuy := s.GetShipToBuy(state)
	s.shipsToBuy = shipsToBuy
	for _, shipToBuy := range shipsToBuy {
		state.Log(fmt.Sprintf("We want to buy a %s", shipToBuy))

		shipCost := database.GetShipCost(shipToBuy, string(state.Ship.Nav.SystemSymbol))

		if shipCost == nil {
			state.Log("We aren't aware of any shipyards selling this ship type yet")
			waypoints, _ := state.Ship.Nav.WaypointSymbol.GetSystemName().GetWaypointsWithTrait(state.Context, constant.TraitShipyard)

			for _, w := range *waypoints {
				if w.HasTrait(constant.TraitShipyard) {
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
				WaitSeconds: 60,
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

		if shipCost.PurchasePrice > state.Agent.Credits {
			continue
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
			continue
		}

		state.Log(fmt.Sprintf("We need %d credits, we currently have %d credits", shipStock.PurchasePrice, state.Agent.Credits))

		if state.Agent.Credits >= shipStock.PurchasePrice {
			state.Log("We can buy a ship")
			shipyard, _ := state.Ship.Nav.WaypointSymbol.GetShipyard(state.Context)
			database.StoreShipCosts(shipyard)
			result, err := state.Agent.BuyShip(state.Context, state.Ship.Nav.WaypointSymbol, shipToBuy)

			if err != nil {
				state.Log(fmt.Sprintf("Error buying ship: %s", err.Error()))
				continue
			}
			database.LogTransaction("satellite", *result.Transaction)

			*state.EventBus <- OrchestratorEvent{Name: "newShip", Data: result.Ship}
			break
		}
	}

	// At this point we should be on the shipyard and waiting, so let's get the next sellComplete event and check there

	state.Log("Waiting for a sell to complete")
	return RoutineResult{
		WaitSeconds: 60,
	}
}

func (s Satellite) onShipyard(state *State, wpd *entity.WaypointData, targetShip string) bool {
	if !wpd.HasTrait(constant.TraitShipyard) {
		return false
	}

	shipyard, _ := wpd.Symbol.GetShipyard(state.Context)
	database.StoreShipCosts(shipyard)
	return shipyard.SellsShipType(targetShip)
}

func (s Satellite) Name() string {
	if s.shipsToBuy == nil {
		return "Satellite"
	}
	if len(s.shipsToBuy) > 0 {
		return fmt.Sprintf("Satellite (Buy %s)", strings.Join(s.shipsToBuy, ", "))
	}
	return "Satellite (Not buying)"
}

func (s Satellite) GetShipToBuy(state *State) []string {
	shipsOfEachType := make(map[constant.ShipRole]int)

	for _, st := range *state.States {
		shipsOfEachType[st.Ship.Registration.Role]++
	}

	state.Log("We currently have:")
	for t, a := range shipsOfEachType {
		state.Log(fmt.Sprintf("%dx of type %s", a, t))
	}

	if shipsOfEachType[constant.ShipRoleExcavator] < state.Config.GetInt("maxExcavatorsBeforeHauler", 4) {
		return []string{"SHIP_LIGHT_HAULER", "SHIP_MINING_DRONE"}
	}

	if shipsOfEachType[constant.ShipRoleExcavator] > 2 && state.Contract != nil && !state.Contract.Fulfilled && state.Agent.Credits < min(state.Contract.Terms.Payment.OnAccepted, state.Config.GetInt("minCreditsToIgnoreContract", 200000)) {
		for _, deliverable := range state.Contract.Terms.Deliver {
			if !util.IsMineable(deliverable.TradeSymbol) {
				state.Log(fmt.Sprintf("We don't want to buy a ship right now as we're doing a contract for unmineable %s", deliverable.TradeSymbol))
				if shipsOfEachType[constant.ShipRoleHauler] == 0 {
					return []string{"SHIP_LIGHT_HAULER", "SHIP_LIGHT_SHUTTLE"}
				}
				if shipsOfEachType[constant.ShipRoleTransport] == 0 {
					return []string{"SHIP_LIGHT_SHUTTLE"}
				}
				return []string{}
			}
		}
	}

	if shipsOfEachType[constant.ShipRoleTransport] < state.Config.GetInt("maxTraders", 4) {
		return []string{"SHIP_LIGHT_SHUTTLE"}
	}

	if shipsOfEachType[constant.ShipRoleHauler] == 0 {
		return []string{"SHIP_LIGHT_HAULER"}
	}

	if shipsOfEachType[constant.ShipRoleSurveyor] == 0 {
		return []string{"SHIP_SURVEYOR"}
	}

	if shipsOfEachType[constant.ShipRoleSatellite] < state.Config.GetInt("maxSatellites", 3) {
		return []string{"SHIP_PROBE"}
	}

	// Ratio of 1 hauler for every 5 excavators
	if shipsOfEachType[constant.ShipRoleHauler]/shipsOfEachType[constant.ShipRoleExcavator] < state.Config.GetInt("haulerToExcavatorRatio", 1/5) {
		return []string{"SHIP_LIGHT_HAULER"}
	}

	if len(*state.States) > state.Config.GetInt("maxShips", 30) {
		return []string{}
	}

	return []string{state.Config.GetString("defaultShipBuy", "SHIP_LIGHT_SHUTTLE")}
}
