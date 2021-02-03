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
	Uid             string `json:"user_id"`           // 아이디
	ProfileImage    string `json:"profile_image"`     // 프로파일 이미지 base64 인코딩
	EmailAddress    string `json:"email"`             // email
	CellPhoneNumber string `json:"cell_phone_number"` // 전화번호
	DayOfBirth      string `json:"day_of_birth"`      // 생년월일
	AuthType        string `json:"auth"`              // 인증방법. google apple kakao
	Meta            string `json:"meta_json"`         // meta json field
}

// user 등록 응답
type RegisterUserResponse struct {
	BaseResponse
}

// user id 중복 체크 응답
type UserIdCheckResponse struct {
	BaseResponse
}
