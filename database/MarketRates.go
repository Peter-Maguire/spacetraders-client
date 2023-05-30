package database

import (
	"fmt"
	"gorm.io/gorm"
	"spacetraders/entity"
	"time"
)

type MarketRates struct {
	Waypoint    entity.Waypoint `gorm:"primaryKey"`
	Good        string          `gorm:"primaryKey"`
	SystemX     int
	SystemY     int
	WaypointX   int
	WaypointY   int
	SellCost    int
	BuyCost     int
	TradeVolume int
	Date        time.Time
}

func StoreMarketRates(system *entity.System, waypointData *entity.WaypointData, goods []entity.MarketGood) {
	rates := make([]MarketRates, len(goods))
	for i, good := range goods {
		rates[i] = MarketRates{
			Waypoint:    waypointData.Symbol,
			Good:        good.Symbol,
			SystemX:     system.X,
			SystemY:     system.Y,
			WaypointX:   waypointData.X,
			WaypointY:   waypointData.Y,
			SellCost:    good.SellPrice,
			BuyCost:     good.PurchasePrice,
			TradeVolume: good.TradeVolume,
			Date:        time.Now(),
		}
	}
	tx := db.Save(rates)
	fmt.Println(tx.Error)
}

func UpdateMarketRates(waypoint entity.Waypoint, goods []entity.MarketGood) {
	for _, good := range goods {
		marketRate := MarketRates{
			Waypoint:    waypoint,
			Good:        good.Symbol,
			SellCost:    good.SellPrice,
			BuyCost:     good.PurchasePrice,
			TradeVolume: good.TradeVolume,
			Date:        time.Now(),
		}
		_ = db.Model(&marketRate).Updates(&marketRate)
	}
}

func GetMarketsSelling(items []string) []MarketRates {
	var rates []MarketRates
	tx := db.Where("good IN ?", items).Find(&rates)
	if tx.Error != nil && tx.Error != gorm.ErrRecordNotFound {
		fmt.Println("GetMarketsSelling error", tx.Error)
	}
	return rates
}

func GetMarkets() []MarketRates {
	var rates []MarketRates
	db.Find(&rates)
	return rates
}

func GetMarketWaypoints() []entity.Waypoint {
	var waypoints []entity.Waypoint
	tx := db.Table("market_rates").Distinct("waypoint").Find(&waypoints)
	fmt.Println(tx.Error)
	return waypoints
}
