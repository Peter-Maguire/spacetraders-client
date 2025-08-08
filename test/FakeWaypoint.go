package test

import (
	"encoding/json"
	"fmt"
	"spacetraders/constant"
	"spacetraders/database"
	"spacetraders/entity"
	"time"
)

func CreateFakeWaypointData(x int, y int, traits []constant.WaypointTrait) *entity.WaypointData {
	waypointTraits := make([]entity.Trait, len(traits))
	for i, trait := range traits {
		waypointTraits[i] = entity.Trait{
			Symbol:      trait,
			Name:        string(trait),
			Description: string(trait),
		}
	}

	return &entity.WaypointData{
		LimitedWaypointData: entity.LimitedWaypointData{
			Symbol: entity.Waypoint(fmt.Sprintf("FAKEWAYPOINT-%d-%d", x, y)),
			Type:   "ASTEROID",
			X:      x,
			Y:      y,
		},
		SystemSymbol:        "FAKESYSTEM",
		Orbitals:            nil,
		Traits:              waypointTraits,
		Chart:               entity.Chart{},
		Faction:             entity.Faction{},
		IsUnderConstruction: false,
	}
}

func CreateFakeDatabaseWaypoint(wp *entity.WaypointData, market *entity.Market, shipyard *entity.ShipyardStock) *database.Waypoint {

	wpData, _ := json.Marshal(wp)
	syData, _ := json.Marshal(wp)
	mkData, _ := json.Marshal(wp)

	return &database.Waypoint{
		Waypoint:     string(wp.Symbol),
		System:       string(wp.SystemSymbol),
		Data:         wpData,
		MarketData:   mkData,
		ShipyardData: syData,
		FirstVisited: time.Now(),
		TimesVisited: 1,
	}
}
