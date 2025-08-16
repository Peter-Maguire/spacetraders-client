package ui

import (
	"fmt"
	"log"
	"net/http"
	"spacetraders/entity"

	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"time"
)

var upgrader = websocket.Upgrader{}

type WebUI struct {
	st *entity.SpaceTraders
}

func Init(st *entity.SpaceTraders) *WebUI {
	webUi := WebUI{
		st: st,
	}
	go broadcastLoop()
	go webUi.initApi()

	sentryHandler := sentryhttp.New(sentryhttp.Options{})

	http.Handle("/metrics", sentryHandler.Handle(promhttp.Handler()))
	http.HandleFunc("/ws", sentryHandler.HandleFunc(ws))
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)
	http.ListenAndServe("0.0.0.0:8080", nil)
	return &webUi
}

var clients []*websocket.Conn

var broadcasts = make(chan BroadcastMessage, 10)

func ws(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	clients = append(clients, c)
}

type BroadcastMessage struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

func MainLog(str string) {
	fmt.Println(str)
	go func() {
		broadcasts <- BroadcastMessage{
			Type: "log",
			Data: str,
		}
	}()
}

func broadcastLoop() {
	for {
		message := <-broadcasts
		for i, c := range clients {
			err := c.WriteJSON(message)
			if err != nil {
				fmt.Println("Write error ", err)
				_ = c.Close()
				clen := len(clients) - 1
				clients[i] = clients[clen]
				clients = clients[:clen]
			}
		}
	}
}

type ShipData struct {
	Stopped          bool                     `json:"stopped"`
	StoppedReason    string                   `json:"stoppedReason,omitempty"`
	WaitingForHttp   bool                     `json:"waitingForHttp"`
	AsleepUntil      *time.Time               `json:"asleepUntil"`
	WaitingForEvent  string                   `json:"waitingForEvent"`
	ShipName         string                   `json:"name"`
	ShipType         string                   `json:"type"`
	Routine          string                   `json:"routine"`
	Nav              entity.ShipNav           `json:"nav"`
	Cargo            entity.ShipCargo         `json:"cargo"`
	Fuel             entity.ShipFuel          `json:"fuel"`
	ConstructionSite *entity.ConstructionSite `json:"constructionSite"`
}

type HttpData struct {
	Active   bool              `json:"active"`
	Requests []HttpRequestList `json:"requests"`
}

type HttpRequestList struct {
	Receivers int    `json:"receivers"`
	Priority  int    `json:"priority"`
	Method    string `json:"method"`
	Path      string `json:"path"`
}

func WriteShipState(shipStates []ShipData, httpState HttpData, contracts map[string]*entity.Contract, agents map[string]*entity.Agent) {
	broadcasts <- BroadcastMessage{
		Type: "state",
		Data: map[string]any{
			"contracts": contracts,
			"agents":    agents,
			"ship":      shipStates,
			"http":      httpState,
		},
	}
}
