package database

import (
	"encoding/json"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"spacetraders/entity"
	"time"
)

type System struct {
	System        string          `gorm:"primaryKey"`
	Data          json.RawMessage `gorm:"type:json"`
	WaypointsData json.RawMessage `gorm:"type:json"`
	X             int
	Y             int
	Page          int
	Visited       bool
	FirstVisited  time.Time
}

func StoreSystem(system *entity.System) {
	rawSystem, _ := json.Marshal(system)
	rawWaypoints, _ := json.Marshal(system.Waypoints)

	storedSystem := System{
		System:        system.Symbol,
		Data:          rawSystem,
		WaypointsData: rawWaypoints,
		X:             system.X,
		Y:             system.Y,
		Visited:       true,
		FirstVisited:  time.Now(),
	}
	db.Clauses(clause.OnConflict{UpdateAll: true}).Save(storedSystem)
}

func GetSystem(system string) *System {
	visitedSystem := System{
		System: system,
	}
	tx := db.Take(&visitedSystem)
	fmt.Println("getSystem", tx.Error)
	if tx.Error == gorm.ErrRecordNotFound {
		return nil
	}
	return &visitedSystem
}

func AddUnvisitedSystems(data []entity.System, page int) {
	systems := make([]System, len(data))
	for i, sys := range data {
		systemData, _ := json.Marshal(sys)
		systems[i] = System{
			System:  sys.Symbol,
			Data:    systemData,
			X:       sys.X,
			Y:       sys.Y,
			Page:    page,
			Visited: false,
		}
	}
	db.Clauses(clause.OnConflict{DoNothing: true}).Save(systems)
}

func VisitSystem(data *entity.System, waypoints *[]entity.WaypointData) {
	systemData, _ := json.Marshal(data)
	waypointsData, _ := json.Marshal(waypoints)
	sys := System{
		System:        data.Symbol,
		Data:          systemData,
		WaypointsData: waypointsData,
		FirstVisited:  time.Now(),
		Visited:       true,
	}
	db.Model(sys).Updates(sys)
}

func GetUnvisitedSystems() []System {
	var systems []System
	tx := db.Where("visited = false").Order("page ASC").Find(&systems)
	if tx.Error == gorm.ErrRecordNotFound {
		return nil
	}

	return systems
}

func GetVisitedSystems() []System {
	var systems []System
	tx := db.Where("visited = true").Order("page DESC").Find(&systems)
	if tx.Error == gorm.ErrRecordNotFound {
		return nil
	}

	return systems
}

func GetSystemData(system string) *entity.System {
	dbSystem := GetSystem(system)
	if dbSystem == nil {
		return nil
	}

	systemData := entity.System{}
	err := json.Unmarshal(dbSystem.Data, &systemData)
	if err != nil {
		fmt.Println("getsystemdata", err)
	}
	return &systemData
}
