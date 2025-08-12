package orchestrator

import (
	"context"
	"fmt"
	"github.com/patrickmn/go-cache"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"os"
	"spacetraders/constant"
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
	States      []*routine.State
	StatesMutex sync.Mutex
	Agent       *entity.Agent
	Contract    *entity.Contract
	Channel     chan routine.OrchestratorEvent
	Shipyard    entity.Waypoint
	Context     context.Context
	Cache       *cache.Cache
	Config      *database.AgentConfig
}

var (
	agentShips = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "st_agent_ships",
		Help: "Current Ship Count",
	}, []string{"agent", "role", "system", "waypoint"})

	shipStates = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "st_ship_state",
		Help: "Ship States",
	}, []string{"agent", "name", "state"})
)

func Init(agentConfig database.Agent) *Orchestrator {
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

	ctx := context.WithValue(context.Background(), "token", agentConfig.Token)

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

	metrics.NumCredits.WithLabelValues(agent.Symbol).Set(float64(agent.Credits))

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
		metrics.ContractProgress.WithLabelValues(agent.Symbol).Set(float64(contract.Terms.Deliver[0].UnitsFulfilled))
		metrics.ContractRequirement.WithLabelValues(agent.Symbol).Set(float64(contract.Terms.Deliver[0].UnitsRequired))
	}

	orc := Orchestrator{
		Agent:    agent,
		Contract: contract,
		Channel:  make(chan routine.OrchestratorEvent),
		Context:  ctx,
		Cache:    cache.New(5*time.Minute, 10*time.Minute),
		Config:   &agentConfig.Config,
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
		agentShips.WithLabelValues(o.Agent.Symbol, string(ship.Registration.Role), string(ship.Nav.SystemSymbol), string(ship.Nav.WaypointSymbol)).Add(1)
		ui.MainLog(fmt.Sprintf("%s: %s type", ship.Registration.Name, ship.Registration.Role))
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
			EventBus:    &o.Channel,
			Config:      o.Config,
			Phase:       o,
		}
		state.Context = context.WithValue(o.Context, "state", &state)
		o.States[i] = &state

		go o.routineLoop(&state)
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
				EventBus: &o.Channel,
				Config:   o.Config,
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
	//defer func() {
	//	if r := recover(); r != nil {
	//		ui.MainLog(fmt.Sprintf("Recovered from panic: %v", r))
	//		state.StoppedReason = fmt.Sprint(r)
	//	}
	//}()
	state.CurrentRoutine = routine.DetermineObjective{}
	for {
		routineName := state.CurrentRoutine.Name()
		if len(routineName) > 500 {
			state.StoppedReason = "Loop Detected - " + routineName
			state.Log("Stopping Routine")
			break
		}
		shipStates.WithLabelValues(o.Agent.Symbol, state.Ship.Symbol, routineName).Set(1)
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
			shipStates.WithLabelValues(o.Agent.Symbol, state.Ship.Symbol, state.CurrentRoutine.Name()).Set(0)
			state.CurrentRoutine = state.ForceRoutine
			state.ForceRoutine = nil
			continue
		}

		if routineResult.SetRoutine != nil {
			shipStates.WithLabelValues(o.Agent.Symbol, state.Ship.Symbol, state.CurrentRoutine.Name()).Set(0)
			state.Log(fmt.Sprintf("%s => %s", state.CurrentRoutine.Name(), routineResult.SetRoutine.Name()))
			state.CurrentRoutine = routineResult.SetRoutine
		}

		if routineResult.Stop {
			shipStates.WithLabelValues(o.Agent.Symbol, state.Ship.Symbol, state.CurrentRoutine.Name()).Set(0)
			state.CurrentRoutine = nil
			state.StoppedReason = routineResult.StopReason
			state.Log("Stopping Routine")
			break
		}

		if routineResult.WaitForEvent != "" {
			state.WaitingForEvent = routineResult.WaitForEvent
			for {
				event := <-*state.EventBus
				switch event.Name {
				case routineResult.WaitForEvent:
					break
				}
			}
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

func (o *Orchestrator) internalGetPhase() constant.Phase {
	systemName := o.Agent.Headquarters.GetSystemName()
	unvisitedWaypoints := database.GetUnvisitedWaypointsInSystem(string(systemName))

	for _, uw := range unvisitedWaypoints {
		data := uw.GetData()
		if data.HasTrait(constant.TraitMarketplace) || data.HasTrait(constant.TraitShipyard) {
			return constant.PhaseExplore
		}
	}

	// TODO: this is the kind of number I'd want tob e able to change
	if len(o.States) < 20 {
		return constant.PhaseExpandFleet
	}

	waypoints, _ := systemName.GetWaypointsOfType(o.Context, constant.WaypointTypeJumpGate)
	for _, waypoint := range *waypoints {
		if waypoint.Type == constant.WaypointTypeJumpGate {
			fullWp, _ := waypoint.GetFullWaypoint(o.Context)
			if fullWp.IsUnderConstruction {
				return constant.PhaseBuildJumpGate
			}
		}
	}

	// TODO: how do we determine this phase is over?
	return constant.PhaseExplore
}

func (o *Orchestrator) GetPhase() constant.Phase {
	if cachedPhase, found := o.Cache.Get("phase"); found {
		return cachedPhase.(constant.Phase)
	}

	phase := o.internalGetPhase()
	o.Cache.Set("phase", phase, cache.DefaultExpiration)
	return phase

}
