package entity

type SellRequest struct {
	Symbol string `json:"symbol"`
	Units  int    `json:"units"`
}
