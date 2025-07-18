package database

import (
	"fmt"
	"spacetraders/entity"
	"time"
)

type MarketExchange struct {
	Waypoint  entity.Waypoint `gorm:"primaryKey"`
	Type      string          `gorm:"primaryKey"`
	Good      string          `gorm:"primaryKey"`
	SystemX   int
	SystemY   int
	WaypointX int
	WaypointY int
	Date      time.Time
}

func StoreMarketExchange(system *entity.System, waypointData *entity.WaypointData, exchangeType string, goods []entity.ExchangeItem) {
	rates := make([]MarketExchange, len(goods))
	for i, good := range goods {
		rates[i] = MarketExchange{
			Waypoint:  waypointData.Symbol,
			Type:      exchangeType,
			Good:      good.Symbol,
			SystemX:   system.X,
			SystemY:   system.Y,
			WaypointX: waypointData.X,
			WaypointY: waypointData.Y,
			Date:      time.Now(),
		}
	}
	tx := db.Save(rates)
	fmt.Println(tx.Error)
}
