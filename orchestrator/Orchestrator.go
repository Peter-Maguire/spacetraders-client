package orchestrator

import (
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
}

var (
	contractRequirement = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "st_contract_requirement",
		Help: "Items required in contract",
	})
	agentShips = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "st_agent_ships",
		Help: "Current Ship Count",
	}, []string{"role", "system", "waypoint"})
)

func Init() *Orchestrator {

	shipFilter := os.Getenv("SHIP_FILTER")

	agent, err := entity.GetAgent()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	metrics.NumCredits.Set(float64(agent.Credits))

	contracts, _ := agent.Contracts()

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
				ui.MainLog("Accepted contract\n")
				err = c.Accept()
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
		ui.MainLog("No current Contract\n")
	} else {
		metrics.ContractProgress.Set(float64(contract.Terms.Deliver[0].UnitsFulfilled))
		contractRequirement.Set(float64(contract.Terms.Deliver[0].UnitsRequired))
	}

	orc := Orchestrator{
		Agent:        agent,
		Contract:     contract,
		Channel:      make(chan routine.OrchestratorEvent),
		CreditTarget: 300000,
	}

	waypoints, _ := agent.Headquarters.GetSystemWaypoints()

	for _, waypoint := range *waypoints {
		if waypoint.HasTrait("SHIPYARD") {
			ui.MainLog(fmt.Sprintln("Found Shipyard at ", waypoint.Symbol, "\n"))
			orc.Shipyard = waypoint.Symbol
			break
		}
	}

	ships, _ := orc.Agent.Ships()

	for _, ship := range *ships {
		agentShips.WithLabelValues(ship.Registration.Role, ship.Nav.SystemSymbol, string(ship.Nav.WaypointSymbol)).Add(1)
		if ship.Registration.Role == "HAULER" {
			ui.MainLog(fmt.Sprintf("%s is HAULER\n", ship.Registration))
			shipCopy := ship
			orc.Haulers = append(orc.Haulers, &shipCopy)
		}
	}

	shipCount := len(*ships)

	ui.MainLog(fmt.Sprintf("We have %d ships", shipCount))

	// TODO: this logic should be more nuanced
	if shipCount < 10 {
		orc.ShipToBuy = "SHIP_MINING_DRONE"
	} else if shipCount < 30 {
		orc.ShipToBuy = "SHIP_ORE_HOUND"
	} else {
		orc.ShipToBuy = "SHIP_LIGHT_HAULER"
	}

	shipyardStock, err := orc.Shipyard.GetShipyard()
	if err == nil {
		go database.StoreShipCosts(shipyardStock)
		for _, stock := range shipyardStock.Ships {
			if stock.Name == orc.ShipToBuy {
				ui.MainLog(fmt.Sprintf("Ship %s is available to buy at %s for %d credits\n", orc.ShipToBuy, orc.Shipyard, stock.PurchasePrice))
				orc.CreditTarget = stock.PurchasePrice
			}
		}
	}

	go orc.runEvents()

	orc.States = make([]*routine.State, len(*ships))

	ui.MainLog(fmt.Sprint("Starting Routines\n"))
	//orc.StatesMutex.Lock()
	for i, ship := range *ships {
		if shipFilter != "" && !strings.Contains(shipFilter, ship.Symbol) {
			ui.MainLog(fmt.Sprintf("Skipping %s because it's not in the ship filter\n", ship.Symbol))
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
		case "sellComplete":
			o.onSellComplete(event.Data.(*entity.Agent))
		case "goodSurveyFound":
			ui.MainLog("Someone found a good survey\n")
			//o.StatesMutex.Lock()
			for _, state := range o.States {
				if state.Ship.IsMiningShip() && state.Survey == nil {
					state.Survey = event.Data.(*entity.Survey)
				}
			}
			//o.StatesMutex.Unlock()
		case "surveyExhausted":
			ui.MainLog("Survey bad\n")
			//o.StatesMutex.Lock()
			for _, state := range o.States {
				if state.Survey == event.Data.(*entity.Survey) {
					state.Survey = nil
				}
			}
			//o.StatesMutex.Unlock()
		case "contractComplete":
			ui.MainLog("Contract completed")
			contracts, err := o.Agent.Contracts()
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
						ui.MainLog("Accepted new Contract " + c.Id)
						//o.StatesMutex.Lock()
						for _, state := range o.States {
							state.Contract = &c
						}
						//o.StatesMutex.Unlock()
					}
				}
			}
		}
	}
}

func (o *Orchestrator) routineLoop(state *routine.State) {
	state.CurrentRoutine = routine.DetermineObjective{}
	for {
		routineResult := state.CurrentRoutine.Run(state)
		state.WaitingForHttp = false
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
	state.CurrentRoutine = nil
	state.Log("!!!! Loop exited!")
}
