package routine

import (
    "fmt"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
    "spacetraders/database"
    "spacetraders/entity"
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

func (s SellExcessInventory) Run(state *State) RoutineResult {
    inventory := state.Ship.Cargo.Inventory

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
    sellableNames := make([]string, 0)

    for _, slot := range inventory {
        // Don't sell antimatter or contract target
        if slot.Symbol == "ANTIMATTER" || slot.Symbol == targetItem {
            continue
        }
        sellable = append(sellable, slot)
        sellableNames = append(sellableNames, slot.Name)
    }

    if len(sellable) == 0 {
        if contractTarget != nil && state.Ship.Cargo.GetSlotWithItem(targetItem) != nil {
            state.Log("All we have left is what we are selling, time to take it away")

            return RoutineResult{
                SetRoutine: NavigateTo{waypoint: contractTarget.DestinationSymbol, next: DeliverContractItem{item: targetItem, returnTo: state.Ship.Nav.WaypointSymbol}},
            }
        }
        if state.Ship.Cargo.IsFull() {
            state.Log("Full of something that you can't sell here!")
            return RoutineResult{
                Stop: true,
            }
        }
    }

    markets := database.GetMarketsSelling(sellableNames)

    if len(markets) == 0 {
        state.Log("Could not sell, no markets were found that are selling what we need")
        return RoutineResult{
            Stop: true,
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
