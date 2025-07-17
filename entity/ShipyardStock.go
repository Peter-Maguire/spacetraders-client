package entity

type ShipyardStock struct {
	Symbol           string              `json:"symbol"`
	ShipTypes        []ShipType          `json:"shipTypes"`
	Transactions     []MarketTransaction `json:"transactions"`
	Ships            []AvailableShip     `json:"ships"`
	ModificationsFee int                 `json:"modifications_fee"`
}

type ShipType struct {
	Type string `json:"type"`
}
