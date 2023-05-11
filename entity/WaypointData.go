package entity

import (
	"fmt"
	"spacetraders/http"
	"time"
)

type LimitedWaypointData struct {
	Symbol Waypoint `json:"symbol"`
	Type   string   `json:"type"`
	X      int      `json:"x"`
	Y      int      `json:"y"`
}

func (lw *LimitedWaypointData) GetFullWaypoint() (*WaypointData, error) {
	return http.Request[WaypointData]("GET", fmt.Sprintf("systems/%s/waypoints/%s", lw.Symbol.GetSystemName(), lw.Symbol), nil)
}

func (lw *LimitedWaypointData) GetSystem() (*System, error) {
	return http.Request[System]("GET", fmt.Sprintf("systems/%s", lw.Symbol.GetSystemName()), nil)
}

type WaypointData struct {
	LimitedWaypointData
	SystemSymbol string    `json:"systemSymbol"`
	Orbitals     []Orbital `json:"orbitals"`
	Traits       []Trait   `json:"traits"`
	Chart        Chart     `json:"chart"`
	Faction      Faction   `json:"faction"`
}

func (w *WaypointData) HasTrait(symbol string) bool {
	for _, trait := range w.Traits {
		if trait.Symbol == symbol {
			return true
		}
	}
	return false
}

type Orbital struct {
	Symbol string `json:"symbol"`
}

type Faction struct {
	Symbol string `json:"symbol"`
}

type Chart struct {
	SubmittedBy string    `json:"submittedBy"`
	SubmittedOn time.Time `json:"submittedOn"`
}
