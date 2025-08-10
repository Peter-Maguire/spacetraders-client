package main

import (
	"context"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"os"
	"spacetraders/database"
	"spacetraders/entity"
	"spacetraders/http"
	"spacetraders/orchestrator"
	"spacetraders/ui"
	"time"
)

var enableUi = false

var orcs []*orchestrator.Orchestrator

var (
	httpBacklog = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "st_http_backlog",
		Help: "Backlog of HTTP Requests",
	})
	routineWaiting = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "st_routine_num_waiting",
		Help: "Number of routines waiting for HTTP requests",
	}, []string{"agent"})
	routineStopped = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "st_routine_num_stopped",
		Help: "Number of routines stopped",
	}, []string{"agent"})
	routineSleeping = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "st_routine_num_sleeping",
		Help: "Number of routines sleeping",
	}, []string{"agent"})
	routinesActive = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "st_routine_count",
		Help: "Number of routines per type",
	}, []string{"name", "agent"})
)

func main() {

	st := entity.SpaceTraders{}
	serverStatus, err := entity.GetServerStatus()
	if err != nil {
		panic(err)
	}
	fmt.Printf("SpaceTraders version %s\n", serverStatus.Version)
	fmt.Printf("%s\n", serverStatus.Status)
	fmt.Printf("Next Reset: %s\n", serverStatus.ServerResets.Next)

	st.ServerStart, _ = time.Parse(time.DateOnly, serverStatus.ResetDate)
	st.ServerEnd, _ = time.Parse(time.RFC3339, serverStatus.ServerResets.Next)

	if os.Getenv("RESET") == "true" {
		fmt.Println("Resetting...")
		time.Sleep(10 * time.Second)
		database.Init()
		database.Reset()
		return
	}

	enableUi = os.Getenv("DISABLE_UI") != "1"
	if enableUi {
		fmt.Println("Starting UI...")
		go ui.Init(&st)
	}

	fmt.Println("Starting Database...")
	database.Init()

	fmt.Println("Starting Request Queue...")
	http.Init()

	fmt.Println("Starting Orchestrators...")
	agents := database.GetEnabledAgents()
	fmt.Printf("%d agents enabled\n", len(agents))

	ctx := context.WithValue(context.Background(), "token", os.Getenv("ACCOUNT_TOKEN"))

	orcs = make([]*orchestrator.Orchestrator, len(agents))
	for i, agent := range agents {

		if agent.Token == "" {
			fmt.Println("Creating new agent", agent.Symbol)
			registerResponse, err := entity.RegisterAgent(ctx, agent.Symbol, agent.Config.GetString("faction", "COSMIC"))
			if err != nil {
				panic(err)
			}
			agent.Token = registerResponse.Token
			database.SetAgentToken(agent.Symbol, agent.Token)
		}

		orc := orchestrator.Init(agent)
		orcs[i] = orc
	}

	if enableUi {
		updateShipStates()
	}
	//ticker := time.NewTicker(1 * time.Second)
	//for {
	//	<-ticker.C
	//}
}

func updateShipStates() {
	ticker := time.NewTicker(500 * time.Millisecond)
	for {
		<-ticker.C
		shipData := make([]ui.ShipData, 0)
		agents := make(map[string]*entity.Agent)
		contracts := make(map[string]*entity.Contract)
		routinesActive.Reset()
		for _, orc := range orcs {
			numWaiting := 0
			numSleeping := 0
			numStopped := 0
			agents[orc.Agent.Symbol] = orc.Agent
			contracts[orc.Agent.Symbol] = orc.Contract
			for _, state := range orc.States {
				if state != nil && state.Ship != nil {

					ship := ui.ShipData{
						Stopped:          state.CurrentRoutine == nil,
						StoppedReason:    state.StoppedReason,
						WaitingForHttp:   state.WaitingForHttp,
						AsleepUntil:      state.AsleepUntil,
						ShipName:         state.Ship.Symbol,
						ShipType:         string(state.Ship.Registration.Role),
						Nav:              *state.Ship.Nav,
						Cargo:            *state.Ship.Cargo,
						Fuel:             *state.Ship.Fuel,
						ConstructionSite: state.ConstructionSite,
					}

					if state.CurrentRoutine == nil {
						numStopped++
					} else {
						routinesActive.WithLabelValues(fmt.Sprintf("%T", state.CurrentRoutine), orc.Agent.Symbol).Add(1)
						ship.Routine = state.CurrentRoutine.Name()
					}
					if state.WaitingForHttp {
						numWaiting++
					}
					if state.AsleepUntil != nil {
						numSleeping++
					}
					shipData = append(shipData, ship)
				}
			}

			routineWaiting.WithLabelValues(orc.Agent.Symbol).Set(float64(numWaiting))
			routineSleeping.WithLabelValues(orc.Agent.Symbol).Set(float64(numSleeping))
			routineStopped.WithLabelValues(orc.Agent.Symbol).Set(float64(numStopped))

		}
		httpData := ui.HttpData{
			Active: http.IsRunningRequests,
		}

		numBacklog := 0
		http.RBufferLock.Lock()
		httpList := make([]ui.HttpRequestList, len(http.RequestBuffer))
		for i, request := range http.RequestBuffer {
			numBacklog++
			httpList[i] = ui.HttpRequestList{
				Receivers: len(request.ReturnChannels),
				Priority:  request.Priority,
				Method:    request.Req.Method,
				Path:      request.OriginalPath,
			}
		}
		http.RBufferLock.Unlock()
		httpBacklog.Set(float64(numBacklog))
		httpData.Requests = httpList
		ui.WriteShipState(shipData, httpData, contracts, agents)
	}
}
