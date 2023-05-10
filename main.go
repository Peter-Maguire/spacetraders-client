package main

import (
	"fmt"
	"spacetraders/entity"
	"spacetraders/http"
	"spacetraders/routine"
	"time"
)

var currentRoutine routine.Routine

func main() {
	http.Init()

	agent, err := http.Request[entity.Agent]("GET", "my/agent", nil)
	if err != nil {
		fmt.Println("error", err)
		return
	}
	contracts, _ := agent.Contracts()

	state := entity.State{
		Agent:    agent,
		Contract: &(*contracts)[0],
	}

	ships, _ := agent.Ships()
	miningShip := (*ships)[0]

	currentRoutine = routine.GetSurvey

	fmt.Println("Starting Routine")
	for {
		routineResult := currentRoutine(&state, &miningShip)
		if routineResult.WaitSeconds > 0 {
			fmt.Printf("Waiting for %d seconds\n", routineResult.WaitSeconds)
			time.Sleep(time.Duration(routineResult.WaitSeconds) * time.Second)
		}

		if routineResult.SetRoutine != nil {
			fmt.Println("Switching Routine")
			currentRoutine = routineResult.SetRoutine
		}
	}
}
