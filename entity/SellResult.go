package entity

type SellResult struct {
	Agent       Agent             `json:"agent"`
	Cargo       ShipCargo         `json:"cargo"`
	Transaction MarketTransaction `json:"transaction"`
}
