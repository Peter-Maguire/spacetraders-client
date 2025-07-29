package database

import (
	"encoding/json"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	TimesVisited int `gorm:"default:0"`
}

func (w *Waypoint) Visit() error {
	w.FirstVisited = time.Now()
	tx := db.Save(&w)
	return tx.Error
}

func (w *Waypoint) GetData() entity.WaypointData {
	wpData := entity.WaypointData{}
	err := json.Unmarshal(w.Data, &wpData)
	if err != nil {
		panic(err)
	}
	return wpData
}

func (w *Waypoint) GetMarketData() *entity.Market {
	if w.MarketData == nil || string(w.MarketData) == "null" {
		return nil
	}
	mData := entity.Market{}
	err := json.Unmarshal(w.MarketData, &mData)
	if err != nil {
		return nil
	}
	return &mData
}

func (w *Waypoint) GetShipyardData() *entity.ShipyardStock {
	if w.ShipyardData == nil || string(w.ShipyardData) == "null" {
		return nil
	}
	mData := entity.ShipyardStock{}
	err := json.Unmarshal(w.ShipyardData, &mData)
	if err != nil {
		return nil
	}
	return &mData
}

func GetLeastVisitedWaypointsInSystem(system string, currentTimesVisited int) []*Waypoint {
	var wp []*Waypoint
	db.Order("times_visited ASC").Where("system = ? AND times_visited < ?", system, currentTimesVisited).Find(&wp)
	return wp
}

func GetUnvisitedWaypointsInSystem(system string) []*Waypoint {
	var wp []*Waypoint
	db.Where("times_visited = 0 AND system = ?", system).Find(&wp)
	return wp
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

func LogWaypoints(data *[]entity.WaypointData) {
	dbWaypoints := make([]Waypoint, len(*data))
	for i, wp := range *data {
		waypointData, _ := json.Marshal(wp)
		dbWaypoints[i] = Waypoint{
			Waypoint:     string(wp.Symbol),
			System:       wp.SystemSymbol,
			Data:         waypointData,
			MarketData:   nil,
			ShipyardData: nil,
			FirstVisited: time.Time{},
			TimesVisited: 0,
		}
	}
	db.Clauses(clause.OnConflict{DoNothing: true}).Save(dbWaypoints)
}

func VisitWaypoint(data *entity.WaypointData, market *entity.Market, shipyard *entity.ShipyardStock) {
	currentWaypoint := GetWaypoint(data.Symbol)
	waypointData, _ := json.Marshal(data)
	shipyardData, _ := json.Marshal(shipyard)
	marketData, _ := json.Marshal(market)
	if currentWaypoint == nil {
		currentWaypoint = &Waypoint{}
	}
	currentWaypoint.TimesVisited++
	currentWaypoint.Waypoint = string(data.Symbol)
	currentWaypoint.System = data.SystemSymbol
	currentWaypoint.Data = waypointData
	currentWaypoint.MarketData = marketData
	currentWaypoint.ShipyardData = shipyardData
	if currentWaypoint.FirstVisited.Unix() < 0 {
		currentWaypoint.FirstVisited = time.Now()
	}
	tx := db.Clauses(clause.OnConflict{UpdateAll: true}).Save(&currentWaypoint)
	fmt.Println(tx.Error)
}
