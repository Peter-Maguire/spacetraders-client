package routine

import (
	"fmt"
	"math"
	"sort"
	"spacetraders/database"
	"spacetraders/entity"
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
					p.deliverable = &deliverable
					break
				}
			}
		}
	}

	if p.deliverable.IsFulfilled() {
		return RoutineResult{SetRoutine: DetermineObjective{}}
	}

	if state.Ship.Cargo.IsFull() {
		if state.Ship.Cargo.GetSlotWithItem(p.deliverable.TradeSymbol).Units > 0 {
			state.Log("We're already full and have something to deliver")
			return RoutineResult{
				SetRoutine: NavigateTo{
					waypoint: p.deliverable.DestinationSymbol,
					next: DeliverContractItem{
						item: p.deliverable.TradeSymbol,
						next: DetermineObjective{},
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

		sort.Slice(markets, func(i, j int) bool {
			iSystemDistance := util.CalcDistance(currentSystem.X, currentSystem.Y, markets[i].SystemX, markets[i].SystemY)
			iWaypointDistance := util.CalcDistance(lw.X, lw.Y, markets[i].WaypointX, markets[i].WaypointY)
			jSystemDistance := util.CalcDistance(currentSystem.X, currentSystem.Y, markets[j].SystemX, markets[j].SystemY)
			jWaypointDistance := util.CalcDistance(lw.X, lw.Y, markets[j].WaypointX, markets[j].WaypointY)
			return iSystemDistance+iWaypointDistance < jSystemDistance+jWaypointDistance
		})

		return RoutineResult{
			SetRoutine: NavigateTo{waypoint: markets[0].Waypoint, next: p},
		}
	}

	market, _ := state.Ship.Nav.WaypointSymbol.GetMarket()

	//database.UpdateMarketRates(state.Ship.Nav.WaypointSymbol, market.TradeGoods)

	tradeGood := market.GetTradeGood(p.deliverable.TradeSymbol)

	if tradeGood == nil {
		state.Log(fmt.Sprintf("%s is not sold here?", p.deliverable.TradeSymbol))
		return RoutineResult{
			WaitSeconds: 10,
		}
	}

	// Either the amount we can fit in our inventory, the amount we can afford or the amount we need - whichever is smaller
	purchaseAmount := int(math.Min(float64(state.Agent.Credits/tradeGood.PurchasePrice), math.Min(float64(p.deliverable.UnitsRequired-p.deliverable.UnitsFulfilled), float64(state.Ship.Cargo.GetRemainingCapacity()))))

	if purchaseAmount == 0 {
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
		state.Log("Enable to purchase: " + err.Error())
		return RoutineResult{
			Stop:       true,
			StopReason: err.Error(),
		}
	}

	if purchaseAmount >= p.deliverable.UnitsRequired-p.deliverable.UnitsFulfilled || purchaseAmount >= state.Ship.Cargo.GetRemainingCapacity() {
		return RoutineResult{
			SetRoutine: NavigateTo{
				waypoint: p.deliverable.DestinationSymbol,
				next: DeliverContractItem{
					item: p.deliverable.TradeSymbol,
					next: DetermineObjective{},
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
