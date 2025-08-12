package routine

import (
	"fmt"
	"sort"
	"spacetraders/database"
	"spacetraders/entity"
	"spacetraders/http"
	"spacetraders/metrics"
	"spacetraders/util"
)

type ProcureContractItem struct {
	deliverable   *entity.ContractDeliverable
	hasSoldExcess bool
}

func (p ProcureContractItem) Run(state *State) RoutineResult {

	contracts, _ := state.Agent.Contracts(state.Context)

	for _, contract := range *contracts {
		if contract.Accepted && !contract.Fulfilled {
			for _, deliverable := range contract.Terms.Deliver {
				if deliverable.TradeSymbol == p.deliverable.TradeSymbol {
					state.Contract = &contract
					p.deliverable = &deliverable
					break
				}
			}
		}
	}

	if !p.hasSoldExcess {
		for _, slot := range state.Ship.Cargo.Inventory {
			if slot.Symbol != "FUEL" && slot.Symbol != p.deliverable.TradeSymbol {
				state.Log("Get rid of what we have to sell first")
				p.hasSoldExcess = true
				return RoutineResult{
					SetRoutine: SellExcessInventory{next: p},
				}
			}
		}
	}

	if p.deliverable.IsFulfilled() {
		return RoutineResult{SetRoutine: DetermineObjective{}}
	}

	unitsRemaining := p.deliverable.UnitsRequired - p.deliverable.UnitsFulfilled
	currentCargo := state.Ship.Cargo.GetSlotWithItem(p.deliverable.TradeSymbol)
	if currentCargo != nil {
		unitsRemaining -= currentCargo.Units
	}

	if state.Ship.Cargo.IsFull() || unitsRemaining <= 0 {
		if state.Ship.Cargo.GetSlotWithItem(p.deliverable.TradeSymbol).Units > 0 {
			state.Log("We're already full and have something to deliver")
			return RoutineResult{
				SetRoutine: NavigateTo{
					waypoint: p.deliverable.DestinationSymbol,
					next: DeliverContractItem{
						item: p.deliverable.TradeSymbol,
						next: NegotiateContract{},
					},
				},
			}
		}
		state.Log("We're full of something else")
		return RoutineResult{SetRoutine: SellExcessInventory{next: p}}
	}

	markets := database.GetMarketsSelling([]string{p.deliverable.TradeSymbol})

	if len(markets) == 0 {
		return RoutineResult{
			Stop:       true,
			StopReason: fmt.Sprintf("No markets selling %s", p.deliverable.TradeSymbol),
		}
	}

	isCurrentlyAtMarket := false

	//for _, m := range markets {
	//	if m.Waypoint == state.Ship.Nav.WaypointSymbol {
	//		isCurrentlyAtMarket = true
	//		break
	//	}
	//}

	if !isCurrentlyAtMarket {
		state.Log("Going to closest market selling this item")
		currentSystem := database.GetSystemData(string(state.Ship.Nav.SystemSymbol))

		if currentSystem == nil {
			currentSystem, _ = state.Ship.Nav.WaypointSymbol.GetSystem(state.Context)
		}

		lw := currentSystem.GetLimitedWaypoint(state.Context, state.Ship.Nav.WaypointSymbol)

		marketCosts := make(map[entity.Waypoint]int)

		for _, market := range markets {
			if market.Waypoint.GetSystemName() != state.Ship.Nav.SystemSymbol {
				// TODO: make this less stupid
				marketCosts[market.Waypoint] = 999999999
				continue
			}
			systemDistance := util.CalcDistance(currentSystem.X, currentSystem.Y, market.SystemX, market.SystemY)
			waypointDistance := util.CalcDistance(lw.X, lw.Y, market.WaypointX, market.WaypointY)
			roundTrips := min(1, unitsRemaining/state.Ship.Cargo.Capacity)
			travelCost := (util.GetFuelCost(systemDistance, state.Ship.Nav.FlightMode) + util.GetFuelCost(waypointDistance, state.Ship.Nav.FlightMode)) * roundTrips
			saleCost := market.BuyCost * unitsRemaining
			marketCosts[market.Waypoint] = travelCost + saleCost
		}

		sort.Slice(markets, func(i, j int) bool {
			return marketCosts[markets[i].Waypoint] < marketCosts[markets[j].Waypoint]
		})

		state.Log(fmt.Sprintf("Cost of retrieving %dx %s at cheapest market (%s) is %d", unitsRemaining, p.deliverable.TradeSymbol, markets[0].Waypoint, marketCosts[markets[0].Waypoint]))

		if marketCosts[markets[0].Waypoint] > state.Contract.Terms.Payment.GetTotalPayment() {
			if util.IsMineable(p.deliverable.TradeSymbol) {
				return RoutineResult{
					SetRoutine: GoToMiningArea{next: MineOres{next: p}},
				}
			}
			state.Log("Having a look for more markets")
			return RoutineResult{
				SetRoutine: Explore{
					marketTargets: []string{p.deliverable.TradeSymbol},
					next:          p,
				},
			}
			//return RoutineResult{
			//	Stop:       true,
			//	StopReason: fmt.Sprintf("%dx %s = %d vs %d total contract payment", unitsRemaining, p.deliverable.TradeSymbol, marketCosts[markets[0].Waypoint], state.Contract.Terms.Payment.GetTotalPayment()),
			//}
		}

		if markets[0].Waypoint != state.Ship.Nav.WaypointSymbol {
			return RoutineResult{
				SetRoutine: NavigateTo{waypoint: markets[0].Waypoint, next: p},
			}
		}
	}

	market, _ := state.Ship.Nav.WaypointSymbol.GetMarket(state.Context)

	database.UpdateMarketRates(state.Ship.Nav.WaypointSymbol, market.TradeGoods)

	tradeGood := market.GetTradeGood(p.deliverable.TradeSymbol)

	if tradeGood == nil {
		state.Log(fmt.Sprintf("%s is not sold here?", p.deliverable.TradeSymbol))
		return RoutineResult{
			WaitSeconds: 10,
		}
	}

	// Either the amount we can fit in our inventory, the trade volume, the amount we can afford or the amount we need - whichever is smaller
	amountPurchasable := state.Agent.Credits / tradeGood.PurchasePrice
	tradeVolume := tradeGood.TradeVolume
	remainingCapacity := state.Ship.Cargo.GetRemainingCapacity()
	purchaseAmount := min(amountPurchasable, tradeVolume, unitsRemaining, remainingCapacity)

	if purchaseAmount <= 0 {
		state.Log("We're not able to purchase anything right now")
		fmt.Println(state.Agent.Credits, tradeGood.PurchasePrice)
		fmt.Println(amountPurchasable, tradeVolume, remainingCapacity, purchaseAmount)
		return RoutineResult{
			WaitForEvent: "sellComplete",
		}
	}

	state.Log(fmt.Sprintf("Attempting to purchase %dx %s", purchaseAmount, p.deliverable.TradeSymbol))
	_ = state.Ship.EnsureNavState(state.Context, entity.NavDocked)

	pr, err := state.Ship.Purchase(state.Context, p.deliverable.TradeSymbol, purchaseAmount)

	if err != nil {

		switch err.Code {
		case http.ErrMarketTradeInsufficientCredits:
			state.Log("Insufficient Funds")
			agent, _ := entity.GetAgent(state.Context)
			metrics.NumCredits.WithLabelValues(state.Agent.Symbol).Set(float64(agent.Credits))
			state.Agent.Credits = agent.Credits
			return RoutineResult{
				WaitSeconds: 10,
			}
		}

		state.Log("Unable to purchase: " + err.Error())
		return RoutineResult{
			Stop:       true,
			StopReason: err.Error(),
		}
	}

	database.LogTransaction("contract", *pr.Transaction)

	sellFuel := market.GetTradeGood("FUEL")

	if sellFuel != nil && state.Ship.Fuel.Capacity-state.Ship.Fuel.Current > 100 {
		state.Log("Refuelling whilst I can")
		rr, _ := state.Ship.Refuel(state.Context)
		if rr != nil {
			database.LogTransaction("contract_refuel", rr.Transaction)
		}
	}

	if purchaseAmount >= unitsRemaining || purchaseAmount >= state.Ship.Cargo.GetRemainingCapacity() {
		state.Log("going to a random waypoint to negotiate after delivery")
		return RoutineResult{
			SetRoutine: NavigateTo{
				waypoint: p.deliverable.DestinationSymbol,
				next: DeliverContractItem{
					item: p.deliverable.TradeSymbol,
					next: GoToRandomFactionWaypoint{
						next: NegotiateContract{},
					},
				},
			},
		}
	}

	state.Log("Waiting to buy some more")
	return RoutineResult{
		WaitSeconds: 60,
	}

}

func (p ProcureContractItem) Name() string {
	return fmt.Sprintf("Procure %dx %s", p.deliverable.UnitsRequired-p.deliverable.UnitsFulfilled, p.deliverable.TradeSymbol)
}
