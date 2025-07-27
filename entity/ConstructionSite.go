package entity

import (
	"context"
	"fmt"
	"spacetraders/http"
)

type ConstructionSite struct {
	Symbol     Waypoint               `json:"symbol"`
	IsComplete bool                   `json:"isComplete"`
	Materials  []ConstructionMaterial `json:"materials"`
}

func (cs *ConstructionSite) GetMaterial(symbol string) *ConstructionMaterial {
	for _, material := range cs.Materials {
		if material.TradeSymbol == symbol {
			return &material
		}
	}
	return nil
}

func (cs *ConstructionSite) Update(ctx context.Context) {
	result, _ := cs.Symbol.GetConstructionSite(ctx)
	if result != nil {
		cs.Materials = result.Materials
		cs.IsComplete = result.IsComplete
	}
}

func (cs *ConstructionSite) Supply(ctx context.Context, ship *Ship, tradeSymbol string, units int) (*SupplyConstructionSiteResponse, *http.HttpError) {
	supply, err := http.Request[SupplyConstructionSiteResponse](ctx, "POST", fmt.Sprintf("systems/%s/waypoints/%s/construction/supply", cs.Symbol.GetSystemName(), cs.Symbol), map[string]any{
		"shipSymbol":  ship.Symbol,
		"tradeSymbol": tradeSymbol,
		"units":       units,
	})

	if err != nil {
		return nil, err
	}

	ship.Cargo = supply.Cargo
	cs.Materials = supply.ConstructionSite.Materials
	cs.IsComplete = supply.ConstructionSite.IsComplete
	return supply, err
}

type ConstructionMaterial struct {
	TradeSymbol string `json:"tradeSymbol"`
	Required    int    `json:"required"`
	Fulfilled   int    `json:"fulfilled"`
}

func (cm *ConstructionMaterial) IsComplete() bool {
	return cm.Fulfilled >= cm.Required
}

func (cm *ConstructionMaterial) GetRemaining() int {
	return cm.Required - cm.Fulfilled
}

type SupplyConstructionSiteResponse struct {
	ConstructionSite *ConstructionSite `json:"construction"`
	Cargo            *ShipCargo        `json:"cargo"`
}
