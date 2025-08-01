package routine

import (
	"fmt"
	"sort"
	"spacetraders/database"
	"spacetraders/entity"
)

type ProcureConstructionSiteItem struct {
	next Routine
}

func (p ProcureConstructionSiteItem) Run(state *State) RoutineResult {

	if state.ConstructionSite == nil {
		return RoutineResult{
			Stop:       true,
			StopReason: "No construction site found",
		}
	}

	if state.ConstructionSite.IsComplete {
		state.Log("Construction site is already complete")
		for _, state := range *state.States {
			state.ConstructionSite = nil
		}
		return RoutineResult{
			SetRoutine: p.next,
		}
	}

	state.ConstructionSite.Update(state.Context)

	requiredMaterials := make([]entity.ConstructionMaterial, 0)
	materialStrings := make([]string, 0)
	for _, material := range state.ConstructionSite.Materials {
		if material.IsComplete() {
			continue
		}

		requiredMaterials = append(requiredMaterials, material)
		materialStrings = append(materialStrings, material.TradeSymbol)
	}

	targetMarkets := database.GetMarketsSellingInSystem(materialStrings, string(state.Ship.Nav.SystemSymbol))

	if len(targetMarkets) == 0 {
		return RoutineResult{
			Stop:       true,
			StopReason: "Unable to find any market selling any materials required for jump gate",
		}
	}

	currentWaypoint := database.GetWaypoint(state.Ship.Nav.WaypointSymbol).GetData()

	sort.Slice(targetMarkets, func(i, j int) bool {
		marketI := targetMarkets[i]
		marketIWaypoint := marketI.GetLimitedWaypointData()
		marketJ := targetMarkets[j]
		marketJWaypoint := marketJ.GetLimitedWaypointData()
		return marketIWaypoint.GetDistanceFrom(currentWaypoint.LimitedWaypointData)+marketI.BuyCost < marketJWaypoint.GetDistanceFrom(currentWaypoint.LimitedWaypointData)+marketJ.BuyCost
	})

	if targetMarkets[0].Waypoint != state.Ship.Nav.WaypointSymbol {
		return RoutineResult{SetRoutine: NavigateTo{next: p, waypoint: targetMarkets[0].Waypoint}}
	}

	state.Ship.EnsureNavState(state.Context, entity.NavDocked)

	system, _ := state.Ship.Nav.WaypointSymbol.GetSystem(state.Context)
	waypointData, _ := state.Ship.Nav.WaypointSymbol.GetWaypointData(state.Context)
	marketData, _ := state.Ship.Nav.WaypointSymbol.GetMarket(state.Context)
	database.StoreMarketRates(system, waypointData, marketData.TradeGoods)
	database.StoreMarketExchange(system, waypointData, "export", marketData.Exports)
	database.StoreMarketExchange(system, waypointData, "import", marketData.Imports)
	database.StoreMarketExchange(system, waypointData, "exchange", marketData.Exchange)

	for _, material := range materialStrings {
		tradeGood := marketData.GetTradeGood(material)
		if tradeGood == nil {
			continue
		}
		amountCanBuy := state.Agent.Credits / tradeGood.PurchasePrice
		constructionSite := state.ConstructionSite.GetMaterial(material)
		amountNeeded := constructionSite.Required - constructionSite.Fulfilled
		amountCanFit := state.Ship.Cargo.GetRemainingCapacity()
		buyAmount := min(amountCanFit, amountCanBuy, tradeGood.TradeVolume, amountNeeded)
		if buyAmount == 0 {
			continue
		}
		state.Log(fmt.Sprintf("Buying %dx %s", buyAmount, material))
		_, err := state.Ship.Purchase(state.Context, material, buyAmount)
		if err != nil {
			state.Log(fmt.Sprintf("Error purchasing item %s: %s", material, err.Error()))
		}
	}

	state.Ship.GetCargo(state.Context)

	if state.Ship.Cargo.IsFull() {
		state.Log("Delivering as our cargo is full")
		return RoutineResult{
			SetRoutine: DeliverConstructionSiteItem{next: p},
		}
	}

	for _, material := range materialStrings {
		slot := state.Ship.Cargo.GetSlotWithItem(material)
		if slot == nil {
			continue
		}
		constructionMaterial := state.ConstructionSite.GetMaterial(material)
		if constructionMaterial == nil {
			continue
		}
		if slot.Units >= constructionMaterial.GetRemaining() {
			state.Log(fmt.Sprintf("Delivering as %s is fulfilled", material))
			return RoutineResult{
				SetRoutine: DeliverConstructionSiteItem{next: p},
			}
		}
	}

	state.Log("Waiting to buy some more")
	return RoutineResult{
		WaitSeconds: 10,
	}

}

func (p ProcureConstructionSiteItem) Name() string {
	return fmt.Sprintf("Procure Construction Materials -> %s", p.next.Name())
}
