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

// user 등록 응답
type RegisterUserResponse struct {
	Token string `json:"token"` //로그인  or 회원가입 성공시 발급되는 토큰. redis에서 관리되며 12시간이후 expire 된다.
	BaseResponse
}

// user id 중복 체크 응답
type UserIdCheckResponse struct {
	BaseResponse
}

type NonFirebaseAuthRequest struct {
	UniqueId string `json:"unique_id"` // `unique_id`: unique_id 한 값. 외부업체 로그인에만 쓰임
}

type NonFirebaseAuthResponse struct {
	BaseResponse
	SignedIn    bool   `json:"signed_in"`    // 이미 가입된 사용자여부
	CustomToken string `json:"custom_token"` // firebase auth를 위한 custom token 값
	Email       string `json:"email"`        // 가입시 사용한 email
}

type FirebaseAuthRequest struct {
	IdToken string `json:"firebase_id_token"` // id token
}

type FirebaseAuthResponse struct {
	BaseResponse
	SignedIn bool   `json:"signed_in"` // 이미 가입된 사용자여부
	Email    string `json:"email"`     // 가입시 사용한 email
}

type LoginRequest struct {
	IdToken      string `json:"firebase_id_token"` // id token
	EmailAddress string `json:"email"`             // email 주소
}

type LoginResponse struct {
	BaseResponse
	SessionToken string `json:"session_token"` // 로그인이 된 경우 server session token 값
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

type SellerAuthResponse struct {
	BaseResponse
}

type ProductPostResponse struct {
	BaseResponse
}

type CartItemAddRequest struct {
	UniqueId     string // `unique_id`: firebase token id 혹은 email address
	Token        string // `token`: login 혹은 user 등록 요청의 응답으로 받은 토큰
	ProductId    string // `product_id`: 카트에 담을 제품의 product id
	SelectedJson string // `selected_json`: 구매할 옵션 리스트
}

type CartItemAddResponse struct {
	BaseResponse
}

type CartItemRemoveRequest struct {
	UniqueId string // `unique_id`: firebase token id 혹은 email address
	Token    string // `token`: login 혹은 user 등록 요청의 응답으로 받은 토큰
	CartId   string // `cart_id`: 삭제할 cart 아이템의 id
}

type CartItemRemoveResponse struct {
	BaseResponse
}

type ReviewPostResponse struct {
	BaseResponse
}
