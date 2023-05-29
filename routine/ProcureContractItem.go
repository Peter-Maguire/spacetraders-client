package routine

import (
	"fmt"
	"math"
	"sort"
	"spacetraders/database"
	"spacetraders/entity"
	"spacetraders/http"
	"spacetraders/util"
)

type ProcureContractItem struct {
	deliverable *entity.ContractDeliverable
}

func (p ProcureContractItem) Run(state *State) RoutineResult {

	contracts, _ := state.Agent.Contracts()

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

	for _, m := range markets {
		if m.Waypoint == state.Ship.Nav.WaypointSymbol {
			isCurrentlyAtMarket = true
			break
		}
	}

	if !isCurrentlyAtMarket {
		state.Log("Going to closest market selling this item")
		currentSystem := database.GetSystemData(state.Ship.Nav.SystemSymbol)

		if currentSystem == nil {
			currentSystem, _ = state.Ship.Nav.WaypointSymbol.GetSystem()
		}

		lw := currentSystem.GetLimitedWaypoint(state.Ship.Nav.WaypointSymbol)

		marketCosts := make(map[entity.Waypoint]int)

		for _, market := range markets {
			systemDistance := util.CalcDistance(currentSystem.X, currentSystem.Y, market.SystemX, market.SystemY)
			waypointDistance := util.CalcDistance(lw.X, lw.Y, market.WaypointX, market.WaypointY)
			// TODO: this doesn't take into account multiple round trips required for larger procurement contracts
			travelCost := util.GetFuelCost(systemDistance, state.Ship.Nav.FlightMode) + util.GetFuelCost(waypointDistance, state.Ship.Nav.FlightMode)
			saleCost := market.BuyCost * unitsRemaining
			marketCosts[market.Waypoint] = travelCost + saleCost
		}

		sort.Slice(markets, func(i, j int) bool {
			return marketCosts[markets[i].Waypoint] < marketCosts[markets[j].Waypoint]
		})

		state.Log(fmt.Sprintf("Cost of retrieving %dx %s at cheapest market is %d", unitsRemaining, p.deliverable.TradeSymbol, marketCosts[markets[0].Waypoint]))

		if marketCosts[markets[0].Waypoint] > state.Contract.Terms.Payment.GetTotalPayment() {
			return RoutineResult{
				Stop:       true,
				StopReason: fmt.Sprintf("%dx %s = %d vs %d total contract payment", unitsRemaining, p.deliverable.TradeSymbol, marketCosts[markets[0].Waypoint], state.Contract.Terms.Payment.GetTotalPayment()),
			}
		}

		return RoutineResult{
			SetRoutine: NavigateTo{waypoint: markets[0].Waypoint, next: p},
		}
	}

	market, _ := state.Ship.Nav.WaypointSymbol.GetMarket()

	database.UpdateMarketRates(state.Ship.Nav.WaypointSymbol, market.TradeGoods)

	tradeGood := market.GetTradeGood(p.deliverable.TradeSymbol)

	if tradeGood == nil {
		state.Log(fmt.Sprintf("%s is not sold here?", p.deliverable.TradeSymbol))
		return RoutineResult{
			WaitSeconds: 10,
		}
	}

	// Either the amount we can fit in our inventory, the trade volume, the amount we can afford or the amount we need - whichever is smaller
	purchaseAmount := int(math.Min(math.Min(float64(state.Agent.Credits/tradeGood.PurchasePrice), float64(tradeGood.TradeVolume)), math.Min(float64(unitsRemaining), float64(state.Ship.Cargo.GetRemainingCapacity()))))

	if purchaseAmount <= 0 {
		state.Log("We're not able to purchase anything right now for some reason")
		return RoutineResult{
			WaitSeconds: 120,
		}
	}

	state.Log(fmt.Sprintf("Attempting to purchase %dx %s", purchaseAmount, p.deliverable.TradeSymbol))
	state.WaitingForHttp = true
	_ = state.Ship.EnsureNavState(entity.NavDocked)
	state.WaitingForHttp = false

	_, err := state.Ship.Purchase(p.deliverable.TradeSymbol, purchaseAmount)

	if err != nil {

		switch err.Code {
		case http.ErrInsufficientFunds:
			state.Log("Insufficient Funds")
			agent, _ := entity.GetAgent()
			state.Agent.Credits = agent.Credits
			return RoutineResult{
				WaitSeconds: 10,
			}
		}

		state.Log("Enable to purchase: " + err.Error())
		return RoutineResult{
			Stop:       true,
			StopReason: err.Error(),
		}
	}

	if purchaseAmount >= unitsRemaining || purchaseAmount >= state.Ship.Cargo.GetRemainingCapacity() {
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

	state.Log("Waiting to buy some more")
	return RoutineResult{
		WaitSeconds: 60,
	}

}

func (p ProcureContractItem) Name() string {
	return fmt.Sprintf("Procure %dx %s", p.deliverable.UnitsRequired-p.deliverable.UnitsFulfilled, p.deliverable.TradeSymbol)
}
