package database

import (
    "fmt"
    "gorm.io/gorm"
    "spacetraders/entity"
    "time"
)

type MarketRates struct {
    Waypoint  string `gorm:"primaryKey"`
    Good      string `gorm:"primaryKey"`
    WaypointX int
    WaypointY int
    SellCost  int
    BuyCost   int
    Date      time.Time
}

func StoreMarketRates(waypointData *entity.WaypointData, goods []entity.MarketGood) {

    rates := make([]MarketRates, len(goods))
    for i, good := range goods {
        rates[i] = MarketRates{
            Waypoint:  string(waypointData.Symbol),
            Good:      good.Symbol,
            WaypointX: waypointData.X,
            WaypointY: waypointData.Y,
            SellCost:  good.SellPrice,
            BuyCost:   good.PurchasePrice,
            Date:      time.Now(),
        }
    }
    db.Save(rates)
}

func GetMarketsSelling(items []string) []MarketRates {
    var rates []MarketRates
    tx := db.Where("good IN ?", items).Find(&rates)
    if tx.Error != gorm.ErrRecordNotFound {
        fmt.Println("GetMarketsSelling error", tx.Error)
    }
    return rates
}
