package entity

import (
	"context"
	"fmt"
	"spacetraders/constant"
	"spacetraders/http"
)

type SystemSymbol string

func (s SystemSymbol) GetSystem(ctx context.Context) (*System, *http.HttpError) {
	return http.Request[System](ctx, "GET", fmt.Sprintf("systems/%s", s), nil)
}

func (s SystemSymbol) GetWaypoints(ctx context.Context) (*[]WaypointData, *http.HttpError) {
	return http.PaginatedRequest[WaypointData](ctx, fmt.Sprintf("systems/%s/waypoints", s), 1, 0)
}

func (s SystemSymbol) GetWaypointsWithTrait(ctx context.Context, trait constant.WaypointTrait) (*[]WaypointData, *http.HttpError) {
	return http.PaginatedRequest[WaypointData](ctx, fmt.Sprintf("systems/%s/waypoints?traits=%s", s, trait), 1, 0)
}

func (s SystemSymbol) GetWaypointsOfType(ctx context.Context, waypointType constant.WaypointType) (*[]WaypointData, *http.HttpError) {
	return http.PaginatedRequest[WaypointData](ctx, fmt.Sprintf("systems/%s/waypoints?type=%s", s, waypointType), 1, 0)
}
