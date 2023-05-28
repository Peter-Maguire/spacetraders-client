package entity

import (
	"fmt"
	"math"
	"spacetraders/http"
)

type System struct {
	Symbol       string                `json:"symbol"`
	SectorSymbol string                `json:"sectorSymbol"`
	Type         string                `json:"type"`
	X            int                   `json:"x"`
	Y            int                   `json:"y"`
	Distance     int                   `json:"distance"`
	Waypoints    []LimitedWaypointData `json:"waypoints"`
	Factions     []interface{}         `json:"factions"`
}

func (s *System) GetWaypoints() (*[]WaypointData, *http.HttpError) {
	return http.Request[[]WaypointData]("GET", fmt.Sprintf("systems/%s/waypoints", s.Symbol), nil)
}

func (s *System) GetLimitedWaypoint(name Waypoint) *LimitedWaypointData {
	for _, wp := range s.Waypoints {
		if wp.Symbol == name {
			return &wp
		}
	}
	return nil
}

func (s *System) GetDistanceFrom(s2 *System) int {
	return int(math.Sqrt(math.Pow(float64(s.X-s2.X), 2) + math.Pow(float64(s.Y-s2.Y), 2)))
}
