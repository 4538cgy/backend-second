package user

import (
	"fmt"
	"github.com/4538cgy/backend-second/api/context"
	"github.com/4538cgy/backend-second/api/route"
	"github.com/4538cgy/backend-second/config"
	"github.com/4538cgy/backend-second/database"
	vcomError "github.com/4538cgy/backend-second/error"
	"github.com/4538cgy/backend-second/log"
	"github.com/4538cgy/backend-second/protocol"
	"github.com/labstack/echo/v4"
	"net/http"
	"time"
)

const (
	// 구글이 아닌 외부업체의 경우 customToken 요청.
	// 요청으로 보낸 unique_id 를 uid 로 firebase의 customToken을 생성해준다.
	// 만약 unique_id 로 회원가입이 이미 되어있는 경우 로그인을 시키고
	// Session Token 을 발행한다.
	loginForNonFirebaseUrl = "/api/user/auth/nfb"
	// 구글 인증인 경우 가입 여부 체크 요청
	// idToken 을 인자로 받으면 VerifyIDToken 을 통해 uid 를 추출하고
	// 해당 uid 로 회원가입이 이미 되어있으면 로그인 시키고
	// Session Token 을 발행한다.
	loginForFirebaseUrl = "/api/user/auth/fb"
	// 실제 로그인 요청
	loginUrl = "/api/user/auth"
)

func init() {
	route.AddRoute(route.NewRouteType(loginForNonFirebaseUrl, "POST"), loginForNonFirebase)
	route.AddRoute(route.NewRouteType(loginForFirebaseUrl, "POST"), loginForFirebase)
	route.AddRoute(route.NewRouteType(loginUrl, "POST"), login)
	route.AddRoute(route.NewRouteType(loginForNonFirebaseUrl, "DELETE"), logout)
}

// 구글 인증이 되어있는 경우, TokenId 를 받아서 uid를 뽑아냄.
// 외부 인증인 경우 외부인증에서 사용하는 uid 를 받아서 customToken 을 생성함.
//
func loginForNonFirebase(ctx echo.Context) error {
	resp := &protocol.NonFirebaseAuthResponse{}
	customContext, ok := ctx.(*context.CustomContext)
	if !ok {
		log.Error("failed to casting echo.Context to api.CustomContext")
		resp.Status = vcomError.InternalError
		resp.Detail = vcomError.MessageUnknownError
		return ctx.JSON(http.StatusInternalServerError, resp)
	}

	// bind requested data
	nonFirebaseLoginRequest := &protocol.NonFirebaseAuthRequest{}
	err := ctx.Bind(nonFirebaseLoginRequest)
	if err != nil {
		log.Error("failed to bind register user request")
		resp.Status = vcomError.InternalError
		resp.Detail = vcomError.MessageBindFailed
		return ctx.JSON(http.StatusInternalServerError, resp)
	}

	// query user is already registered
	timer := time.NewTimer(time.Duration(config.Get().Api.HandleTimeoutMS) * time.Second)
	defer timer.Stop()

	// send query to Database
	result := make(chan database.SelectQueryResult)
	select {
	case customContext.SelectQueryWritePump() <- database.NewSelectTransaction(
		fmt.Sprintf("SELECT email FROM vcommerce.user WHERE user_id='%s'", nonFirebaseLoginRequest.UniqueId),
		result,
	):
	case <-timer.C:
		resp.Status = vcomError.ApiOperationRequestTimeout
		resp.Detail = vcomError.MessageOperationTimeout
		return ctx.JSON(http.StatusInternalServerError, resp)
	}

	signedInUser := false
	var email string
	// receive result
	select {
	case res := <-result:
		if res.Err != nil {
			resp.Status = vcomError.DatabaseOperationError
			resp.Detail = res.Err.Error()
			return ctx.JSON(http.StatusInternalServerError, resp)
		}
		if res.Rows.Next() {
			signedInUser = true
			res.Rows.Scan(&email)
		}
	case <-timer.C:
		resp.Status = vcomError.ApiOperationResponseTimeout
		resp.Detail = vcomError.MessageOperationTimeout
		return ctx.JSON(http.StatusInternalServerError, resp)
	}

	if signedInUser {
		log.Info("Already registered user.")
		resp.SignedIn = true
		resp.Status = vcomError.QueryResultOk
		resp.Email = email
		return ctx.JSON(http.StatusOK, resp)
	}
	token, err := customContext.CreateCustomToken(nonFirebaseLoginRequest.UniqueId)
	if err != nil {
		msg := fmt.Sprintf("firebase customtoken failed. %s", err)
		resp.Status = vcomError.FirebaseTokenCreateFailed
		resp.Detail = msg
		return ctx.JSON(http.StatusInternalServerError, resp)
	}

	resp.Status = vcomError.QueryResultOk
	resp.CustomToken = token
	return ctx.JSON(http.StatusOK, resp)
}

func loginForFirebase(ctx echo.Context) error {
	resp := &protocol.FirebaseAuthResponse{}
	customContext, ok := ctx.(*context.CustomContext)
	if !ok {
		log.Error("failed to casting echo.Context to api.CustomContext")
		resp.Status = vcomError.InternalError
		resp.Detail = vcomError.MessageUnknownError
		return ctx.JSON(http.StatusInternalServerError, resp)
	}

	// bind requested data
	firebaseLoginRequest := &protocol.FirebaseAuthRequest{}
	err := ctx.Bind(firebaseLoginRequest)
	if err != nil {
		log.Error("failed to bind register user request")
		resp.Status = vcomError.InternalError
		resp.Detail = vcomError.MessageBindFailed
		return ctx.JSON(http.StatusInternalServerError, resp)
	}

	uid, err := customContext.VerifyIDToken(firebaseLoginRequest.IdToken)
	if err != nil {
		msg := fmt.Sprintf("firebase verify failed. %s", err)
		resp.Status = vcomError.FirebaseVerifyTokenFailed
		resp.Detail = msg
		return ctx.JSON(http.StatusInternalServerError, resp)
	}

	// query user is already registered
	timer := time.NewTimer(time.Duration(config.Get().Api.HandleTimeoutMS) * time.Second)
	defer timer.Stop()

	// send query to Database
	result := make(chan database.SelectQueryResult)
	select {
	case customContext.SelectQueryWritePump() <- database.NewSelectTransaction(
		fmt.Sprintf("SELECT email FROM vcommerce.user WHERE user_id='%s'", uid),
		result,
	):
	case <-timer.C:
		resp.Status = vcomError.ApiOperationRequestTimeout
		resp.Detail = vcomError.MessageOperationTimeout
		return ctx.JSON(http.StatusInternalServerError, resp)
	}

	signedInUser := false
	var email string
	// receive result
	select {
	case res := <-result:
		if res.Err != nil {
			resp.Status = vcomError.DatabaseOperationError
			resp.Detail = res.Err.Error()
			return ctx.JSON(http.StatusInternalServerError, resp)
		}
		if res.Rows.Next() {
			signedInUser = true
			res.Rows.Scan(&email)
		}
	case <-timer.C:
		resp.Status = vcomError.ApiOperationResponseTimeout
		resp.Detail = vcomError.MessageOperationTimeout
		resp.Email = email
		return ctx.JSON(http.StatusInternalServerError, resp)
	}

	if signedInUser {
		log.Info("Already registered user.")
		resp.SignedIn = true
		resp.Status = vcomError.QueryResultOk
		return ctx.JSON(http.StatusOK, resp)
	}

	resp.Status = vcomError.QueryResultOk
	resp.SignedIn = false // not registered yet.
	return ctx.JSON(http.StatusOK, resp)
}

func login(ctx echo.Context) error {
	resp := &protocol.LoginResponse{}
	customContext, ok := ctx.(*context.CustomContext)
	if !ok {
		log.Error("failed to casting echo.Context to api.CustomContext")
		resp.Status = vcomError.InternalError
		resp.Detail = vcomError.MessageUnknownError
		return ctx.JSON(http.StatusInternalServerError, resp)
	}

	// bind requested data
	loginRequest := &protocol.LoginRequest{}
	err := ctx.Bind(loginRequest)
	if err != nil {
		log.Error("failed to bind register user request")
		resp.Status = vcomError.InternalError
		resp.Detail = vcomError.MessageBindFailed
		return ctx.JSON(http.StatusInternalServerError, resp)
	}

	uid, userEmail, err := customContext.GetUserEmail(loginRequest.IdToken)
	if err != nil {
		msg := fmt.Sprintf("firebase GetUserInfo failed. err: %s", err)
		resp.Status = vcomError.FirebaseUserInfoFailed
		resp.Detail = msg
		return ctx.JSON(http.StatusInternalServerError, resp)
	}

	if userEmail != loginRequest.EmailAddress {
		msg := fmt.Sprintf("firebase Validation failed. err: %s", err)
		resp.Status = vcomError.FirebaseUserInfoFailed
		resp.Detail = msg
		return ctx.JSON(http.StatusInternalServerError, resp)
	}

	// query user is already registered
	timer := time.NewTimer(time.Duration(config.Get().Api.HandleTimeoutMS) * time.Second)
	defer timer.Stop()

	// send query to Database
	result := make(chan database.SelectQueryResult)
	select {
	case customContext.SelectQueryWritePump() <- database.NewSelectTransaction(
		fmt.Sprintf("SELECT user_id FROM vcommerce.user WHERE user_id='%s'", uid),
		result,
	):
	case <-timer.C:
		resp.Status = vcomError.ApiOperationRequestTimeout
		resp.Detail = vcomError.MessageOperationTimeout
		return ctx.JSON(http.StatusInternalServerError, resp)
	}

	signedInUser := false
	// receive result
	select {
	case res := <-result:
		if res.Err != nil {
			resp.Status = vcomError.DatabaseOperationError
			resp.Detail = res.Err.Error()
			return ctx.JSON(http.StatusInternalServerError, resp)
		}
		if res.Rows.Next() {
			signedInUser = true
		}
	case <-timer.C:
		resp.Status = vcomError.ApiOperationResponseTimeout
		resp.Detail = vcomError.MessageOperationTimeout
		return ctx.JSON(http.StatusInternalServerError, resp)
	}

	if !signedInUser {
		resp.Status = vcomError.UserNotFound
		resp.Detail = vcomError.MessageUserNotRegistered
		return ctx.JSON(http.StatusInternalServerError, resp)
	}

	serverSessionToken, err := customContext.InsertSession(uid, timer)
	if err != nil {
		// TODO rollback 해야하는지 여부 확인
		resp.Status = vcomError.SessionInsertionFailed
		resp.Detail = vcomError.MessageIOFailed
	}
	log.Info("Already registered user.")
	resp.SessionToken = serverSessionToken
	resp.Status = vcomError.QueryResultOk
	return ctx.JSON(http.StatusOK, resp)
}

func logout(ctx echo.Context) error {

	return nil
}
