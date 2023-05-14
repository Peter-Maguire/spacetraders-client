package entity

import "spacetraders/http"

type Agent struct {
	AccountId    string   `json:"accountId"`
	Symbol       string   `json:"symbol"`
	Headquarters Waypoint `json:"headquarters"`
	Credits      int      `json:"credits"`
}

func (a *Agent) Ships() (*[]Ship, error) {
	return http.Request[[]Ship]("GET", "my/ships", nil)
}

func (a *Agent) Contracts() (*[]Contract, error) {
	return http.Request[[]Contract]("GET", "my/contracts", nil)
}

func (a *Agent) BuyShip(shipyard Waypoint, shipType string) (*ShipPurchaseResult, error) {
	result, err := http.Request[ShipPurchaseResult]("POST", "my/ships", ShipPurchaseRequest{
		ShipType:       shipType,
		WaypointSymbol: shipyard,
	})
	return result, err
}
