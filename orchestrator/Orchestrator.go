package orchestrator

import (
	"context"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"os"
	"spacetraders/database"
	"spacetraders/entity"
	"spacetraders/metrics"
	"spacetraders/routine"
	"spacetraders/ui"
	"strings"
	"sync"
	"time"
)

type Orchestrator struct {
	States       []*routine.State
	StatesMutex  sync.Mutex
	Agent        *entity.Agent
	Contract     *entity.Contract
	Haulers      []*entity.Ship
	Channel      chan routine.OrchestratorEvent
	CreditTarget int
	ShipToBuy    string
	Shipyard     entity.Waypoint
	Context      context.Context
}

var (
	agentShips = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "st_agent_ships",
		Help: "Current Ship Count",
	}, []string{"role", "system", "waypoint"})

	shipStates = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "st_ship_state",
		Help: "Ship States",
	}, []string{"name", "state"})
)

func Init(token string) *Orchestrator {
	//TODO:
	// There should be an overall goal for stages of development which determines the actions of each type
	// In no particular order, these are the various stages I see
	// 1. Construct jump gate (Mining, fulfilling contracts and purchasing materials to build the jump gate)
	// 2. Expand fleet (Mining and fulfilling contracts in order to have enough ships to make money)
	// 3. Begin charting (Start charting uncharted waypoints in other systems)
	// 4. Begin trading  (Replace mining fleet with trading fleet)
	// ALSO:
	// - It would be great to be able to set specific parameters (e.g how many ships of each type to buy, wait times, etc)
	//   between different agents, and compare the outcomes for each agent per reset

	ctx := context.WithValue(context.Background(), "token", token)

	agent, err := entity.GetAgent(ctx)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	hqSystem := agent.Headquarters.GetSystemName()
	if database.GetSystem(string(hqSystem)) == nil {
		ui.MainLog("We haven't stored our main system yet")
		systemData, err := entity.GetSystem(ctx, hqSystem)
		if err != nil {
			fmt.Println("failed to store system", err)
			os.Exit(1)
		}
		fmt.Println(systemData)
		database.StoreSystem(systemData)

		waypointData, _ := systemData.GetWaypoints(ctx)
		database.LogWaypoints(waypointData)
	}

	metrics.NumCredits.Set(float64(agent.Credits))

	contracts, _ := agent.Contracts(ctx)

	var contract *entity.Contract

	for _, c := range *contracts {
		if !c.Fulfilled && c.Accepted {
			contract = &c
			for _, term := range contract.Terms.Deliver {
				ui.MainLog(fmt.Sprintf("We are delivering %dx %s to %s for %d credits", term.UnitsRequired, term.TradeSymbol, term.DestinationSymbol, contract.Terms.Payment.GetTotalPayment()))
			}
			break
		}
	}

	if contract == nil {
		for _, c := range *contracts {
			if !c.Fulfilled {
				ui.MainLog("Accepted contract")
				err = c.Accept(ctx)
				if err == nil {
					contract = &c
				} else {
					ui.MainLog(err.Error())
				}
				break
			}
		}
	}

	if contract == nil {
		ui.MainLog("No current Contract")
	} else {
		metrics.ContractProgress.Set(float64(contract.Terms.Deliver[0].UnitsFulfilled))
		metrics.ContractRequirement.Set(float64(contract.Terms.Deliver[0].UnitsRequired))
	}

	orc := Orchestrator{
		Agent:        agent,
		Contract:     contract,
		Channel:      make(chan routine.OrchestratorEvent),
		CreditTarget: 80000,
		Context:      ctx,
	}

	go orc.start()

	//orc.StatesMutex.Unlock()
	return &orc
}

func (o *Orchestrator) start() {
	ships, err2 := o.Agent.Ships(o.Context)
	if err2 != nil {
		fmt.Println(err2)
	}
	shipCount := len(*ships)
	ui.MainLog(fmt.Sprintf("We have %d ships:", shipCount))
	for _, ship := range *ships {
		agentShips.WithLabelValues(string(ship.Registration.Role), string(ship.Nav.SystemSymbol), string(ship.Nav.WaypointSymbol)).Add(1)
		ui.MainLog(fmt.Sprintf("%s: %s type", ship.Registration.Name, ship.Registration.Role))
		if ship.Registration.Role == "HAULER" {
			ui.MainLog(fmt.Sprintf("%s is HAULER", ship.Registration))
			shipCopy := ship
			o.Haulers = append(o.Haulers, &shipCopy)
		}
	}

	go o.runEvents()

	o.States = make([]*routine.State, len(*ships))
	shipFilter := os.Getenv("SHIP_FILTER")
	ui.MainLog(fmt.Sprint("Starting Routines"))
	//orc.StatesMutex.Lock()
	for i, ship := range *ships {
		if shipFilter != "" && !strings.Contains(shipFilter, ship.Symbol) {
			ui.MainLog(fmt.Sprintf("Skipping %s because it's not in the ship filter", ship.Symbol))
			continue
		}
		shipPtr := ship
		state := routine.State{
			Agent:       o.Agent,
			Contract:    o.Contract,
			Ship:        &shipPtr,
			States:      &o.States,
			StatesMutex: &o.StatesMutex,
			Haulers:     o.Haulers,
			EventBus:    o.Channel,
		}
		state.Context = context.WithValue(o.Context, "state", &state)
		o.States[i] = &state

		go o.routineLoop(&state)
	}
}

func (o *Orchestrator) startNewContract() {
	contracts, _ := o.Agent.Contracts(o.Context)
	for _, c := range *contracts {
		if !c.Fulfilled {
			ui.MainLog("Accepted contract")
			err := c.Accept(o.Context)
			if err == nil {
				o.Contract = &c
			} else {
				ui.MainLog(err.Error())
			}
			break
		}
	}
}

func (o *Orchestrator) runEvents() {
	for {
		event := <-o.Channel
		switch event.Name {
		case "newShip":
			ship := event.Data.(*entity.Ship)
			newState := routine.State{
				Agent:    o.Agent,
				Contract: o.Contract,
				Ship:     ship,
				Haulers:  o.Haulers,
				EventBus: o.Channel,
				States:   &o.States,
			}
			newState.Context = context.WithValue(o.Context, "state", &newState)
			ui.MainLog(fmt.Sprintln("New ship", ship.Symbol))
			o.States = append(o.States, &newState)
			go o.routineLoop(&newState)

		case "goodSurveyFound":
			ui.MainLog("Someone found a good survey")
			//o.StatesMutex.Lock()
			for _, state := range o.States {
				if state.Ship.IsMiningShip() && state.Survey == nil {
					state.Survey = event.Data.(*entity.Survey)
				}
			}
			//o.StatesMutex.Unlock()
		case "surveyExhausted":
			ui.MainLog("Survey bad")
			//o.StatesMutex.Lock()
			for _, state := range o.States {
				if state.Survey == event.Data.(*entity.Survey) {
					state.Survey = nil
				}
			}
			//o.StatesMutex.Unlock()
		case "contractComplete":
			ui.MainLog("Contract completed")
			o.Contract = nil
			for _, state := range o.States {
				state.Contract = nil
				//state.ForceRoutine = routine.DetermineObjective{}
			}
		case "newContract":
			ui.MainLog("New contract")
			contract := event.Data.(*entity.Contract)
			o.Contract = contract
			for _, state := range o.States {
				state.Contract = contract
				//state.ForceRoutine = routine.DetermineObjective{}
			}
		}
	}
}

func (o *Orchestrator) routineLoop(state *routine.State) {
	state.CurrentRoutine = routine.DetermineObjective{}
	for {
		routineName := state.CurrentRoutine.Name()
		if len(routineName) > 500 {
			state.StoppedReason = "Loop Detected - " + routineName
			state.Log("Stopping Routine")
			break
		}
		shipStates.WithLabelValues(state.Ship.Symbol, routineName).Set(1)
		routineResult := state.CurrentRoutine.Run(state)
		if routineResult.WaitSeconds > 0 {
			//state.Log(fmt.Sprintf("Waiting for %d seconds", routineResult.WaitSeconds))
			sleepTime := time.Duration(routineResult.WaitSeconds) * time.Second
			asleepUntil := time.Now().Add(sleepTime)
			state.AsleepUntil = &asleepUntil
			time.Sleep(sleepTime)
		}

		if routineResult.WaitUntil != nil {
			waitTime := routineResult.WaitUntil.Sub(time.Now())
			// TODO: solve the clock issue instead of this
			waitTime += 5
			//state.Log(fmt.Sprintf("Waiting until %s (%.f seconds)", routineResult.WaitUntil, waitTime.Seconds()))
			state.AsleepUntil = routineResult.WaitUntil
			time.Sleep(waitTime)
		}
		state.AsleepUntil = nil

		if state.ForceRoutine != nil {
			state.Log("Forced routine change")
			shipStates.WithLabelValues(state.Ship.Symbol, state.CurrentRoutine.Name()).Set(0)
			state.CurrentRoutine = state.ForceRoutine
			state.ForceRoutine = nil
			continue
		}

		if routineResult.SetRoutine != nil {
			shipStates.WithLabelValues(state.Ship.Symbol, state.CurrentRoutine.Name()).Set(0)
			state.Log(fmt.Sprintf("%s => %s", state.CurrentRoutine.Name(), routineResult.SetRoutine.Name()))
			state.CurrentRoutine = routineResult.SetRoutine
		}

		if routineResult.Stop {
			shipStates.WithLabelValues(state.Ship.Symbol, state.CurrentRoutine.Name()).Set(0)
			state.CurrentRoutine = nil
			state.StoppedReason = routineResult.StopReason
			state.Log("Stopping Routine")
			break
		}
	}
	state.CurrentRoutine = nil
	state.Log("!!!! Loop exited!")
}

func (o *Orchestrator) GetAgent() *entity.Agent {
	return o.Agent
}

func (o *Orchestrator) GetContract() *entity.Contract {
	return o.Contract
}
