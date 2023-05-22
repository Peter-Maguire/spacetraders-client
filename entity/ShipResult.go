package entity

type ShipChartResult struct {
	Chart    *Chart        `json:"chart"`
	Waypoint *WaypointData `json:"waypoint"`
}

type ShipJumpResult struct {
	Cooldown *Cooldown `json:"cooldown"`
	Nav      ShipNav   `json:"nav"`
}

type ShipPurchaseResult struct {
	Ship        *Ship              `json:"ship"`
	Agent       *Agent             `json:"agent"`
	Transaction *MarketTransaction `json:"transaction"`
}

type ShipWarpResult struct {
	Fuel ShipFuel `json:"fuel"`
	Nav  ShipNav  `json:"nav"`
}

type ShipScanWaypointsResult struct {
	Cooldown  *Cooldown       `json:"cooldown"`
	Waypoints *[]WaypointData `json:"waypoints"`
}

type ShipScanSystemsResult struct {
	Cooldown *Cooldown `json:"cooldown"`
	Systems  *[]System `json:"systems"`
}

type ItemPurchaseResult struct {
	Agent       *Agent             `json:"agent"`
	Cargo       *ShipCargo         `json:"cargo"`
	Transaction *MarketTransaction `json:"transaction"`
}
