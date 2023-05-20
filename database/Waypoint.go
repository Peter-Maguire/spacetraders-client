package database

import (
    "encoding/json"
    "gorm.io/gorm"
    "spacetraders/entity"
    "time"
)

type Waypoint struct {
    Waypoint     string `gorm:"primaryKey"`
    System       string
    Data         []byte `gorm:"type:json"`
    MarketData   []byte `gorm:"type:json"`
    ShipyardData []byte `gorm:"type:json"`
    FirstVisited time.Time
}

func GetWaypoint(waypoint entity.Waypoint) *Waypoint {
    visitedWaypoint := Waypoint{
        Waypoint: string(waypoint),
    }
    tx := db.Take(&visitedWaypoint)
    if tx.Error == gorm.ErrRecordNotFound {
        return nil
    }
    return &visitedWaypoint
}

func GetShipyards(symbol string) *[]Waypoint {
    var wp []Waypoint
    tx := db.Where("shipyard_data is not null").Find(&wp)
    if tx.Error == gorm.ErrRecordNotFound {
        return nil
    }
    return &wp
}

func VisitWaypoint(data *entity.WaypointData, market *entity.Market, shipyard *entity.ShipyardStock) {
    var shipyardData, marketData []byte
    waypointData, _ := json.Marshal(data)
    if market != nil {
        marketData, _ = json.Marshal(market)
    }
    if shipyard != nil {
        shipyardData, _ = json.Marshal(shipyard)
    }
    db.Create(Waypoint{
        Waypoint:     string(data.Symbol),
        System:       data.Symbol.GetSystemName(),
        Data:         waypointData,
        ShipyardData: shipyardData,
        MarketData:   marketData,
        FirstVisited: time.Now(),
    })
}
