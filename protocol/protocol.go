package protocol

type Code int
type BaseResponse struct {
	Status Code   `json:"status"`
	Detail string `json:"detail"`
}

// email 사용중인지 여부 체크 응답
type EmailCheckResponse struct {
	BaseResponse
}

// user 등록 요청
type RegisterUserRequest struct {
	UniqueId        string `json:"unique_id"`         // firebase token id 혹은 email address
	UserId          string `json:"user_id"`           // 아이디
	ProfileImage    string `json:"profile_image"`     // 프로파일 이미지 base64 인코딩
	EmailAddress    string `json:"email"`             // email
	CellPhoneNumber string `json:"cell_phone_number"` // 전화번호
	DayOfBirth      string `json:"day_of_birth"`      // 생년월일
	AuthType        string `json:"auth"`              // 인증방법. google|apple|kakao|email
	Meta            string `json:"meta_json"`         // meta json field
}

// user 등록 응답
type RegisterUserResponse struct {
	Token string `json:"token"` //로그인  or 회원가입 성공시 발급되는 토큰. redis에서 관리되며 12시간이후 expire 된다.
	BaseResponse
}

// user id 중복 체크 응답
type UserIdCheckResponse struct {
	BaseResponse
}

type LoginRequest struct {
	UniqueId string `json:"unique_id"` // firebase token id 혹은 email address
	Password string `json:"password"`  // md5 로 인코딩된 비밀번호. firebase token 인 경우 무시
	AuthType string `json:"auth"`      // 인증방법. google|apple|kakao|email
}

type LoginResponse struct {
	Token string `json:"token"` //로그인  or 회원가입 성공시 발급되는 토큰. redis에서 관리되며 12시간이후 expire 된다.
	BaseResponse
}
