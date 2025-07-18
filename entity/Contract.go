package entity

import (
	"context"
	"fmt"
	"spacetraders/http"
	"time"
)

type Contract struct {
	Id            string        `json:"id"`
	FactionSymbol string        `json:"factionSymbol"`
	Type          string        `json:"type"`
	Terms         ContractTerms `json:"terms"`
	Accepted      bool          `json:"accepted"`
	Fulfilled     bool          `json:"fulfilled"`
	Expiration    time.Time     `json:"expiration"`
}

func (c *Contract) Accept(ctx context.Context) *http.HttpError {
	_, err := http.Request[any](ctx, "POST", fmt.Sprintf("my/contracts/%s/accept", c.Id), nil)
	return err
}

func (c *Contract) Deliver(ctx context.Context, shipSymbol string, tradeSymbol string, units int) (*ContractDeliverResult, *http.HttpError) {
	deliverResult, err := http.Request[ContractDeliverResult](ctx, "POST", fmt.Sprintf("my/contracts/%s/deliver", c.Id), ContractDeliveryRequest{
		ShipSymbol:  shipSymbol,
		TradeSymbol: tradeSymbol,
		Units:       units,
	})

	if deliverResult != nil {
		c.Terms = deliverResult.Contract.Terms
	}

	return deliverResult, err
}

func (c *Contract) Fulfill(ctx context.Context) error {
	_, err := http.Request[any](ctx, "POST", fmt.Sprintf("my/contracts/%s/fulfill", c.Id), nil)
	return err
}

type ContractTerms struct {
	Deadline time.Time             `json:"deadline"`
	Payment  ContractPayment       `json:"payment"`
	Deliver  []ContractDeliverable `json:"deliver"`
}

func (ct *ContractTerms) GetDeliverable(item string) *ContractDeliverable {
	for _, deliverable := range ct.Deliver {
		if deliverable.TradeSymbol == item {
			return &deliverable
		}
	}
	return nil
}

type ContractPayment struct {
	OnAccepted  int `json:"onAccepted"`
	OnFulfilled int `json:"onFulfilled"`
}

func (cp ContractPayment) GetTotalPayment() int {
	return cp.OnAccepted + cp.OnFulfilled
}

type ContractDeliverable struct {
	TradeSymbol       string   `json:"tradeSymbol"`
	DestinationSymbol Waypoint `json:"destinationSymbol"`
	UnitsRequired     int      `json:"unitsRequired"`
	UnitsFulfilled    int      `json:"unitsFulfilled"`
}

func (cd *ContractDeliverable) IsFulfilled() bool {
	return cd.UnitsFulfilled >= cd.UnitsRequired
}

type ContractDeliveryRequest struct {
	ShipSymbol  string `json:"shipSymbol"`
	TradeSymbol string `json:"tradeSymbol"`
	Units       int    `json:"units"`
}
