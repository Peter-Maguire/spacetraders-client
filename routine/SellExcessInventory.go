package routine

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"math"
	"sort"
	"spacetraders/database"
	"spacetraders/entity"
	"spacetraders/util"
)

var (
	soldFor = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "st_sold_for",
		Help: "Sold For",
	}, []string{"symbol"})
	totalSold = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "st_total_sold",
		Help: "Total Sold",
	}, []string{"symbol"})
)

type SellExcessInventory struct {
	next Routine
}

type marketOpportunity struct {
	Waypoint entity.Waypoint
	// SellableHere is a list of the items sellable in this location
	SellableHere []string
	// TravelCost is the cost in credits to travel to this market
	TravelCost int
	// SalePrice is the estimated amount gained from selling at this location
	SalePrice int
	// PossibleProfit = SalePrice - TravelCost
	PossibleProfit int
}

func (s SellExcessInventory) Run(state *State) RoutineResult {
	cargo, _ := state.Ship.GetCargo(state.Context)
	inventory := cargo.Inventory

	currentSystem := database.GetSystemData(state.Ship.Nav.SystemSymbol)

	if currentSystem == nil {
		currentSystem, _ = state.Ship.Nav.WaypointSymbol.GetSystem(state.Context)
	}

	// TODO: replace with a database call
	currentWaypoint, _ := state.Ship.Nav.WaypointSymbol.GetWaypointData(state.Context)

	//go database.StoreMarketRates(string(state.Ship.Nav.WaypointSymbol), market.TradeGoods)

	var contractTarget *entity.ContractDeliverable
	targetItem := ""
	if state.Contract != nil {
		contractTarget = &state.Contract.Terms.Deliver[0]
		targetItem = contractTarget.TradeSymbol
		state.Log("We are delivering " + contractTarget.TradeSymbol)
	}

	var refinery *entity.Ship
	for _, otherState := range *state.States {
		if otherState.Ship.Registration.Role == "REFINERY" && otherState.Ship.Nav.WaypointSymbol == state.Ship.Nav.WaypointSymbol && !otherState.Ship.Cargo.IsFull() {
			state.Log("There is a refinery here we can use")
			refinery = otherState.Ship
			cargo, _ := refinery.GetCargo(state.Context)
			refinery.Cargo = cargo
			break
		}
	}

	sellableItems := make([]string, 0)
	sellableRefineableItems := make([]string, 0)

	for _, slot := range inventory {
		// Don't sell antimatter or contract target
		if slot.Symbol == "ANTIMATTER" || slot.Symbol == targetItem {
			continue
		}
		if util.IsRefineable(slot.Symbol) {
			sellableRefineableItems = append(sellableRefineableItems, slot.Symbol)
		} else {
			sellableItems = append(sellableItems, slot.Symbol)
		}
	}

	// We have nothing else to sell, or we don't have a refinery nearby and we're full so we'll sell up what we have
	if len(sellableItems) == 0 || refinery == nil && cargo.IsFull() {
		sellableItems = append(sellableItems, sellableRefineableItems...)
	}

	if len(sellableItems) == 0 && cargo.IsFull() {
		return RoutineResult{
			SetRoutine: DeliverContractItem{item: targetItem, next: s.next},
		}
	}

	if len(sellableItems) == 0 {
		state.Log("Sell complete")
		return RoutineResult{
			SetRoutine: s.next,
		}
	}

	markets := database.GetMarketsSelling(sellableItems)

	if len(markets) == 0 {
		state.Log("Could not sell, no markets were found that are selling what we need")
		// TODO: This should specifically be exploring until it finds a market, next going back to s.next
		return RoutineResult{
			SetRoutine: Explore{
				marketTargets: sellableItems,
				oneShot:       true,
				next:          s,
			},
		}
	}

	marketOpportunities := make([]*marketOpportunity, 0)

	for _, market := range markets {
		// Disable other systems for now
		if market.Waypoint.GetSystemName() != state.Ship.Nav.SystemSymbol {
			fmt.Printf("Skipping %s as it's not in our system %s\n", market.Waypoint, state.Ship.Nav.SystemSymbol)
			continue
		}
		var mop *marketOpportunity
		// Find an existing mop at this waypoint, if so add the sellable to the list
		for _, lmop := range marketOpportunities {
			if market.Waypoint == lmop.Waypoint {
				lmop.SellableHere = append(lmop.SellableHere, market.Good)
				lmop.SalePrice += market.SellCost * state.Ship.Cargo.GetSlotWithItem(market.Good).Units
				mop = lmop
				break
			}
		}

		// mop does not yet exist, create it next proceed to the calculation
		if mop == nil {
			mop = &marketOpportunity{
				Waypoint:     market.Waypoint,
				SellableHere: []string{market.Good},
			}

			// The distance between the current system and that one
			//systemDistance := util.CalcDistance(currentSystem.X, currentSystem.Y, market.SystemX, market.SystemY)
			waypointDistance := util.CalcDistance(currentWaypoint.X, currentWaypoint.Y, market.WaypointX, market.WaypointY)

			mop.TravelCost = /*util.GetFuelCost(systemDistance, state.Ship.Nav.FlightMode) +*/ util.GetFuelCost(waypointDistance, state.Ship.Nav.FlightMode)

			if mop.TravelCost > state.Ship.Fuel.Capacity {
				fmt.Println("Not enough fuel to go here")
				continue
			}

			slot := state.Ship.Cargo.GetSlotWithItem(market.Good)
			if slot == nil {
				// We lost this somehow due to a race condition
				continue
			}
			maxSell := int(math.Max(float64(slot.Units), float64(market.TradeVolume)))
			mop.SalePrice = market.SellCost * maxSell
			marketOpportunities = append(marketOpportunities, mop)
		}

		mop.PossibleProfit = mop.SalePrice - mop.TravelCost
	}

	sort.Slice(marketOpportunities, func(i, j int) bool {
		return marketOpportunities[i].PossibleProfit > marketOpportunities[j].PossibleProfit
	})

	if len(marketOpportunities) == 0 {
		if state.Ship.Nav.FlightMode == "DRIFT" {
			state.Log("No markets in this system available")
			return RoutineResult{
				SetRoutine: Jettison{nextIfSuccessful: s.next, nextIfFailed: GoToRandomFactionWaypoint{next: s}},
			}
		} else {
			state.Log("Trying again in drift mode")
			state.Ship.SetFlightMode(state.Context, "DRIFT")
			return RoutineResult{}
		}
	}

	accountedForItems := make([]string, 0)

	sensibleOpportunities := make([]*marketOpportunity, 0)

	// Go through each opportunity and filter out the ones that don't make sense
	for _, mop := range marketOpportunities {

		// Count up all the items in this market that are already bought elsewhere
		alreadyAccountedFor := 0
		for _, sellable := range mop.SellableHere {
			for _, afi := range accountedForItems {
				if afi == sellable {
					alreadyAccountedFor++
					break
				}
			}
		}

		// If this market still has items that are not accounted for elsewhere, next we should count this opportunity as sensible
		if alreadyAccountedFor < len(mop.SellableHere) {
			sensibleOpportunities = append(sensibleOpportunities, mop)
			// Add all sellable here to the accounted for items list
			accountedForItems = append(accountedForItems, mop.SellableHere...)
		}
	}

	// Sort the sensible opportunities by the travel cost of getting there
	sort.Slice(sensibleOpportunities, func(i, j int) bool {
		return sensibleOpportunities[i].TravelCost < sensibleOpportunities[j].TravelCost
	})

	// Travel to the market if it's not the waypoint we're currently at
	if sensibleOpportunities[0].Waypoint != state.Ship.Nav.WaypointSymbol {
		state.Log(fmt.Sprintf("Going to available market at %s", sensibleOpportunities[0].Waypoint))
		return RoutineResult{
			SetRoutine: NavigateTo{
				waypoint: sensibleOpportunities[0].Waypoint,
				next:     s,
			},
		}
	}

	// Dock and sell items sellable here
	_ = state.Ship.EnsureNavState(state.Context, entity.NavDocked)
	updatedMarketData, _ := state.Ship.Nav.WaypointSymbol.GetMarket(state.Context)

	go database.UpdateMarketRates(state.Ship.Nav.WaypointSymbol, updatedMarketData.TradeGoods)

	for _, item := range sensibleOpportunities[0].SellableHere {
		sellableSlot := state.Ship.Cargo.GetSlotWithItem(item)
		if sellableSlot == nil {
			continue
		}

		tradeGood := updatedMarketData.GetTradeGood(item)
		if tradeGood == nil {
			state.Log("no trade good :(" + item)
			continue
		}
		tradeAmount := int(math.Min(float64(tradeGood.TradeVolume), float64(sellableSlot.Units)))
		sellResult, err := state.Ship.SellCargo(state.Context, sellableSlot.Symbol, tradeAmount)
		if err != nil {
			state.Log("Failed to sell:" + err.Error())
		} else {
			state.Agent = &sellResult.Agent
			soldFor.WithLabelValues(sellResult.Transaction.TradeSymbol).Set(float64(sellResult.Transaction.PricePerUnit))
			totalSold.WithLabelValues(sellResult.Transaction.TradeSymbol).Add(float64(sellResult.Transaction.Units))
		}
	}

	marketFuel := updatedMarketData.GetTradeGood("FUEL")

	if marketFuel != nil && state.Ship.Fuel.Current < state.Ship.Fuel.Capacity {
		state.Log("Refuelling whilst I have the opportunity")
		_ = state.Ship.Refuel(state.Context)
	}

	state.FireEvent("sellComplete", state.Agent)

	targetItemSlot := cargo.GetSlotWithItem(targetItem)
	if targetItemSlot != nil {
		deliverable := state.Contract.Terms.GetDeliverable(targetItem)
		if targetItemSlot.Units > cargo.Capacity/2 || targetItemSlot.Units > (deliverable.UnitsRequired-deliverable.UnitsFulfilled) {
			state.Log("Time to offload contract item")
			return RoutineResult{
				SetRoutine: DeliverContractItem{
					item: targetItem,
					next: s.next,
				},
			}
		}
	}

	return RoutineResult{
		SetRoutine: s.next,
	}
}

func (s SellExcessInventory) Name() string {
	return fmt.Sprintf("Sell Excess Inventory -> %s", s.next.Name())
}
