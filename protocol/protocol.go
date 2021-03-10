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

// formValue
type SellerAuthRequest struct {
	UniqueId                  string // `unique_id`: firebase token id 혹은 email address
	Token                     string // `token`: login 혹은 user 등록 요청의 응답으로 받은 토큰.
	SellerType                int    // `seller_type`: 개인회원 0, 기업회원 1
	CompanyRegistrationNumber string // `company_registration_number`: 사업자 등록번호.
	CompanyOwnerName          string // `owner_name`: 사업주 이름
	CompanyName               string // `company_name`: 기업체 이름
	ChannelName               string // `channel_name`: 채널 이름
	ChannelUrl                string // `channel_url`: 채널 url
	ChannelDescription        string // `channel_description`: 채널 설명
	BankName                  string // `bank_name`: 은행 이름
	BankAccountNumber         string // `bank_account_number`: 계좌번호
}

type SellerAuthResponse struct {
	BaseResponse
}

type VideoInfo struct {
	Order int    `json:"order"` // 순서
	Id    string `json:"id"`    // unique id
	Url   string `json:"url"`   // 업로드 된
}

type Category struct {
	Grade int    `json:"grade"`
	Name  string `json:"name"`
}

type PostProductRequest struct {
	UniqueId        string // `unique_id`: firebase token id 혹은 email address
	Token           string // `token`: login 혹은 user 등록 요청의 응답으로 받은 토큰.
	Title           string // `title`: 등록 상품 제목
	VideoIdListJson string // video id list json.
	CategoryJson    string // `category_info_json`: 카테고리정보
	BasePrice       int    // `base_price`: 기본 금액
	BaseAmount      int    // `base_amount`: 기본 개수
	OptionJson      string // `option_json`: 옵션
}

type PostProductResponse struct {
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
