package entity

type ShipPurchaseRequest struct {
	ShipType       string   `json:"shipType"`
	WaypointSymbol Waypoint `json:"waypointSymbol"`
}
