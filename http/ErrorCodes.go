package http

type ErrorCode int

const (
	ErrCooldown               ErrorCode = 4000
	ErrNavigateInTransit      ErrorCode = 4200
	ErrInsufficientFuelForNav ErrorCode = 4203
	ErrCargoFull              ErrorCode = 4228
	ErrShipAtDestination      ErrorCode = 4204
	ErrCannotExtractHere      ErrorCode = 4205
	ErrInsufficientAntimatter ErrorCode = 4212
	ErrShipInTransit          ErrorCode = 4214

	ErrShipSurveyVerification ErrorCode = 4220

	ErrShipSurveyExpired   ErrorCode = 4221
	ErrShipSurveyExhausted ErrorCode = 4224
	ErrInsufficientFunds   ErrorCode = 4600
	ErrNoFactionPresence   ErrorCode = 4700
)
