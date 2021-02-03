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
	"github.com/labstack/echo/v4"
	"net/http"
	"time"
)

const (
	checkEmailUri   = "/api/user/email"
	checkUserId     = "/api/user/id"
	registerUserUri = "/api/user"

	paramEmail = "email"
	paramUser  = "id"
)

func init() {
	route.AddRoute(route.NewRouteType(registerUserUri, "POST"), registerUser)
	route.AddRoute(route.NewRouteType(checkEmailUri, "GET"), checkEmail)
	route.AddRoute(route.NewRouteType(checkUserId, "GET"), checkUser)
}

func registerUser(ctx echo.Context) error {
	resp := &protocol.RegisterUserResponse{}
	customContext, ok := ctx.(*context.CustomContext)
	if !ok {
		log.Error("failed to casting echo.Context to api.CustomContext")
		resp.Status = vcomError.InternalError
		resp.Detail = vcomError.MessageUnknownError
		return ctx.JSON(http.StatusInternalServerError, resp)
	}

	registerUserReq := &protocol.RegisterUserRequest{}
	err := ctx.Bind(registerUserReq)
	if err != nil {
		log.Error("failed to bind register user request")
		resp.Status = vcomError.InternalError
		resp.Detail = vcomError.MessageBindFailed
	}

	timer := time.NewTimer(time.Duration(config.Get().Api.HandleTimeoutMS) * time.Second)
	defer timer.Stop()

	values := []interface{}{
		registerUserReq.Uid,
		registerUserReq.DayOfBirth,
		registerUserReq.ProfileImage,
		registerUserReq.EmailAddress,
		registerUserReq.AuthType,
		registerUserReq.Meta,
	}
	resultCh := make(chan database.InsertQueryResult)

	select {
	case customContext.InsertQueryWritePump() <- database.NewInsertQuery(query.InsertUserQuery, values, resultCh):
	case <-timer.C:
		resp.Status = vcomError.ApiOperationRequestTimeout
		resp.Detail = vcomError.MessageOperationTimeout
		return ctx.JSON(http.StatusInternalServerError, resp)
	}

	select {
	case res := <-resultCh:
		if res.Err != nil {
			resp.Status = vcomError.DatabaseOperationError
			resp.Detail = res.Err.Error()
			return ctx.JSON(http.StatusInternalServerError, resp)
		}
		resp.Status = vcomError.QueryResultOk

	case <-timer.C:
		resp.Status = vcomError.ApiOperationResponseTimeout
		resp.Detail = vcomError.MessageOperationTimeout
		return ctx.JSON(http.StatusInternalServerError, resp)
	}
	return nil
}

func checkEmail(ctx echo.Context) error {
	resp := &protocol.EmailCheckResponse{}
	customContext, ok := ctx.(*context.CustomContext)
	if !ok {
		log.Error("failed to casting echo.Context to api.CustomContext")
		resp.Status = vcomError.InternalError
		resp.Detail = vcomError.MessageUnknownError
		return ctx.JSON(http.StatusInternalServerError, resp)
	}
	emailAddress := ctx.QueryParam(paramEmail)
	if emailAddress == "" {
		log.Error("no query param.")
		resp.Status = vcomError.InternalError
		resp.Detail = vcomError.MessageQueryParamNotfound
	}

	timer := time.NewTimer(time.Duration(config.Get().Api.HandleTimeoutMS) * time.Second)
	defer timer.Stop()

	result := make(chan database.SelectQueryResult)
	select {
	case customContext.SelectQueryWritePump() <- database.NewSelectQuery(
		fmt.Sprintf("select address from vcommerce.emails where address='%s'", emailAddress),
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
		resp.Status = vcomError.QueryResultOk
		if res.Rows.Next() {
			resp.Status = vcomError.EmailCheckErrorBeingUsed
			resp.Detail = vcomError.MessageEmailBeingUsed
		}
	case <-timer.C:
		resp.Status = vcomError.ApiOperationResponseTimeout
		resp.Detail = vcomError.MessageOperationTimeout
		return ctx.JSON(http.StatusInternalServerError, resp)
	}
	return ctx.JSON(http.StatusOK, resp)
}

func checkUser(ctx echo.Context) error {
	resp := &protocol.UserIdCheckResponse{}
	customContext, ok := ctx.(*context.CustomContext)
	if !ok {
		log.Error("failed to casting echo.Context to api.CustomContext")
		resp.Status = vcomError.InternalError
		resp.Detail = vcomError.MessageUnknownError
		return ctx.JSON(http.StatusInternalServerError, resp)
	}
	userId := ctx.QueryParam(paramUser)
	if userId == "" {
		log.Error("no query param.")
		resp.Status = vcomError.InternalError
		resp.Detail = vcomError.MessageQueryParamNotfound
	}

	timer := time.NewTimer(time.Duration(config.Get().Api.HandleTimeoutMS) * time.Second)
	defer timer.Stop()

	result := make(chan database.SelectQueryResult)
	select {
	case customContext.SelectQueryWritePump() <- database.NewSelectQuery(
		fmt.Sprintf("select user_id from vcommerce.user where user_id='%s'", userId),
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
		resp.Status = vcomError.QueryResultOk
		if res.Rows.Next() {
			resp.Status = vcomError.UserIdCheckErrorBeingUsed
			resp.Detail = vcomError.MessageUserIdBeingUsed
		}
	case <-timer.C:
		resp.Status = vcomError.ApiOperationResponseTimeout
		resp.Detail = vcomError.MessageOperationTimeout
		return ctx.JSON(http.StatusInternalServerError, resp)
	}
	return ctx.JSON(http.StatusOK, resp)
}
