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
	MessageUserNotRegistered  = "not registered user"
	MessageIOFailed           = "I/O failed"
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
	InvalidAuthType = 5
	UserNotFound    = 6

	SessionInsertionFailed = 100

	ApiOperationRequestTimeout  = 300
	ApiOperationResponseTimeout = 301

	DatabaseOperationError = 1000

	FirebaseTokenCreateFailed = 2000
	FirebaseVerifyTokenFailed = 2001
	FirebaseUserInfoFailed    = 2002
)
