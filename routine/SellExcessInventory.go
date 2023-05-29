package routine

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
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
	inventory := state.Ship.Cargo.Inventory

	currentSystem := database.GetSystemData(state.Ship.Nav.SystemSymbol)

	if currentSystem == nil {
		currentSystem, _ = state.Ship.Nav.WaypointSymbol.GetSystem()
	}

	// TODO: replace with a database call
	currentWaypoint, _ := state.Ship.Nav.WaypointSymbol.GetWaypointData()

	fmt.Println(currentSystem)

	//go database.StoreMarketRates(string(state.Ship.Nav.WaypointSymbol), market.TradeGoods)

	var contractTarget *entity.ContractDeliverable
	targetItem := ""
	if state.Contract != nil {
		contractTarget = &state.Contract.Terms.Deliver[0]
		targetItem = contractTarget.TradeSymbol
		state.Log("We are delivering " + contractTarget.TradeSymbol)
	}

	sellableItems := make([]string, 0)

	for _, slot := range inventory {
		// Don't sell antimatter or contract target
		if slot.Symbol == "ANTIMATTER" || slot.Symbol == targetItem {
			continue
		}
		sellableItems = append(sellableItems, slot.Symbol)
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
		return RoutineResult{
			Stop:       true,
			StopReason: "No markets available to sell to",
		}
	}

	marketOpportunities := make([]*marketOpportunity, 0)

	for _, market := range markets {
		// Disable other systems for now
		if market.Waypoint.GetSystemName() != state.Ship.Nav.SystemSymbol {
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

		// mop does not yet exist, create it then proceed to the calculation
		if mop == nil {
			mop = &marketOpportunity{
				Waypoint:     market.Waypoint,
				SellableHere: []string{market.Good},
			}

			// The distance between the current system and that one
			systemDistance := util.CalcDistance(currentSystem.X, currentSystem.Y, market.SystemX, market.SystemY)
			waypointDistance := util.CalcDistance(currentWaypoint.X, currentWaypoint.Y, market.WaypointX, market.WaypointY)

			mop.TravelCost = systemDistance + util.GetFuelCost(waypointDistance, state.Ship.Nav.FlightMode)
			mop.SalePrice = market.SellCost * state.Ship.Cargo.GetSlotWithItem(market.Good).Units
			marketOpportunities = append(marketOpportunities, mop)
		}

		mop.PossibleProfit = mop.SalePrice - mop.TravelCost
	}

	if len(marketOpportunities) == 0 {
		state.Log("No markets in this system available")
		return RoutineResult{
			Stop:       true,
			StopReason: "No markets available to sell to in this system",
		}
	}

	sort.Slice(marketOpportunities, func(i, j int) bool {
		return marketOpportunities[i].PossibleProfit > marketOpportunities[j].PossibleProfit
	})

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
		// If this market still has items that are not accounted for elsewhere, then we should count this opportunity as sensible
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
	state.WaitingForHttp = true
	_ = state.Ship.EnsureNavState(entity.NavDocked)
	updatedMarketData, _ := state.Ship.Nav.WaypointSymbol.GetMarket()
	state.WaitingForHttp = false

	go database.UpdateMarketRates(state.Ship.Nav.WaypointSymbol, updatedMarketData.TradeGoods)

	for _, item := range sensibleOpportunities[0].SellableHere {
		sellableSlot := state.Ship.Cargo.GetSlotWithItem(item)
		state.WaitingForHttp = true
		sellResult, err := state.Ship.SellCargo(sellableSlot.Symbol, sellableSlot.Units)
		state.WaitingForHttp = false
		if err != nil {
			state.Log("Failed to sell:" + err.Error())
		} else {
			state.Agent = &sellResult.Agent
			soldFor.WithLabelValues(sellResult.Transaction.TradeSymbol).Set(float64(sellResult.Transaction.PricePerUnit))
			totalSold.WithLabelValues(sellResult.Transaction.TradeSymbol).Add(float64(sellResult.Transaction.Units))
		}
	}
	state.FireEvent("sellComplete", state.Agent)

	return RoutineResult{
		SetRoutine: s.next,
	}
}

func (s SellExcessInventory) Name() string {
	return "Sell Excess Inventory"
}
