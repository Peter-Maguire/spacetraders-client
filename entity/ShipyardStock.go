package entity

type ShipyardStock struct {
	Symbol    string     `json:"symbol"`
	ShipTypes []ShipType `json:"ShipTypes"`
}

type ShipType struct {
	Type string `json:"type"`
}
