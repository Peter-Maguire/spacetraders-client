package entity

import (
	"fmt"
	"spacetraders/http"
	"strings"
)

type Waypoint string

func (w *Waypoint) GetSystemName() string {
	strw := string(*w)
	return strw[:strings.LastIndex(strw, "-")]
}

func (w *Waypoint) GetSystem() (*System, error) {
	return http.Request[System]("GET", fmt.Sprintf("systems/%s", w.GetSystemName()), nil)
}

func (w *Waypoint) GetSystemWaypoints() (*[]WaypointData, *http.HttpError) {
	return http.Request[[]WaypointData]("GET", fmt.Sprintf("systems/%s/waypoints", w.GetSystemName()), nil)
}

func (w *Waypoint) GetMarket() (*Market, error) {
	return http.Request[Market]("GET", fmt.Sprintf("systems/%s/waypoints/%s/market", w.GetSystemName(), *w), nil)
}

func (w *Waypoint) GetShipyard() (*ShipyardStock, *http.HttpError) {
	return http.Request[ShipyardStock]("GET", fmt.Sprintf("systems/%s/waypoints/%s/shipyard", w.GetSystemName(), *w), nil)
}

func (w *Waypoint) GetWaypointData() (*WaypointData, error) {
	return http.Request[WaypointData]("GET", fmt.Sprintf("systems/%s/waypoints/%s", w.GetSystemName(), *w), nil)
}
