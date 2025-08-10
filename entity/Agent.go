package entity

import (
	"context"
	"fmt"
	"spacetraders/http"
)

type Agent struct {
	AccountId    string   `json:"accountId"`
	Symbol       string   `json:"symbol"`
	Headquarters Waypoint `json:"headquarters"`
	Credits      int      `json:"credits"`
}

type AgentRegistration struct {
	Faction string `json:"faction"`
	Symbol  string `json:"symbol"`
}

type RegisterAgentResponse struct {
	Token    string    `json:"token"`
	Agent    *Agent    `json:"agent"`
	Faction  *Faction  `json:"faction"`
	Contract *Contract `json:"contract"`
	Ships    []*Ship   `json:"ships"`
}

func RegisterAgent(ctx context.Context, symbol string, faction string) (*RegisterAgentResponse, *http.HttpError) {
	return http.Request[RegisterAgentResponse](ctx, "POST", "register", AgentRegistration{
		Faction: faction,
		Symbol:  symbol,
	})
}

func GetAgent(ctx context.Context) (*Agent, *http.HttpError) {
	return http.Request[Agent](ctx, "GET", "my/agent", nil)
}

func (a *Agent) Ships(ctx context.Context) (*[]Ship, error) {
	return http.PaginatedRequest[Ship](ctx, "my/ships", 1, 0)
}

func (a *Agent) Contracts(ctx context.Context) (*[]Contract, error) {
	return http.PaginatedRequest[Contract](ctx, "my/contracts", 1, 0)
}

func (a *Agent) Systems(ctx context.Context, page int) (*[]System, *http.HttpError) {
	return http.Request[[]System](ctx, "GET", fmt.Sprintf("systems?total=20&page=%d", page), nil)
}

func (a *Agent) GetSystem(ctx context.Context, system string) (*System, *http.HttpError) {
	return http.Request[System](ctx, "GET", fmt.Sprintf("systems/%s", system), nil)
}

func (a *Agent) BuyShip(ctx context.Context, shipyard Waypoint, shipType string) (*ShipPurchaseResult, *http.HttpError) {
	result, err := http.Request[ShipPurchaseResult](ctx, "POST", "my/ships", ShipPurchaseRequest{
		ShipType:       shipType,
		WaypointSymbol: shipyard,
	})

	if err != nil && err.Code == http.ErrPurchaseShipCredits {
		fmt.Println(err.Data)
		//a.Credits = err.Data["creditsAvailable"].(int)
	}

	if result != nil {
		a.Credits = result.Agent.Credits
	}

	return result, err
}

func (a *Agent) GetShip(ctx context.Context, ship string) (*Ship, *http.HttpError) {
	return http.Request[Ship](ctx, "GET", fmt.Sprintf("my/ships/%s", ship), nil)
}
