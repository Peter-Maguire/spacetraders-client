package entity

import (
	"fmt"
	"spacetraders/http"
	"time"
)

type Contract struct {
	Id            string        `json:"id"`
	FactionSymbol string        `json:"factionSymbol"`
	Type          string        `json:"type"`
	Terms         ContractTerms `json:"terms"`
}

func (c *Contract) Accept() *http.HttpError {
	_, err := http.Request[any]("POST", fmt.Sprintf("my/contracts/%s/accept", c.Id), nil)
	return err
}

func (c *Contract) Deliver(shipSymbol string, tradeSymbol string, units int) (*ContractDeliverResult, *http.HttpError) {
	deliverResult, err := http.Request[ContractDeliverResult]("POST", fmt.Sprintf("my/contracts/%s/deliver", c.Id), ContractDeliveryRequest{
		ShipSymbol:  shipSymbol,
		TradeSymbol: tradeSymbol,
		Units:       units,
	})

	if deliverResult != nil {
		c.Terms = deliverResult.Contract.Terms
	}

	return deliverResult, err
}

func (c *Contract) Fulfill() error {
	_, err := http.Request[any]("POST", fmt.Sprintf("my/contracts/%s/fulfill", c.Id), nil)
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

type ContractDeliverable struct {
	TradeSymbol       string   `json:"tradeSymbol"`
	DestinationSymbol Waypoint `json:"destinationSymbol"`
	UnitsRequired     int      `json:"unitsRequired"`
	UnitsFulfilled    int      `json:"unitsFulfilled"`
}

type ContractDeliveryRequest struct {
	ShipSymbol  string `json:"shipSymbol"`
	TradeSymbol string `json:"tradeSymbol"`
	Units       int    `json:"units"`
}
