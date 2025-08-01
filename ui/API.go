package ui

import (
	"encoding/json"
	"net/http"
	"spacetraders/database"
	"spacetraders/entity"
)

func (wu *WebUI) initApi() {
	http.HandleFunc("/waypoints", func(writer http.ResponseWriter, request *http.Request) {
		encoder := json.NewEncoder(writer)

		shipyardData := database.GetWaypoints()
		encoder.Encode(shipyardData)
	})

	http.HandleFunc("/agent", func(writer http.ResponseWriter, request *http.Request) {
		encoder := json.NewEncoder(writer)
		if wu.st == nil {
			encoder.Encode(map[string]any{})
		}
		agents := make(map[string]*entity.Agent)
		for _, orc := range wu.st.Orchestrators {
			agent := orc.GetAgent()
			agents[agent.Symbol] = agent
		}
		encoder.Encode(agents)
	})

	http.HandleFunc("/contracts", func(writer http.ResponseWriter, request *http.Request) {
		encoder := json.NewEncoder(writer)
		if wu.st == nil {
			encoder.Encode(map[string]any{})
		}
		contracts := make(map[string]*entity.Contract)
		for _, orc := range wu.st.Orchestrators {
			agent := orc.GetAgent()
			contract := orc.GetContract()
			contracts[agent.Symbol] = contract
		}
		encoder.Encode(contracts)
	})

	http.HandleFunc("/status", func(writer http.ResponseWriter, request *http.Request) {
		encoder := json.NewEncoder(writer)

		encoder.Encode(wu.st)
	})

	http.HandleFunc("/systems", func(writer http.ResponseWriter, request *http.Request) {
		encoder := json.NewEncoder(writer)

		systems := database.GetSystems()

		encoder.Encode(systems)
	})
}
