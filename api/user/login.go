package user

import (
	context2 "context"
	"fmt"
	"github.com/4538cgy/backend-second/api/context"
	"github.com/4538cgy/backend-second/api/route"
	"github.com/4538cgy/backend-second/config"
	"github.com/4538cgy/backend-second/database"
	vcomError "github.com/4538cgy/backend-second/error"
	"github.com/4538cgy/backend-second/log"
	"github.com/4538cgy/backend-second/protocol"
	"github.com/4538cgy/backend-second/query"
	"github.com/4538cgy/backend-second/util"
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
		fmt.Sprintf("SELECT user_id FROM vcommerce.user WHERE user_id='%s'", nonFirebaseLoginRequest.UniqueId),
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

	if signedInUser {
		log.Info("Already registered user.")
		resp.SignedIn = true
		resp.Status = vcomError.QueryResultOk
		return ctx.JSON(http.StatusOK, resp)
	}

	client, err := customContext.App().Auth(context2.Background())
	if err != nil {
		msg := fmt.Sprintf("firebase auth failed. %s", err)
		resp.Status = vcomError.FirebaseTokenCreateFailed
		resp.Detail = msg
		return ctx.JSON(http.StatusInternalServerError, resp)
	}

	ctx2 := context2.Background()
	token, err := client.CustomToken(ctx2, nonFirebaseLoginRequest.UniqueId)
	if err != nil {
		msg := fmt.Sprintf("firebase customToken get failed. %s", err)
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

	client, err := customContext.App().Auth(context2.Background())
	if err != nil {
		msg := fmt.Sprintf("firebase auth failed. %s", err)
		resp.Status = vcomError.FirebaseAuthFailed
		resp.Detail = msg
		return ctx.JSON(http.StatusInternalServerError, resp)
	}

	token, err := client.VerifyIDToken(context2.Background(), firebaseLoginRequest.IdToken)
	if err != nil {
		msg := fmt.Sprintf("firebase auth failed. %s", err)
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
		fmt.Sprintf("SELECT user_id FROM vcommerce.user WHERE user_id='%s'", token.UID),
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

	client, err := customContext.App().Auth(context2.Background())
	if err != nil {
		msg := fmt.Sprintf("firebase auth failed. %s", err)
		resp.Status = vcomError.FirebaseAuthFailed
		resp.Detail = msg
		return ctx.JSON(http.StatusInternalServerError, resp)
	}

	token, err := client.VerifyIDToken(context2.Background(), loginRequest.IdToken)
	if err != nil {
		msg := fmt.Sprintf("firebase auth failed. %s", err)
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
		fmt.Sprintf("SELECT user_id FROM vcommerce.user WHERE user_id='%s'", token.UID),
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

	log.Info("Already registered user.")
	serverSessionToken := util.RandString()
	values := []interface{}{
		serverSessionToken,
		token.UID,
	}
	resultCh := make(chan database.CudQueryResult)
	select {
	case customContext.InsertQueryWritePump() <- database.NewCudTransaction(query.InsertSession, values, resultCh):
	case <-timer.C:
		log.Error("failed to exec query")
		resp.Status = vcomError.ApiOperationRequestTimeout
		resp.Detail = vcomError.MessageOperationTimeout
		return ctx.JSON(http.StatusInternalServerError, resp)
	}

	select {
	case res := <-resultCh:
		if res.Err != nil {
			log.Error("database operation failed.")
			resp.Status = vcomError.DatabaseOperationError
			resp.Detail = res.Err.Error()
			return ctx.JSON(http.StatusInternalServerError, resp)
		}

	case <-timer.C:
		// TODO rollback needed
		log.Error("database operation timeout.")
		resp.Status = vcomError.ApiOperationResponseTimeout
		resp.Detail = vcomError.MessageOperationTimeout
		return ctx.JSON(http.StatusInternalServerError, resp)
	}
	resp.SessionToken = serverSessionToken
	resp.Status = vcomError.QueryResultOk
	return ctx.JSON(http.StatusOK, resp)
}

func logout(ctx echo.Context) error {

	return nil
}
