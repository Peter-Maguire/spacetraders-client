package entity

import (
	"fmt"
	"spacetraders/http"
)

type Agent struct {
	AccountId    string   `json:"accountId"`
	Symbol       string   `json:"symbol"`
	Headquarters Waypoint `json:"headquarters"`
	Credits      int      `json:"credits"`
}

func GetAgent() (*Agent, *http.HttpError) {
	return http.Request[Agent]("GET", "my/agent", nil)
}

func (a *Agent) Ships() (*[]Ship, error) {
	return http.PaginatedRequest[Ship]("my/ships", 1, 0)
}

func (a *Agent) Contracts() (*[]Contract, error) {
	return http.Request[[]Contract]("GET", "my/contracts", nil)
}

func (a *Agent) Systems(page int) (*[]System, error) {
	return http.Request[[]System]("GET", fmt.Sprintf("systems?total=20&page=%d", page), nil)
}

func (a *Agent) GetSystem(system string) (*System, *http.HttpError) {
	return http.Request[System]("GET", fmt.Sprintf("systems/%s", system), nil)
}

func (a *Agent) BuyShip(shipyard Waypoint, shipType string) (*ShipPurchaseResult, *http.HttpError) {
	result, err := http.Request[ShipPurchaseResult]("POST", "my/ships", ShipPurchaseRequest{
		ShipType:       shipType,
		WaypointSymbol: shipyard,
	})
	return result, err
}
