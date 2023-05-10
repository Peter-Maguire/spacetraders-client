package entity

import "time"

type ExtractionResult struct {
	Extraction ExtractionData `json:"extraction"`
	Cooldown   Cooldown       `json:"cooldown"`
	Cargo      ShipCargo      `json:"cargo"`
}

type Cooldown struct {
	ShipSymbol       string    `json:"shipSymbol"`
	TotalSeconds     int       `json:"totalSeconds"`
	RemainingSeconds int       `json:"remainingSeconds"`
	Expiration       time.Time `json:"expiration"`
}

type ExtractionData struct {
	ShipSymbol string `json:"shipSymbol"`
	Yield      Yield  `json:"yield"`
}

type Yield struct {
	Symbol string `json:"symbol"`
	Units  int    `json:"units"`
}
