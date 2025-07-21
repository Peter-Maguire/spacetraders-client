package orchestrator

//
//func (o *Orchestrator) onSellComplete(agent *entity.Agent) {
//	ui.MainLog(fmt.Sprintf("Credits now: %d", agent.Credits))
//
//	metrics.NumCredits.Set(float64(agent.Credits))
//
//	// We have 35 or more ships, we're at the limit of how many ships are useful
//	if len(o.States) >= 35 {
//		return
//	}
//
//	if agent.Credits < o.CreditTarget {
//		ui.MainLog("Not enough credits to buy a new ship yet")
//		return
//	}
//
//	for _, state := range o.States {
//		if state.Ship.Registration.Role == "SATELLITE" {
//			ui.MainLog(fmt.Sprintf("Sending %s to buy a %s", state.Ship.Symbol, o.ShipToBuy))
//			state.ForceRoutine = routine.Satellite{
//				ShipToBuy: o.ShipToBuy,
//			}
//			break
//		}
//	}
//
//	if agent.Credits >= o.CreditTarget && o.Shipyard != "" && o.ShipToBuy != "" {
//		// TODO: figure this shit out
//		result, err := agent.BuyShip(o.Context, o.Shipyard, o.ShipToBuy)
//		if err == nil && result != nil {
//			state := routine.State{
//				Agent:    agent,
//				Contract: o.Contract,
//				Ship:     result.Ship,
//				Haulers:  o.Haulers,
//				EventBus: o.Channel,
//				States:   &o.States,
//			}
//			ui.MainLog(fmt.Sprintln("New ship", result.Ship.Symbol))
//			o.States = append(o.States, &state)
//			go o.routineLoop(&state)
//		} else {
//			if err.Data != nil && err.Data["creditsNeeded"] != nil {
//				o.CreditTarget = int(err.Data["creditsNeeded"].(float64))
//				ui.MainLog(fmt.Sprintln("Need ", o.CreditTarget))
//			}
//			ui.MainLog(fmt.Sprintln("Purchase error", err))
//		}
//	}
//
//}
