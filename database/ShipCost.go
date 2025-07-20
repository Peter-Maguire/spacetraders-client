package database

import (
	"fmt"
	"gorm.io/gorm/clause"
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
	tx := db.Clauses(clause.OnConflict{DoNothing: true}).Save(output)
	fmt.Println("Shipyard data error", tx.Error)
}

func GetShipCosts() []*ShipCost {
	shipCosts := make([]*ShipCost, 0)
	db.Find(&shipCosts)
	return shipCosts
}
