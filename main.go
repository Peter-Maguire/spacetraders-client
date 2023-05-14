package main

import (
	"fmt"
	"spacetraders/database"
	"spacetraders/entity"
	"spacetraders/http"
	"spacetraders/routine"
	"time"
)

func main() {
	database.Init()
	http.Init()

	agent, err := http.Request[entity.Agent]("GET", "my/agent", nil)
	if err != nil {
		fmt.Println("error", err)
		return
	}
	contracts, _ := agent.Contracts()

	ships, _ := agent.Ships()

	orchestorChan := make(chan routine.OrchestratorEvent)

	states := make([]*routine.State, len(*ships))

	fmt.Println("Starting Routines")
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

	waypoints, _ := agent.Headquarters.GetSystemWaypoints()

	var shipyard entity.Waypoint

	for _, waypoint := range *waypoints {
		if waypoint.HasTrait("SHIPYARD") {
			fmt.Println("Found shipyard at ", waypoint.Symbol)
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
				fmt.Printf("Ship %s is available to buy at %s for %d credits\n", shipToBuy, shipyard, stock.PurchasePrice)
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
					fmt.Println(result)
					state := routine.State{
						Contract: &(*contracts)[0],
						Ship:     result.Ship,
						EventBus: orchestorChan,
					}
					fmt.Println("New ship", result.Ship.Symbol)
					states = append(states, &state)
					go routineLoop(&state)
				} else {
					if err.Data != nil && err.Data["creditsNeeded"] != nil {
						creditTarget = err.Data["creditsNeeded"].(int)
						fmt.Println("Need ", creditTarget)
					}
					fmt.Println("Purchase error", err)
				}
			}
			fmt.Println("Credits now: ", agent.Credits)
		case "goodSurveyFound":
			fmt.Println("Someone found a good survey")
			for _, state := range states {
				if state.Ship.IsMiningShip() && state.Survey == nil {
					state.Survey = event.Data.(*entity.Survey)
				}
			}
		}
	}

}

func routineLoop(state *routine.State) {
	var currentRoutine routine.Routine = routine.DetermineObjective{}
	for {
		routineResult := currentRoutine.Run(state)
		if routineResult.WaitSeconds > 0 {
			state.Log(fmt.Sprintf("Waiting for %d seconds", routineResult.WaitSeconds))
			time.Sleep(time.Duration(routineResult.WaitSeconds) * time.Second)
		}

		if routineResult.WaitUntil != nil {
			waitTime := routineResult.WaitUntil.Sub(time.Now())
			state.Log(fmt.Sprintf("Waiting until %s (%.f seconds)", routineResult.WaitUntil, waitTime.Seconds()))
			time.Sleep(waitTime)
		}

		if state.ForceRoutine != nil {
			state.Log("Forced routine change")
			currentRoutine = state.ForceRoutine
			state.ForceRoutine = nil
			continue
		}

		if routineResult.SetRoutine != nil {
			state.Log(fmt.Sprintf("%s -> %s", currentRoutine.Name(), routineResult.SetRoutine.Name()))
			currentRoutine = routineResult.SetRoutine
		}

		if routineResult.Stop {
			state.Log("Stopping Routine")
			break
		}
	}
}
