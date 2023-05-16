package ui

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

var upgrader = websocket.Upgrader{}

func Init() {

	go broadcastLoop()

	http.HandleFunc("/ws", ws)
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)
	http.ListenAndServe("0.0.0.0:8080", nil)

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
	fmt.Print(str)
	broadcasts <- BroadcastMessage{
		Type: "log",
		Data: str,
	}
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

func WriteShipState(shipStates string, httpState string) {
	broadcasts <- BroadcastMessage{
		Type: "state",
		Data: []string{shipStates, httpState},
	}
}
