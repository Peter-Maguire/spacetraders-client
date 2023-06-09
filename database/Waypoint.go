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

type ScannedWaypoint struct {
    Waypoint     entity.Waypoint     `json:"waypoint"`
    System       string              `json:"system"`
    WaypointData entity.WaypointData `json:"waypointData"`
}

func GetWaypoints() []ScannedWaypoint {
    var wps []Waypoint
    tx := db.Find(&wps)
    if tx.Error == gorm.ErrRecordNotFound {
        return nil
    }

    waypoints := make([]ScannedWaypoint, len(wps))

    for i, wp := range wps {
        waypoints[i] = ScannedWaypoint{
            Waypoint:     entity.Waypoint(wp.Waypoint),
            System:       wp.System,
            WaypointData: entity.WaypointData{},
        }
        _ = json.Unmarshal(wp.Data, &waypoints[i].WaypointData)
    }

    return waypoints
}

func VisitWaypoint(data *entity.WaypointData) {
    waypointData, _ := json.Marshal(data)
    db.Create(Waypoint{
        Waypoint:     string(data.Symbol),
        System:       data.Symbol.GetSystemName(),
        Data:         waypointData,
        FirstVisited: time.Now(),
    })
}
