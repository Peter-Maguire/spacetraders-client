package entity

import (
	"context"
	"fmt"
	"math"
	"spacetraders/constant"
	"spacetraders/http"
	"time"
)

type LimitedWaypointData struct {
	Symbol Waypoint              `json:"symbol"`
	Type   constant.WaypointType `json:"type"`
	X      int                   `json:"x"`
	Y      int                   `json:"y"`
}

func (lw *LimitedWaypointData) GetFullWaypoint(ctx context.Context) (*WaypointData, error) {
	return http.Request[WaypointData](ctx, "GET", fmt.Sprintf("systems/%s/waypoints/%s", lw.Symbol.GetSystemName(), lw.Symbol), nil)
}

func (lw *LimitedWaypointData) GetSystem(ctx context.Context) (*System, error) {
	return http.Request[System](ctx, "GET", fmt.Sprintf("systems/%s", lw.Symbol.GetSystemName()), nil)
}

// GetDistanceFrom formula stolen from https://canary.discord.com/channels/792864705139048469/852291054957887498/1109740523339657216
func (lw *LimitedWaypointData) GetDistanceFrom(lw2 LimitedWaypointData) int {
	return int(math.Sqrt(math.Pow(float64(lw.X-lw2.X), 2) + math.Pow(float64(lw.Y-lw2.Y), 2)))
}

type WaypointData struct {
	LimitedWaypointData
	SystemSymbol        string    `json:"systemSymbol"`
	Orbitals            []Orbital `json:"orbitals"`
	Traits              []Trait   `json:"traits"`
	Chart               Chart     `json:"chart"`
	Faction             Faction   `json:"faction"`
	IsUnderConstruction bool      `json:"isUnderConstruction"`
}

func (w *WaypointData) HasTrait(symbol constant.WaypointTrait) bool {
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
