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

	ctx := context.WithValue(context.Background(), "token", token)

	shipFilter := os.Getenv("SHIP_FILTER")

	agent, err := entity.GetAgent(ctx)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	hqSystem := agent.Headquarters.GetSystemName()
	if database.GetSystem(hqSystem) == nil {
		ui.MainLog("We haven't stored our main system yet")
		systemData, err := entity.GetSystem(ctx, hqSystem)
		if err != nil {
			fmt.Println("failed to store system", err)
			os.Exit(1)
		}
		fmt.Println(systemData)
		database.StoreSystem(systemData)
	}

	metrics.NumCredits.Set(float64(agent.Credits))

	contracts, _ := agent.Contracts(ctx)

	var contract *entity.Contract

	for _, c := range *contracts {
		if !c.Fulfilled && c.Accepted {
			contract = &c
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

	waypoints, _ := agent.Headquarters.GetSystemWaypoints(ctx)
	database.LogWaypoints(waypoints)

	// TODO: fix shipyard logic
	orc.Shipyard = ""

	for _, w := range *waypoints {
		if w.HasTrait("SHIPYARD") {
			shipyard, _ := w.Symbol.GetShipyard(ctx)
			if shipyard.SellsShipType(orc.ShipToBuy) {
				orc.Shipyard = w.Symbol
				break
			}
			// TODO: find the closest
		}
	}

	//for _, waypoint := range *waypoints {
	//	if waypoint.HasTrait("SHIPYARD") {
	//		ui.MainLog(fmt.Sprintf("Found Shipyard at %s", waypoint.Symbol))
	//		orc.Shipyard = waypoint.Symbol
	//		break
	//	}
	//}

	ships, err2 := orc.Agent.Ships(ctx)
	if err2 != nil {
		fmt.Println(err2)
	}
	shipCount := len(*ships)
	ui.MainLog(fmt.Sprintf("We have %d ships:", shipCount))
	for _, ship := range *ships {
		agentShips.WithLabelValues(ship.Registration.Role, ship.Nav.SystemSymbol, string(ship.Nav.WaypointSymbol)).Add(1)
		ui.MainLog(fmt.Sprintf("%s: %s type", ship.Registration.Name, ship.Registration.Role))
		if ship.Registration.Role == "HAULER" {
			ui.MainLog(fmt.Sprintf("%s is HAULER", ship.Registration))
			shipCopy := ship
			orc.Haulers = append(orc.Haulers, &shipCopy)
		}
	}

	// TODO: this logic should be more nuanced
	if shipCount < 30 {
		orc.ShipToBuy = "SHIP_MINING_DRONE"
	} else {
		orc.ShipToBuy = "SHIP_LIGHT_HAULER"
	}

	// TODO: This should be merged into the explore logic
	// TODO: Does this even work when we're not actually there?
	//shipyardStock, err := orc.Shipyard.GetShipyard(ctx)
	//if err == nil {
	//	ui.MainLog(fmt.Sprintf("Shipyard at %s has %d types, %d available", shipyardStock.Symbol, len(shipyardStock.ShipTypes), len(shipyardStock.Ships)))
	//	if len(shipyardStock.Ships) > 0 {
	//		// TODO: This should maybe store the available ship types here even if we don't know the price
	//		go database.StoreShipCosts(shipyardStock)
	//	}
	//	for _, stock := range shipyardStock.Ships {
	//		if stock.Name == orc.ShipToBuy {
	//			ui.MainLog(fmt.Sprintf("Ship %s is available to buy at %s for %d credits", orc.ShipToBuy, orc.Shipyard, stock.PurchasePrice))
	//			orc.CreditTarget = stock.PurchasePrice
	//		}
	//	}
	//}

	go orc.runEvents()

	orc.States = make([]*routine.State, len(*ships))

	ui.MainLog(fmt.Sprint("Starting Routines"))
	//orc.StatesMutex.Lock()
	for i, ship := range *ships {
		if shipFilter != "" && !strings.Contains(shipFilter, ship.Symbol) {
			ui.MainLog(fmt.Sprintf("Skipping %s because it's not in the ship filter", ship.Symbol))
			continue
		}
		shipPtr := ship
		state := routine.State{
			Agent:       agent,
			Contract:    contract,
			Ship:        &shipPtr,
			States:      &orc.States,
			StatesMutex: &orc.StatesMutex,
			Haulers:     orc.Haulers,
			EventBus:    orc.Channel,
		}
		state.Context = context.WithValue(ctx, "state", &state)
		orc.States[i] = &state

		go orc.routineLoop(&state)
	}
	//orc.StatesMutex.Unlock()
	return &orc
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
			for _, state := range o.States {
				state.Contract = nil
				//state.ForceRoutine = routine.DetermineObjective{}
			}
		case "newContract":
			ui.MainLog("New contract")
			contract := event.Data.(*entity.Contract)
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
		shipStates.WithLabelValues(state.Ship.Symbol, state.CurrentRoutine.Name()).Set(1)
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
			state.Log(fmt.Sprintf("%s -> %s", state.CurrentRoutine.Name(), routineResult.SetRoutine.Name()))
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
