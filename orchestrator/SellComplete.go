package orchestrator

import (
    "fmt"
    "spacetraders/entity"
    "spacetraders/routine"
    "spacetraders/ui"
)

func (o *Orchestrator) onSellComplete(agent *entity.Agent) {
    ui.MainLog(fmt.Sprintf("Credits now: %d\n", agent.Credits))

    // We have 35 or more ships, we're at the limit of how many ships are useful
    if len(o.States) >= 35 {
        return
    }

    if agent.Credits >= o.CreditTarget && o.Shipyard != "" && o.ShipToBuy != "" {
        result, err := agent.BuyShip(o.Shipyard, o.ShipToBuy)
        if err == nil && result != nil {
            state := routine.State{
                Contract: o.Contract,
                Ship:     result.Ship,
                EventBus: o.Channel,
            }
            ui.MainLog(fmt.Sprintln("New ship", result.Ship.Symbol))
            o.States = append(o.States, &state)
            go o.routineLoop(&state)
        } else {
            if err.Data != nil && err.Data["creditsNeeded"] != nil {
                o.CreditTarget = int(err.Data["creditsNeeded"].(float64))
                ui.MainLog(fmt.Sprintln("Need ", o.CreditTarget))
            }
            ui.MainLog(fmt.Sprintln("Purchase error", err))
        }
    }

}
