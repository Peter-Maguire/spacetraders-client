package ui

import (
    "encoding/json"
    "net/http"
    "spacetraders/database"
)

func initApi() {
    http.HandleFunc("/shipyards", func(writer http.ResponseWriter, request *http.Request) {
        encoder := json.NewEncoder(writer)

        shipyardData := database.GetWaypoints()
        encoder.Encode(shipyardData)

    })
}
