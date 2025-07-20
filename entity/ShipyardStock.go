package entity

type ShipyardStock struct {
	Symbol           string              `json:"symbol"`
	ShipTypes        []ShipType          `json:"shipTypes"`
	Transactions     []MarketTransaction `json:"transactions"`
	Ships            []AvailableShip     `json:"ships"`
	ModificationsFee int                 `json:"modifications_fee"`
}

func (ss *ShipyardStock) SellsShipType(shipType string) bool {
	for _, t := range ss.ShipTypes {
		if t.Type == shipType {
			return true
		}
	}
	return false
}

type ShipType struct {
	Type string `json:"type"`
}
