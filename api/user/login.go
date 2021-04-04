package user

import (
	context2 "context"
	"fmt"
	"github.com/4538cgy/backend-second/api/context"
	"github.com/4538cgy/backend-second/api/route"
	vcomError "github.com/4538cgy/backend-second/error"
	"github.com/4538cgy/backend-second/log"
	"github.com/4538cgy/backend-second/protocol"
	"github.com/labstack/echo/v4"
	"net/http"
)

const (
	authUri = "/api/user/auth"
)

type authType int

const (
	authTypeFirebase = authType(1)
	authTypeEmail    = authType(2)
)

var authMap = map[string]authType{
	"google": authTypeFirebase,
	"kakao":  authTypeFirebase,
	"apple":  authTypeFirebase,
	"email":  authTypeEmail,
}

func init() {
	route.AddRoute(route.NewRouteType(authUri, "POST"), login)
	route.AddRoute(route.NewRouteType(authUri, "DELETE"), logout)
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

	client, err := customContext.App().Auth(context2.Background())
	if err != nil {
		msg := fmt.Sprintf("firebase auth failed. %s", err)
		log.Errorf(msg)
		resp.Status = vcomError.InternalError
		resp.Detail = msg
		return ctx.JSON(http.StatusInternalServerError, resp)
	}

	loginRequest := &protocol.LoginRequest{}
	err = ctx.Bind(loginRequest)
	if err != nil {
		log.Error("failed to bind register user request")
		resp.Status = vcomError.InternalError
		resp.Detail = vcomError.MessageBindFailed
	}

	ctx2 := context2.Background()
	token, err := client.CustomToken(ctx2, loginRequest.Token)
	if err != nil {
		msg := fmt.Sprintf("firebase customToken get failed. %s", err)
		log.Errorf(msg)
		resp.Status = vcomError.InternalError
		resp.Detail = msg
		return ctx.JSON(http.StatusInternalServerError, resp)
	}

	log.Infof("custom token generated: %s", token)
	resp.Status = vcomError.QueryResultOk
	resp.Token = token
	return ctx.JSON(http.StatusOK, resp)

	/*
			timer := time.NewTimer(time.Duration(config.Get().Api.HandleTimeoutMS) * time.Second)
			defer timer.Stop()

			var queryString string
			auth := authMap[loginRequest.AuthType]
			switch auth {
			case authTypeFirebase:
				queryString = fmt.Sprintf("select user_id from vcommerce.user where unique_id='%s' limit 1", loginRequest.UniqueId)
			case authTypeEmail:
				queryString = fmt.Sprintf("select user_id from vcommerce.user where unique_id='%s' and password='%s' limit 1", loginRequest.UniqueId, loginRequest.Password)
			default:
				resp.Status = vcomError.InvalidAuthType
				resp.Detail = vcomError.MessageInvalidAuthType
				return ctx.JSON(http.StatusOK, resp)
			}
			result := make(chan database.SelectQueryResult)
			select {
			case customContext.SelectQueryWritePump() <- database.NewSelectTransaction(
				queryString,
				result,
			):
			case <-timer.C:
				resp.Status = vcomError.ApiOperationRequestTimeout
				resp.Detail = vcomError.MessageOperationTimeout
				return ctx.JSON(http.StatusInternalServerError, resp)
			}

			select {
			case res := <-result:
				if res.Err != nil {
					resp.Status = vcomError.DatabaseOperationError
					resp.Detail = res.Err.Error()
					return ctx.JSON(http.StatusInternalServerError, resp)
				}
				resp.Status = vcomError.UserNotFound
				if res.Rows.Next() {
					resp.Status = vcomError.QueryResultOk
					resp.Detail = vcomError.MessageEmailBeingUsed
				}
			case <-timer.C:
				resp.Status = vcomError.ApiOperationResponseTimeout
				resp.Detail = vcomError.MessageOperationTimeout
				return ctx.JSON(http.StatusInternalServerError, resp)
			}
	q
			resp.Token = token

			// insert session
			values := []interface{}{
				resp.Token,
				loginRequest.UniqueId,
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
			resp.Status = vcomError.QueryResultOk

			return ctx.JSON(http.StatusOK, resp)

	*/
}

func logout(ctx echo.Context) error {

	return nil
}
