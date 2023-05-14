package main

import (
	"fmt"
	"os"
	"spacetraders/database"
	"spacetraders/entity"
	"spacetraders/http"
	"spacetraders/routine"
	"spacetraders/ui"
	"time"
)

var states []*routine.State

var enableUi = false

func main() {
	enableUi = os.Getenv("DISABLE_UI") != "1"
	if enableUi {
		go ui.Init()
	}
	http.Init()
	database.Init()

	agent, err := http.Request[entity.Agent]("GET", "my/agent", nil)
	if err != nil {
		ui.MainLog(fmt.Sprint("error", err))
		return
	}
	contracts, _ := agent.Contracts()

	ships, _ := agent.Ships()

	orchestorChan := make(chan routine.OrchestratorEvent)

	states = make([]*routine.State, len(*ships))

	ui.MainLog(fmt.Sprint("Starting Routines"))
	for i, ship := range *ships {
		shipPtr := ship
		state := routine.State{
			Agent:    agent,
			Contract: &(*contracts)[0],
			Ship:     &shipPtr,
			EventBus: orchestorChan,
		}
		states[i] = &state

		go routineLoop(&state)
	}

	if enableUi {
		updateShipStates()
	}

	waypoints, _ := agent.Headquarters.GetSystemWaypoints()

	var shipyard entity.Waypoint

	for _, waypoint := range *waypoints {
		if waypoint.HasTrait("SHIPYARD") {
			ui.MainLog(fmt.Sprintln("Found shipyard at ", waypoint.Symbol))
			shipyard = waypoint.Symbol
			break
		}
	}

	var shipToBuy = "SHIP_ORE_HOUND"

	creditTarget := 185702

	shipyardStock, err := shipyard.GetShipyard()
	if err == nil {
		go database.StoreShipCosts(shipyardStock)
		for _, stock := range shipyardStock.Ships {
			if stock.Name == shipToBuy {
				ui.MainLog(fmt.Sprintf("Ship %s is available to buy at %s for %d credits\n", shipToBuy, shipyard, stock.PurchasePrice))
				creditTarget = stock.PurchasePrice
			}
		}
	}

	for {
		event := <-orchestorChan
		switch event.Name {
		case "sellComplete":
			agent := event.Data.(*entity.Agent)
			if agent.Credits >= creditTarget && shipyard != "" {
				result, err := agent.BuyShip(shipyard, shipToBuy)
				if err == nil && result != nil {
					//ui.MainLog(fmt.Sprintln(result))
					state := routine.State{
						Contract: &(*contracts)[0],
						Ship:     result.Ship,
						EventBus: orchestorChan,
					}
					ui.MainLog(fmt.Sprintln("New ship", result.Ship.Symbol))
					states = append(states, &state)
					go routineLoop(&state)
				} else {
					if err.Data != nil && err.Data["creditsNeeded"] != nil {
						creditTarget = int(err.Data["creditsNeeded"].(float64))
						ui.MainLog(fmt.Sprintln("Need ", creditTarget))
					}
					ui.MainLog(fmt.Sprintln("Purchase error", err))
				}
			}
			ui.MainLog(fmt.Sprintln("Credits now: ", agent.Credits))
		case "goodSurveyFound":
			ui.MainLog("Someone found a good survey\n")
			for _, state := range states {
				if state.Ship.IsMiningShip() && state.Survey == nil {
					state.Survey = event.Data.(*entity.Survey)
				}
			}
		case "contractComplete":
			ui.MainLog("Contract completed")
			contracts, err := agent.Contracts()
			if err != nil {
				ui.MainLog("Contract get error " + err.Error())
				os.Exit(1)
			}
			for _, c := range *contracts {
				if c.Accepted == false {
					err = c.Accept()
					if err != nil {
						ui.MainLog("Contract accept error " + err.Error())
						os.Exit(1)
					} else {
						ui.MainLog("Accepted new contract " + c.Id)
						for _, state := range states {
							state.Contract = &c
						}
					}
				}
			}
		}
	}

}

func updateShipStates() {
	ticker := time.NewTicker(500 * time.Millisecond)
	go func() {
		for {
			<-ticker.C
			output := fmt.Sprintf("Request Backlog: %d", http.Waiting)
			if http.IsRunningRequests {
				output += " (Active)\n"
			} else {
				output += "\n"
			}
			for _, state := range states {
				if state != nil && state.Ship != nil {
					output += state.Ship.Symbol
					if state.CurrentRoutine != nil {
						output += fmt.Sprintf(" (%s)", state.CurrentRoutine.Name())
					} else {
						output += " Stopped"
					}
					if state.AsleepUntil != nil {
						output += fmt.Sprintf(" Sleeping for %.f seconds", state.AsleepUntil.Sub(time.Now()).Seconds())
					}
					output += "\n"
				}
			}
			ui.WriteShipState(output)
		}
	}()
}

func routineLoop(state *routine.State) {
	state.CurrentRoutine = routine.DetermineObjective{}
	for {
		routineResult := state.CurrentRoutine.Run(state)

		if routineResult.WaitSeconds > 0 {
			state.Log(fmt.Sprintf("Waiting for %d seconds", routineResult.WaitSeconds))
			sleepTime := time.Duration(routineResult.WaitSeconds) * time.Second
			asleepUntil := time.Now().Add(sleepTime)
			state.AsleepUntil = &asleepUntil
			time.Sleep(sleepTime)
		}

		if routineResult.WaitUntil != nil {
			waitTime := routineResult.WaitUntil.Sub(time.Now())
			state.Log(fmt.Sprintf("Waiting until %s (%.f seconds)", routineResult.WaitUntil, waitTime.Seconds()))
			state.AsleepUntil = routineResult.WaitUntil
			time.Sleep(waitTime)
		}
		state.AsleepUntil = nil

		if state.ForceRoutine != nil {
			state.Log("Forced routine change")
			state.CurrentRoutine = state.ForceRoutine
			state.ForceRoutine = nil
			continue
		}

		if routineResult.SetRoutine != nil {
			state.Log(fmt.Sprintf("%s -> %s", state.CurrentRoutine.Name(), routineResult.SetRoutine.Name()))
			state.CurrentRoutine = routineResult.SetRoutine
		}

		if routineResult.Stop {
			state.CurrentRoutine = nil
			state.Log("Stopping Routine")
			break
		}
	}
}
