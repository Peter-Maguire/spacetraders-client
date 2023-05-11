package main

import (
	"fmt"
	"spacetraders/entity"
	"spacetraders/http"
	"spacetraders/routine"
	"time"
)

func main() {
	http.Init()

	agent, err := http.Request[entity.Agent]("GET", "my/agent", nil)
	if err != nil {
		fmt.Println("error", err)
		return
	}
	contracts, _ := agent.Contracts()

	ships, _ := agent.Ships()

	fmt.Println("Starting Routines")
	for _, ship := range *ships {
		shipPtr := ship
		state := routine.State{
			Agent:    agent,
			Contract: &(*contracts)[0],
			Ship:     &shipPtr,
		}

		go routineLoop(&state)
	}

	forever := make(chan bool)
	<-forever
}

func routineLoop(state *routine.State) {
	currentRoutine := routine.GoToAsteroidField
	for {
		routineResult := currentRoutine(state)
		if routineResult.WaitSeconds > 0 {
			state.Log(fmt.Sprintf("Waiting for %d seconds", routineResult.WaitSeconds))
			time.Sleep(time.Duration(routineResult.WaitSeconds) * time.Second)
		}

		if routineResult.SetRoutine != nil {
			state.Log("Switching Routine")
			currentRoutine = routineResult.SetRoutine
		}
	}
}
