package entity

type System struct {
	Symbol       string                `json:"symbol"`
	SectorSymbol string                `json:"sectorSymbol"`
	Type         string                `json:"type"`
	X            int                   `json:"x"`
	Y            int                   `json:"y"`
	Waypoints    []LimitedWaypointData `json:"waypoints"`
	Factions     []interface{}         `json:"factions"`
}
