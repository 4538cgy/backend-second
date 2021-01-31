package error

const (
	MessageEmailBeingUsed     = "email address is being used"
	MessageOperationTimeout   = "operation timeout"
	MessageUnknownError       = "unknown error message"
	MessageQueryParamNotfound = "no query param"
	MessageBindFailed         = "bind failed"
)

// Response status detail code
const (
	InternalError = -1

	QueryResultOk            = 1
	EmailCheckErrorBeingUsed = 2

	ApiOperationRequestTimeout  = 300
	ApiOperationResponseTimeout = 301

	DatabaseOperationError = 1000
	DatabaseResultError    = 1001
)
