package http

type ErrorCode int

const (
	ErrCooldown               ErrorCode = 4000
	ErrInsufficientFuelForNav ErrorCode = 4203
	ErrCargoFull              ErrorCode = 4228
)
