package entity

import (
	"context"
	"fmt"
	"spacetraders/http"
	"strings"
)

type Waypoint string

func (w *Waypoint) GetSystemName() string {
	strw := string(*w)
	return strw[:strings.LastIndex(strw, "-")]
}

func (w *Waypoint) GetSystem(ctx context.Context) (*System, *http.HttpError) {
	return http.Request[System](ctx, "GET", fmt.Sprintf("systems/%s", w.GetSystemName()), nil)
}

func (w *Waypoint) GetSystemWaypoints(ctx context.Context) (*[]WaypointData, *http.HttpError) {
	return http.PaginatedRequest[WaypointData](ctx, fmt.Sprintf("systems/%s/waypoints", w.GetSystemName()), 1, 0)
}

func (w *Waypoint) GetMarket(ctx context.Context) (*Market, *http.HttpError) {
	return http.Request[Market](ctx, "GET", fmt.Sprintf("systems/%s/waypoints/%s/market", w.GetSystemName(), *w), nil)
}

func (w *Waypoint) GetShipyard(ctx context.Context) (*ShipyardStock, *http.HttpError) {
	return http.Request[ShipyardStock](ctx, "GET", fmt.Sprintf("systems/%s/waypoints/%s/shipyard", w.GetSystemName(), *w), nil)
}

func (w *Waypoint) GetWaypointData(ctx context.Context) (*WaypointData, *http.HttpError) {
	return http.Request[WaypointData](ctx, "GET", fmt.Sprintf("systems/%s/waypoints/%s", w.GetSystemName(), *w), nil)
}

func (w *Waypoint) GetConstructionSite(ctx context.Context) (*ConstructionSite, *http.HttpError) {
	return http.Request[ConstructionSite](ctx, "GET", fmt.Sprintf("systems/%s/waypoints/%s/construction", w.GetSystemName(), *w), nil)
}
