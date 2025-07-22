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

	serverStatus, _ := entity.GetServerStatus()
	fmt.Printf("SpaceTraders version %s\n", serverStatus.Version)
	fmt.Printf("%s\n", serverStatus.Status)
	fmt.Printf("Next Reset: %s\n", serverStatus.ResetDate)

	if os.Getenv("TOKEN") == "" {
		fmt.Println("Token not provided")
		return
	}

	fmt.Println("Starting Database...")
	database.Init()

	fmt.Println("Starting Request Queue...")
	http.Init()

	fmt.Println("Starting Orchestrators...")
	tokens := strings.Split(os.Getenv("TOKEN"), ",")
	orcs = make([]*orchestrator.Orchestrator, len(tokens))
	for i, token := range tokens {
		orcs[i] = orchestrator.Init(token)
	}

	enableUi = os.Getenv("DISABLE_UI") != "1"
	if enableUi {
		fmt.Println("Starting UI...")
		go ui.Init(orcs[0])
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
		for _, orc := range orcs {
			numWaiting := 0
			numSleeping := 0
			numStopped := 0
			routinesActive.Reset()
			for i, state := range orc.States {
				if state != nil && state.Ship != nil {

					shipData = append(shipData, ui.ShipData{
						Stopped:        state.CurrentRoutine == nil,
						StoppedReason:  state.StoppedReason,
						WaitingForHttp: state.WaitingForHttp,
						AsleepUntil:    state.AsleepUntil,
						ShipName:       state.Ship.Symbol,
						ShipType:       string(state.Ship.Registration.Role),
						Nav:            *state.Ship.Nav,
						Cargo:          *state.Ship.Cargo,
					})
					if state.CurrentRoutine == nil {
						numStopped++
					} else {
						routinesActive.WithLabelValues(state.CurrentRoutine.Name()).Add(1)
						shipData[i].Routine = state.CurrentRoutine.Name()
					}
					if state.WaitingForHttp {
						numWaiting++
					}
					if state.AsleepUntil != nil {
						numSleeping++
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

			ui.WriteShipState(shipData, httpData)
		}
	}
}
