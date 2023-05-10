package routine

import (
	"fmt"
	"spacetraders/entity"
)

func SellExcessInventory(state *entity.State, ship *entity.Ship) RoutineResult {
	inventory := ship.Cargo.Inventory

	market, _ := ship.Nav.WaypointSymbol.GetMarket()

	contractTarget := state.Contract.Terms.Deliver[0]
	targetItem := contractTarget.TradeSymbol

	fmt.Println("We are delivering ", contractTarget)

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
		if ship.Cargo.GetSlotWithItem(targetItem) != nil {
			fmt.Println("All we have left is what we are selling, time to take it away")

			return RoutineResult{
				SetRoutine: NavigateTo(contractTarget.DestinationSymbol, DeliverContractItem(targetItem, ship.Nav.WaypointSymbol)),
			}

		}
	}

	fmt.Printf("Got %d items to sell\n", len(sellable))

	// dock ship
	_ = ship.EnsureNavState(entity.NavDocked)

	for _, sellableSlot := range sellable {
		fmt.Printf("Selling %dx %s\n", sellableSlot.Units, sellableSlot.Symbol)
		sellResult, err := ship.SellCargo(sellableSlot.Symbol, sellableSlot.Units)
		if err != nil {
			fmt.Println("Failed to sell:", err.Data)
		}

		state.Agent = &sellResult.Agent
	}

	// Turns out you can't do this
	//if state.Agent.Credits > 20000 {
	//    fmt.Println("Using excess money to buy target items")
	//    return RoutineResult{
	//        SetRoutine: BuyTargetItem,
	//    }
	//}

	return RoutineResult{
		SetRoutine: MineOres,
	}
}
