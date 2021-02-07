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
	"github.com/4538cgy/backend-second/query"
	"github.com/4538cgy/backend-second/util"
	"github.com/labstack/echo/v4"
	"net/http"
	"time"
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

	loginRequest := &protocol.LoginRequest{}
	err := ctx.Bind(loginRequest)
	if err != nil {
		log.Error("failed to bind register user request")
		resp.Status = vcomError.InternalError
		resp.Detail = vcomError.MessageBindFailed
	}

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
	case customContext.SelectQueryWritePump() <- database.NewSelectQuery(
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

	resp.Token = util.RandString()

	// insert session
	values := []interface{}{
		resp.Token,
		loginRequest.UniqueId,
	}
	resultCh := make(chan database.InsertQueryResult)
	select {
	case customContext.InsertQueryWritePump() <- database.NewInsertQuery(query.InsertSession, values, resultCh):
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
}

func logout(ctx echo.Context) error {

	return nil
}
