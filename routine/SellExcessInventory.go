package routine

import (
    "fmt"
    "math"
    "spacetraders/entity"
)

type SellExcessInventory struct {
    next Routine
}

func (s SellExcessInventory) Run(state *State) RoutineResult {
    inventory := state.Ship.Cargo.Inventory

    minInventoryAmount := 1000

    for _, slot := range inventory {
        if slot.Symbol != "ANTIMATTER" && slot.Units < minInventoryAmount {
            minInventoryAmount = slot.Units
        }
    }

    // TransferToHauler ignores items less than 10 in the inventory
    minInventoryAmount = int(math.Max(float64(minInventoryAmount), 10))

    for _, hauler := range state.Haulers {
        if hauler.Nav.WaypointSymbol == state.Ship.Nav.WaypointSymbol && state.Ship.Symbol != hauler.Symbol && hauler.Cargo.Units+minInventoryAmount < hauler.Cargo.Capacity {
            return RoutineResult{SetRoutine: TransferToHauler{s.next, hauler}}
        }
    }

    state.WaitingForHttp = true
    market, err := state.Ship.Nav.WaypointSymbol.GetMarket()
    state.WaitingForHttp = false

    if err != nil {
        state.Log("Market error" + err.Error())
        return RoutineResult{WaitSeconds: 10}
    }

    //go database.StoreMarketRates(string(state.Ship.Nav.WaypointSymbol), market.TradeGoods)

    var contractTarget *entity.ContractDeliverable
    targetItem := ""
    if state.Contract != nil {
        contractTarget = &state.Contract.Terms.Deliver[0]
        targetItem = contractTarget.TradeSymbol
        state.Log("We are delivering " + contractTarget.TradeSymbol)
    }

    // Sellable = not antimatter, not required for the contract and sellable at this market
    sellable := make([]entity.ShipInventorySlot, 0)

    for _, slot := range inventory {
        // Don't sell antimatter or contract target
        if slot.Symbol == "ANTIMATTER" || slot.Symbol == targetItem {
            continue
        }
        tradeGood := market.GetTradeGood(slot.Symbol)
        if tradeGood != nil {
            state.Log(fmt.Sprintf("We can trade our %s here for %d credits", tradeGood.Symbol, tradeGood.SellPrice))
            sellable = append(sellable, slot)
        }
    }

    if len(sellable) == 0 && contractTarget != nil {
        if state.Ship.Cargo.GetSlotWithItem(targetItem) != nil {
            state.Log("All we have left is what we are selling, time to take it away")

            return RoutineResult{
                SetRoutine: NavigateTo{waypoint: contractTarget.DestinationSymbol, next: DeliverContractItem{item: targetItem, returnTo: state.Ship.Nav.WaypointSymbol}},
            }

        }
    }

    //fmt.Printf("Got %d items to sell\n", len(sellable))

    // dock ship
    state.WaitingForHttp = true
    _ = state.Ship.EnsureNavState(entity.NavDocked)
    state.WaitingForHttp = false

    for _, sellableSlot := range sellable {
        state.Log(fmt.Sprintf("Selling %dx %s", sellableSlot.Units, sellableSlot.Symbol))
        state.WaitingForHttp = true
        sellResult, err := state.Ship.SellCargo(sellableSlot.Symbol, sellableSlot.Units)
        state.WaitingForHttp = false
        if err != nil {
            state.Log("Failed to sell:" + err.Error())
        } else {
            state.Agent = &sellResult.Agent
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
