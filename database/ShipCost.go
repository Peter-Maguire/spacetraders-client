package database

import (
    "spacetraders/entity"
    "time"
)

type ShipCost struct {
    Waypoint      string `gorm:"primaryKey"`
    ShipType      string `gorm:"primaryKey"`
    PurchasePrice int
    Date          time.Time
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
