package entity

import "time"

type ExtractionResult struct {
	Extraction ExtractionData       `json:"extraction"`
	Cooldown   Cooldown             `json:"cooldown"`
	Cargo      ShipCargo            `json:"cargo"`
	Events     []ShipConditionEvent `json:"events"`
}

type SiphonResult struct {
	Siphon   ExtractionData       `json:"siphon"`
	Cooldown Cooldown             `json:"cooldown"`
	Cargo    ShipCargo            `json:"cargo"`
	Events   []ShipConditionEvent `json:"events"`
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

type ShipConditionEvent struct {
	Symbol      string `json:"symbol"`
	Component   string `json:"component"`
	Name        string `json:"name"`
	Description string `json:"description"`
}
