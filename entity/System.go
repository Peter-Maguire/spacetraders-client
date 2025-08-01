package entity

import (
	"context"
	"fmt"
	"math"
	"spacetraders/constant"
	"spacetraders/http"
)

type System struct {
	Symbol       SystemSymbol          `json:"symbol"`
	SectorSymbol string                `json:"sectorSymbol"`
	Type         string                `json:"type"`
	X            int                   `json:"x"`
	Y            int                   `json:"y"`
	Distance     int                   `json:"distance"`
	Waypoints    []LimitedWaypointData `json:"waypoints"`
	Factions     []interface{}         `json:"factions"`
}

func GetSystem(ctx context.Context, system SystemSymbol) (*System, *http.HttpError) {
	return system.GetSystem(ctx)
}

func (s *System) GetWaypoints(ctx context.Context) (*[]WaypointData, *http.HttpError) {
	return s.Symbol.GetWaypoints(ctx)
}

func (s *System) GetWaypointsWithTrait(ctx context.Context, trait constant.WaypointTrait) (*[]WaypointData, *http.HttpError) {
	return s.Symbol.GetWaypointsWithTrait(ctx, trait)
}

func (s *System) GetWaypointsOfType(ctx context.Context, waypointType constant.WaypointType) (*[]WaypointData, *http.HttpError) {
	return s.Symbol.GetWaypointsOfType(ctx, waypointType)
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
