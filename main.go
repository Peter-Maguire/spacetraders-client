package main

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"os"
	"spacetraders/database"
	"spacetraders/entity"
	"spacetraders/http"
	"spacetraders/orchestrator"
	"spacetraders/ui"
	"strings"
	"time"
)

var enableUi = false

var orcs []*orchestrator.Orchestrator

var (
	httpBacklog = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "st_http_backlog",
		Help: "Backlog of HTTP Requests",
	})
	routineWaiting = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "st_routine_num_waiting",
		Help: "Number of routines waiting for HTTP requests",
	})
	routineStopped = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "st_routine_num_stopped",
		Help: "Number of routines stopped",
	})
	routineSleeping = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "st_routine_num_sleeping",
		Help: "Number of routines sleeping",
	})
	routinesActive = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "st_routine_count",
		Help: "Number of routines per type",
	}, []string{"name"})
)

func main() {

	st := entity.SpaceTraders{}
	serverStatus, _ := entity.GetServerStatus()
	fmt.Printf("SpaceTraders version %s\n", serverStatus.Version)
	fmt.Printf("%s\n", serverStatus.Status)
	fmt.Printf("Next Reset: %s\n", serverStatus.ServerResets.Next)

	st.ServerStart, _ = time.Parse(time.DateOnly, serverStatus.ResetDate)
	st.ServerEnd, _ = time.Parse(time.RFC3339, serverStatus.ServerResets.Next)

	if os.Getenv("TOKEN") == "" {
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
	tokens := strings.Split(os.Getenv("TOKEN"), ",")
	orcs = make([]*orchestrator.Orchestrator, len(tokens))
	st.Orchestrators = make([]entity.Orchestrator, len(tokens))
	for i, token := range tokens {
		orc := orchestrator.Init(token)
		orcs[i] = orc
		st.Orchestrators[i] = orc
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
	ticker := time.NewTicker(1 * time.Second)
	for {
		<-ticker.C
		shipData := make([]ui.ShipData, 0)
		agents := make(map[string]*entity.Agent)
		contracts := make(map[string]*entity.Contract)
		numWaiting := 0
		numSleeping := 0
		numStopped := 0
		routinesActive.Reset()
		for _, orc := range orcs {
			agents[orc.Agent.Symbol] = orc.Agent
			contracts[orc.Agent.Symbol] = orc.Contract
			for _, state := range orc.States {
				if state != nil && state.Ship != nil {

					ship := ui.ShipData{
						Stopped:        state.CurrentRoutine == nil,
						StoppedReason:  state.StoppedReason,
						WaitingForHttp: state.WaitingForHttp,
						AsleepUntil:    state.AsleepUntil,
						ShipName:       state.Ship.Symbol,
						ShipType:       string(state.Ship.Registration.Role),
						Nav:            *state.Ship.Nav,
						Cargo:          *state.Ship.Cargo,
						Fuel:           *state.Ship.Fuel,
					}

					if state.CurrentRoutine == nil {
						numStopped++
					} else {
						routinesActive.WithLabelValues(state.CurrentRoutine.Name()).Add(1)
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

		httpData.Requests = httpList
		routineWaiting.Set(float64(numWaiting))
		routineSleeping.Set(float64(numSleeping))
		routineStopped.Set(float64(numStopped))
		httpBacklog.Set(float64(numBacklog))

		ui.WriteShipState(shipData, httpData, contracts, agents)
	}
}
