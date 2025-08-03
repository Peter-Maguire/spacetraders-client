package http

type ErrorCode int

const (
	ErrResponseSerialization   ErrorCode = 3000
	ErrUnprocessableInput      ErrorCode = 3001
	ErrAllErrorHandlersFailed  ErrorCode = 3002
	ErrSystemStatusMaintenance ErrorCode = 3100
	ErrReset                   ErrorCode = 3200

	ErrCooldownConflict ErrorCode = 4000
	ErrWaypointNoAccess ErrorCode = 4001

	ErrTokenEmpty                       ErrorCode = 4100
	ErrTokenMissingSubject              ErrorCode = 4101
	ErrTokenInvalidSubject              ErrorCode = 4102
	ErrMissingTokenRequest              ErrorCode = 4103
	ErrInvalidTokenRequest              ErrorCode = 4104
	ErrInvalidTokenSubject              ErrorCode = 4105
	ErrAccountNotExists                 ErrorCode = 4106
	ErrAgentNotExists                   ErrorCode = 4107
	ErrAccountHasNoAgent                ErrorCode = 4108
	ErrTokenInvalidVersion              ErrorCode = 4109
	ErrRegisterAgentSymbolReserved      ErrorCode = 4110
	ErrRegisterAgentConflictSymbol      ErrorCode = 4111
	ErrRegisterAgentNoStartingLocations ErrorCode = 4112
	ErrTokenResetDateMismatch           ErrorCode = 4113
	ErrInvalidAccountRole               ErrorCode = 4114
	ErrInvalidToken                     ErrorCode = 4115
	ErrMissingAccountTokenRequest       ErrorCode = 4116

	ErrNavigateInTransit          ErrorCode = 4200
	ErrNavigateInvalidDestination ErrorCode = 4201
	ErrNavigateOutsideSystem      ErrorCode = 4202
	ErrNavigateInsufficientFuel   ErrorCode = 4203
	ErrNavigateSameDestination    ErrorCode = 4204
	ErrShipExtractInvalidWaypoint ErrorCode = 4205
	ErrShipExtractPermission      ErrorCode = 4206
	// ErrInsufficientAntimatter doesn't seem to exist anymore
	ErrInsufficientAntimatter               ErrorCode = 4212
	ErrShipInTransit                        ErrorCode = 4214
	ErrShipMissingSensorArrays              ErrorCode = 4215
	ErrPurchaseShipCredits                  ErrorCode = 4216
	ErrShipCargoExceedsLimit                ErrorCode = 4217
	ErrShipCargoMissing                     ErrorCode = 4218
	ErrShipCargoUnitCount                   ErrorCode = 4219
	ErrShipSurveyVerification               ErrorCode = 4220
	ErrShipSurveyExpiration                 ErrorCode = 4221
	ErrShipSurveyWaypointType               ErrorCode = 4222
	ErrShipSurveyOrbit                      ErrorCode = 4223
	ErrShipSurveyExhausted                  ErrorCode = 4224
	ErrShipCargoFull                        ErrorCode = 4228
	ErrWaypointCharted                      ErrorCode = 4230
	ErrShipTransferShipNotFound             ErrorCode = 4231
	ErrShipTransferAgentConflict            ErrorCode = 4232
	ErrShipTransferSameShipConflict         ErrorCode = 4233
	ErrShipTransferLocationConflict         ErrorCode = 4234
	ErrWarpInsideSystem                     ErrorCode = 4235
	ErrShipNotInOrbit                       ErrorCode = 4236
	ErrShipInvalidRefineryGood              ErrorCode = 4237
	ErrShipInvalidRefineryType              ErrorCode = 4238
	ErrShipMissingRefinery                  ErrorCode = 4239
	ErrShipMissingSurveyor                  ErrorCode = 4240
	ErrShipMissingWarpDrive                 ErrorCode = 4241
	ErrShipMissingMineralProcessor          ErrorCode = 4242
	ErrShipMissingMiningLasers              ErrorCode = 4243
	ErrShipNotDocked                        ErrorCode = 4244
	ErrPurchaseShipNotPresent               ErrorCode = 4245
	ErrShipMountNoShipyard                  ErrorCode = 4246
	ErrShipMissingMount                     ErrorCode = 4247
	ErrShipMountInsufficientCredits         ErrorCode = 4248
	ErrShipMissingPower                     ErrorCode = 4249
	ErrShipMissingSlots                     ErrorCode = 4250
	ErrShipMissingMounts                    ErrorCode = 4251
	ErrShipMissingCrew                      ErrorCode = 4252
	ErrShipExtractDestabilized              ErrorCode = 4253
	ErrShipJumpInvalidOrigin                ErrorCode = 4254
	ErrShipJumpInvalidWaypoint              ErrorCode = 4255
	ErrShipJumpOriginUnderConstruction      ErrorCode = 4256
	ErrShipMissingGasProcessor              ErrorCode = 4257
	ErrShipMissingGasSiphons                ErrorCode = 4258
	ErrShipSiphonInvalidWaypoint            ErrorCode = 4259
	ErrShipSiphonPermission                 ErrorCode = 4260
	ErrWaypointNoYield                      ErrorCode = 4261
	ErrShipJumpDestinationUnderConstruction ErrorCode = 4262
	ErrShipScrapInvalidTrait                ErrorCode = 4263
	ErrShipRepairInvalidTrait               ErrorCode = 4264
	ErrAgentInsufficientCredits             ErrorCode = 4265
	ErrShipModuleNoShipyard                 ErrorCode = 4266
	ErrShipModuleNotInstalled               ErrorCode = 4267
	ErrShipModuleInsufficientCredits        ErrorCode = 4268
	ErrCantSlowDownWhileInTransit           ErrorCode = 4269
	ErrShipExtractInvalidSurveyLocation     ErrorCode = 4270
	ErrShipTransferDockedOrbitConflict      ErrorCode = 4271

	ErrAcceptContractNotAuthorized ErrorCode = 4500
	ErrAcceptContractConflict      ErrorCode = 4501
	ErrFulfillContractDelivery     ErrorCode = 4502
	ErrContractDeadline            ErrorCode = 4503
	ErrContractFulfilled           ErrorCode = 4504
	ErrContractNotAccepted         ErrorCode = 4505
	ErrContractNotAuthorized       ErrorCode = 4506
	ErrShipDeliverTerms            ErrorCode = 4508
	ErrShipDeliverFulfilled        ErrorCode = 4509
	ErrShipDeliverInvalidLocation  ErrorCode = 4510
	ErrExistingContract            ErrorCode = 4511

	ErrMarketTradeInsufficientCredits ErrorCode = 4600
	ErrMarketTradeNoPurchase          ErrorCode = 4601
	ErrMarketTradeNotSold             ErrorCode = 4602
	ErrMarketTradeUnitLimit           ErrorCode = 4604
	ErrShipNotAvailableForPurchase    ErrorCode = 4605

	ErrWaypointNoFaction ErrorCode = 4700

	ErrConstructionMaterialNotRequired ErrorCode = 4800
	ErrConstructionMaterialFulfilled   ErrorCode = 4801
	ErrShipConstructionInvalidLocation ErrorCode = 4802

	ErrUnsupportedMediaType ErrorCode = 5000
)
