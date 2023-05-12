package routine

import (
	"fmt"
	"spacetraders/database"
	"spacetraders/entity"
)

func SellExcessInventory(state *State) RoutineResult {
	inventory := state.Ship.Cargo.Inventory

	market, _ := state.Ship.Nav.WaypointSymbol.GetMarket()

	go database.StoreMarketRates(string(state.Ship.Nav.WaypointSymbol), market.TradeGoods)

	contractTarget := state.Contract.Terms.Deliver[0]
	targetItem := contractTarget.TradeSymbol

	state.Log("We are delivering " + contractTarget.TradeSymbol)

	// Sellable = not antimatter, not required for the contract and sellable at this market
	sellable := make([]entity.ShipInventorySlot, 0)

	for _, slot := range inventory {
		// Don't sell antimatter or contract target
		if slot.Symbol == "ANTIMATTER" || slot.Symbol == targetItem {
			continue
		}
		tradeGood := market.GetTradeGood(slot.Symbol)
		if tradeGood != nil {
			fmt.Printf("We can trade our %s here for %d credits\n", tradeGood.Symbol, tradeGood.SellPrice)
			sellable = append(sellable, slot)
		}
	}

	if len(sellable) == 0 {
		if state.Ship.Cargo.GetSlotWithItem(targetItem) != nil {
			state.Log("All we have left is what we are selling, time to take it away")

			return RoutineResult{
				SetRoutine: NavigateTo(contractTarget.DestinationSymbol, DeliverContractItem(targetItem, state.Ship.Nav.WaypointSymbol)),
			}

		}
	}

	fmt.Printf("Got %d items to sell\n", len(sellable))

	// dock ship
	_ = state.Ship.EnsureNavState(entity.NavDocked)

	for _, sellableSlot := range sellable {
		state.Log(fmt.Sprintf("Selling %dx %s", sellableSlot.Units, sellableSlot.Symbol))
		sellResult, err := state.Ship.SellCargo(sellableSlot.Symbol, sellableSlot.Units)
		if err != nil {
			state.Log("Failed to sell:" + err.Error())
		} else {
			state.Agent = &sellResult.Agent
		}

	}

	state.FireEvent("sellComplete", state.Agent)

	return RoutineResult{
		SetRoutine: GetSurvey,
	}
}
