package database

import (
	"spacetraders/entity"
	"time"
)

type MarketRates struct {
	Waypoint string
	Good     string
	SellCost int
	BuyCost  int
	Date     time.Time
}

func StoreMarketRates(waypoint string, goods []entity.MarketGood) {
	rates := make([]MarketRates, len(goods))
	for i, good := range goods {
		rates[i] = MarketRates{
			Waypoint: waypoint,
			Good:     good.Symbol,
			SellCost: good.SellPrice,
			BuyCost:  good.PurchasePrice,
			Date:     time.Now(),
		}
	}
	db.Create(rates)
}
