package entity

type ShipPurchaseResult struct {
	Ship        *Ship              `json:"ship"`
	Agent       *Agent             `json:"agent"`
	Transaction *MarketTransaction `json:"transaction"`
}
