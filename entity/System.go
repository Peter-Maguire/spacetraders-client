package entity

import (
	"context"
	"fmt"
	"math"
	"spacetraders/constant"
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

func GetSystem(ctx context.Context, system string) (*System, *http.HttpError) {
	return http.Request[System](ctx, "GET", fmt.Sprintf("systems/%s", system), nil)
}

func (s *System) GetWaypoints(ctx context.Context) (*[]WaypointData, *http.HttpError) {
	return http.Request[[]WaypointData](ctx, "GET", fmt.Sprintf("systems/%s/waypoints", s.Symbol), nil)
}

func (s *System) GetLimitedWaypoint(ctx context.Context, name Waypoint) *LimitedWaypointData {
	for _, wp := range s.Waypoints {
		if wp.Symbol == name {
			return &wp
		}
	}
	return nil
}

func (s *System) GetJumpGate(ctx context.Context) *LimitedWaypointData {
	for _, wp := range s.Waypoints {
		if wp.Type == constant.WaypointTypeJumpGate {
			fullWaypoint, _ := wp.GetFullWaypoint(ctx)
			fmt.Println(fullWaypoint)
			if !fullWaypoint.IsUnderConstruction {
				return &wp
			}
		}
	}
	return nil
}

func (s *System) GetDistanceFrom(s2 *System) int {
	return int(math.Sqrt(math.Pow(float64(s.X-s2.X), 2) + math.Pow(float64(s.Y-s2.Y), 2)))
}
