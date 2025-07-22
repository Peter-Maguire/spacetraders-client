package ui

import (
	"encoding/json"
	"net/http"
	"spacetraders/database"
)

func (wu *WebUI) initApi() {
	http.HandleFunc("/waypoints", func(writer http.ResponseWriter, request *http.Request) {
		encoder := json.NewEncoder(writer)

		shipyardData := database.GetWaypoints()
		encoder.Encode(shipyardData)
	})

	http.HandleFunc("/agent", func(writer http.ResponseWriter, request *http.Request) {
		encoder := json.NewEncoder(writer)

		encoder.Encode(wu.orc.GetAgent())
	})

	http.HandleFunc("/contracts", func(writer http.ResponseWriter, request *http.Request) {
		encoder := json.NewEncoder(writer)

		encoder.Encode(wu.orc.GetContract())
	})
}
