package database

import "spacetraders/entity"

type ShipCost struct {
	Waypoint      string
	ShipType      string
	PurchasePrice int
}

func StoreShipCosts(costs *entity.ShipyardStock) {
	output := make([]ShipCost, len(costs.Ships))
	for i, ship := range costs.Ships {
		output[i] = ShipCost{
			Waypoint:      costs.Symbol,
			ShipType:      ship.Type,
			PurchasePrice: ship.PurchasePrice,
		}
	}
	db.Create(output)
}
