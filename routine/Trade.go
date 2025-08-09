package routine

import (
	"fmt"
	"sort"
	"spacetraders/constant"
	"spacetraders/database"
	"spacetraders/entity"
	"spacetraders/util"
)

type Trade struct {
}

type tradeProfitOpportunity struct {
	MarketFrom  database.MarketRates
	MarketTo    database.MarketRates
	BuyCost     int
	SellRevenue int
	TravelCost  int
	Profit      int
	Distance    int
}

func (t Trade) Run(state *State) RoutineResult {

	markets := database.GetMarkets()

	totalFuelCost := 0
	fuelMarkets := 0

	for _, market := range markets {
		if market.Good != "FUEL" || market.Waypoint.GetSystemName() != state.Ship.Nav.SystemSymbol {
			continue
		}
		totalFuelCost += market.BuyCost
		fuelMarkets++
	}

	currentDbWaypoint := database.GetWaypoint(state.Ship.Nav.WaypointSymbol)
	currentWaypointData := currentDbWaypoint.GetData()

	avgFuelCost := totalFuelCost / fuelMarkets

	profitOpportunities := make([]tradeProfitOpportunity, 0)

	for _, fromMarket := range markets {
		if fromMarket.Waypoint.GetSystemName() != state.Ship.Nav.SystemSymbol {
			continue
		}

		for _, toMarket := range markets {
			if toMarket.Waypoint == fromMarket.Waypoint || toMarket.Waypoint.GetSystemName() != state.Ship.Nav.SystemSymbol {
				continue
			}

			// Only markets buying what we're selling
			if toMarket.Good != fromMarket.Good {
				continue
			}

			if toMarket.SellCost <= fromMarket.BuyCost {
				continue
			}

			fromMarketWaypoint := database.GetWaypoint(fromMarket.Waypoint)
			toMarketWaypoint := database.GetWaypoint(toMarket.Waypoint)

			fromMarketWaypointData := fromMarketWaypoint.GetData()
			toMarketWaypointData := toMarketWaypoint.GetData()

			distance := fromMarketWaypointData.GetDistanceFrom(toMarketWaypointData.LimitedWaypointData) +
				currentWaypointData.GetDistanceFrom(fromMarketWaypointData.LimitedWaypointData)

			flightFuel := util.GetFuelCost(distance, state.Ship.Nav.FlightMode)
			fuelCost := (flightFuel / 100) * avgFuelCost

			op := tradeProfitOpportunity{
				MarketFrom:  fromMarket,
				MarketTo:    toMarket,
				BuyCost:     fromMarket.BuyCost * state.Ship.Cargo.Capacity,
				SellRevenue: toMarket.SellCost * state.Ship.Cargo.Capacity,
				TravelCost:  fuelCost,
				Distance:    distance,
			}

			if state.Ship.Cargo.GetSlotWithItem(fromMarket.Good) != nil {
				op.BuyCost = 0
			}

			op.Profit = op.SellRevenue - op.BuyCost - op.TravelCost
			if op.Profit < 0 {
				continue
			}

			profitOpportunities = append(profitOpportunities, op)
		}
	}

	sort.Slice(profitOpportunities, func(i, j int) bool {
		return profitOpportunities[i].Profit > profitOpportunities[j].Profit
	})

	opsPerWaypoint := make(map[entity.Waypoint]int)
	for _, op := range profitOpportunities {
		opsPerWaypoint[op.MarketFrom.Waypoint]++
	}

	var bestOpportunity *tradeProfitOpportunity

	for _, op := range profitOpportunities {
		if state.Ship.Cargo.GetSlotWithItem(op.MarketFrom.Good) != nil {
			bestOpportunity = &op
			break
		}
	}

	if bestOpportunity == nil {
		for _, op := range profitOpportunities {
			if op.MarketFrom.Waypoint == state.Ship.Nav.WaypointSymbol {
				bestOpportunity = &op
				break
			}
			if len(state.GetShipsWithRoleAtOrGoingToWaypoint(constant.ShipRoleTransport, op.MarketFrom.Waypoint)) > opsPerWaypoint[op.MarketFrom.Waypoint] {
				continue
			}
			bestOpportunity = &op
			break
		}
	}

	if bestOpportunity == nil {
		return RoutineResult{
			Stop:       true,
			StopReason: "Unable to find any good trade opportunities",
		}
	}

	state.Log(fmt.Sprintf("Best opportunity is trading %s at %s with profit %d", bestOpportunity.MarketFrom.Good, bestOpportunity.MarketFrom.Waypoint, bestOpportunity.Profit))

	slot := state.Ship.Cargo.GetSlotWithItem(bestOpportunity.MarketFrom.Good)

	if slot != nil {
		state.Log(fmt.Sprintf("Delivering some %s", bestOpportunity.MarketFrom.Good))
		if bestOpportunity.MarketTo.Waypoint != state.Ship.Nav.WaypointSymbol {
			return RoutineResult{
				SetRoutine: NavigateTo{
					waypoint: bestOpportunity.MarketTo.Waypoint,
					next:     t,
				},
			}
		}

		state.Ship.EnsureNavState(state.Context, entity.NavDocked)

		market, _ := state.Ship.Nav.WaypointSymbol.GetMarket(state.Context)
		database.UpdateMarketRates(state.Ship.Nav.WaypointSymbol, market.TradeGoods)

		sellSuccess := false
		for _, inv := range state.Ship.Cargo.Inventory {
			tg := market.GetTradeGood(inv.Symbol)
			if tg == nil {
				continue
			}

			numSells := max(inv.Units/tg.TradeVolume, 1)
			fmt.Println("numSells", numSells)
			for i := 0; i < numSells; i++ {
				updatedSlot := state.Ship.Cargo.GetSlotWithItem(inv.Symbol)
				sr, err := state.Ship.SellCargo(state.Context, inv.Symbol, min(updatedSlot.Units, tg.TradeVolume))
				if err != nil {
					fmt.Println(err)
					break
				}
				if sr != nil {
					sellSuccess = true
					state.Ship.Cargo.Inventory = sr.Cargo.Inventory
					state.Agent.Credits = sr.Agent.Credits
					state.Log(fmt.Sprintf("We now have %d credits", state.Agent.Credits))
				}

			}
		}

		if !sellSuccess {
			return RoutineResult{
				Stop:       true,
				StopReason: "No sell completed successfully",
			}
		}

		state.Log("Trade complete")

		market, _ = state.Ship.Nav.WaypointSymbol.GetMarket(state.Context)
		database.UpdateMarketRates(state.Ship.Nav.WaypointSymbol, market.TradeGoods)

		return RoutineResult{}
	}

	if bestOpportunity.MarketFrom.Waypoint != state.Ship.Nav.WaypointSymbol {
		return RoutineResult{
			SetRoutine: NavigateTo{
				waypoint: bestOpportunity.MarketFrom.Waypoint,
				next:     t,
			},
		}
	}

	market, _ := state.Ship.Nav.WaypointSymbol.GetMarket(state.Context)
	database.UpdateMarketRates(state.Ship.Nav.WaypointSymbol, market.TradeGoods)

	tg := market.GetTradeGood(bestOpportunity.MarketFrom.Good)
	if tg == nil {
		state.Log(fmt.Sprintf("This market no longer sells %s", bestOpportunity.MarketFrom.Good))
		return RoutineResult{}
	}

	buyAmount := min(state.Ship.Cargo.GetRemainingCapacity(), tg.TradeVolume, state.Agent.Credits/tg.PurchasePrice)

	if buyAmount <= 0 {
		state.Log("We can't currently buy anything...")
		if state.Ship.Cargo.IsFull() {
			return RoutineResult{
				SetRoutine: SellExcessInventory{next: t},
			}
		}

		return RoutineResult{
			WaitSeconds: 90,
		}
	}

	state.Log(fmt.Sprintf("Trying to buy %dx %s", buyAmount, bestOpportunity.MarketFrom.Good))
	state.Ship.EnsureNavState(state.Context, entity.NavDocked)

	numBuys := buyAmount / tg.TradeVolume

	successfulBuy := false
	for i := 0; i < numBuys; i++ {
		_, err := state.Ship.Purchase(state.Context, tg.Symbol, min(buyAmount, tg.TradeVolume))
		if err != nil {
			state.Log(err.Message)
			break
		} else {
			successfulBuy = true
		}
	}
	if !successfulBuy {
		return RoutineResult{
			WaitSeconds:       90,
		}
	}

	return RoutineResult{}
}

func (t Trade) Name() string {
	return "Trade"
}
