package error

// internal error value
const (
	InternalErrorRequestTimeout = "internal error. request timeout"
)

const (
	MessageEmailBeingUsed     = "email address is being used"
	MessageUserIdBeingUsed    = "already registered user id"
	MessageOperationTimeout   = "operation timeout"
	MessageUnknownError       = "unknown error message"
	MessageQueryParamNotfound = "no query param"
	MessageBindFailed         = "bind failed"
	MessageInvalidAuthType    = "auth method not supported"
)

// Response status detail code
const (
	InternalError = -1

	QueryResultOk = 1
	// email check
	EmailCheckErrorBeingUsed = 2
	// userid check
	UserIdCheckErrorBeingUsed = 3
	// login or create account
	InvalidAuthType = 4
	UserNotFound    = 5

	ApiOperationRequestTimeout  = 300
	ApiOperationResponseTimeout = 301

	DatabaseOperationError = 1000
)
