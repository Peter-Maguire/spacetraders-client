package entity

type ShipyardStock struct {
	Symbol       string              `json:"symbol"`
	ShipTypes    []ShipType          `json:"ShipTypes"`
	Transactions []MarketTransaction `json:"transactions"`
	Ships        []AvailableShip     `json:"ships"`
}

type ShipType struct {
	Type string `json:"type"`
}
