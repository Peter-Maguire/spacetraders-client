package database

import (
	"encoding/json"
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

func GetLeastVisitedWaypointsInSystem(system string) []*Waypoint {
	var wp []*Waypoint
	db.Order("first_visited ASC").Where("system = ?", system).Find(&wp)
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
		}
	}
	db.Clauses(clause.OnConflict{DoNothing: true}).Save(dbWaypoints)
}

func VisitWaypoint(data *entity.WaypointData, market *entity.Market, shipyard *entity.ShipyardStock) {
	waypointData, _ := json.Marshal(data)
	shipyardData, _ := json.Marshal(shipyard)
	marketData, _ := json.Marshal(market)
	db.Clauses(clause.OnConflict{UpdateAll: true}).Save(Waypoint{
		Waypoint:     string(data.Symbol),
		System:       data.Symbol.GetSystemName(),
		Data:         waypointData,
		MarketData:   marketData,
		ShipyardData: shipyardData,
		FirstVisited: time.Now(),
	})
}
