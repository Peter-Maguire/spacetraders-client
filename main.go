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

	for {
		event := <-orchestorChan
		switch event.Name {
		case "sellComplete":
			agent := event.Data.(*entity.Agent)
			if agent.Credits >= 87720 {
				result, err := agent.BuyShip("X1-ZA40-68707C", "SHIP_MINING_DRONE")
				if err != nil {
					state := routine.State{
						Contract: &(*contracts)[0],
						Ship:     result.Ship,
						EventBus: orchestorChan,
					}
					fmt.Println("New ship", result.Ship.Symbol)
					states = append(states, &state)
					go routineLoop(&state)
				} else {
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
	currentRoutine := routine.DetermineObjective
	for {
		routineResult := currentRoutine(state)
		if routineResult.WaitSeconds > 0 {
			state.Log(fmt.Sprintf("Waiting for %d seconds", routineResult.WaitSeconds))
			time.Sleep(time.Duration(routineResult.WaitSeconds) * time.Second)
		}

		if state.ForceRoutine != nil {
			state.Log("Forced routine change")
			currentRoutine = state.ForceRoutine
			state.ForceRoutine = nil
			continue
		}

		if routineResult.SetRoutine != nil {
			state.Log("Switching Routine")
			currentRoutine = routineResult.SetRoutine
		}

		if routineResult.Stop {
			state.Log("Stopping Routine")
			break
		}
	}
}
