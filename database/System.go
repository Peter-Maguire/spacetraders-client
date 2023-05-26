package database

import (
    "encoding/json"
    "gorm.io/gorm"
    "spacetraders/entity"
    "time"
)

type System struct {
    System        string          `gorm:"primaryKey"`
    Data          json.RawMessage `gorm:"type:json"`
    WaypointsData json.RawMessage `gorm:"type:json"`
    FirstVisited  time.Time
}

func GetSystem(system string) *System {
    visitedSystem := System{
        System: system,
    }
    tx := db.Take(&visitedSystem)
    if tx.Error == gorm.ErrRecordNotFound {
        return nil
    }
    return &visitedSystem
}

func VisitSystem(data *entity.System, waypoints *[]entity.WaypointData) {
    systemData, _ := json.Marshal(data)
    waypointsData, _ := json.Marshal(waypoints)
    db.Create(System{
        System:        data.Symbol,
        Data:          systemData,
        WaypointsData: waypointsData,
        FirstVisited:  time.Now(),
    })
}

func GetSystemData(system string) *entity.System {
    dbSystem := GetSystem(system)
    if dbSystem == nil {
        return nil
    }

    systemData := entity.System{}
    json.Unmarshal(dbSystem.Data, &systemData)
    return &systemData
}
