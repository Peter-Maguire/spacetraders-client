package http

type ErrorCode int

const (
	ErrCooldown               ErrorCode = 4000
	ErrNavigateInTransit      ErrorCode = 4200
	ErrInsufficientFuelForNav ErrorCode = 4203
	ErrCargoFull              ErrorCode = 4228
	ErrShipAtDestination      ErrorCode = 4204
	ErrCannotExtractHere      ErrorCode = 4205
	ErrAlreadyInSystem        ErrorCode = 4208
	ErrInsufficientAntimatter ErrorCode = 4212

	ErrPurchaseShipInsufficientFunds ErrorCode = 4216
	ErrTransferCargoFull             ErrorCode = 4217
	ErrCargoUnitCountError           ErrorCode = 4219
	ErrShipInTransit                 ErrorCode = 4214

	ErrShipSurveyVerification ErrorCode = 4220

	ErrShipSurveyExpired         ErrorCode = 4221
	ErrShipSurveyExhausted       ErrorCode = 4224
	ErrNoWarpDrive               ErrorCode = 4241
	ErrOverExtracted             ErrorCode = 4253
	ErrJumpGateUnderConstruction ErrorCode = 4256

	ErrContractTermsMet ErrorCode = 4509

	ErrInsufficientFunds ErrorCode = 4600
	ErrNoFactionPresence ErrorCode = 4700
)
