package entity

import "time"

type WaypointData struct {
	SystemSymbol string    `json:"systemSymbol"`
	Symbol       Waypoint  `json:"symbol"`
	Type         string    `json:"type"`
	X            int       `json:"x"`
	Y            int       `json:"y"`
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
