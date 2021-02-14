package seller

import (
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
	sellerAuth = "/api/sell/auth"
)

type sellerAuthType int

const (
	sellerWaitAuthentication = sellerAuthType(0)
	sellerAuthenticated      = sellerAuthType(1)
)

func init() {
	route.AddRoute(route.NewRouteType(sellerAuth, "POST"), authSeller)
}

func authSeller(ctx echo.Context) error {
	resp := &protocol.SellerAuthResponse{}
	customContext, ok := ctx.(*context.CustomContext)
	if !ok {
		log.Error("failed to casting echo.Context to api.CustomContext")
		resp.Status = vcomError.InternalError
		resp.Detail = vcomError.MessageUnknownError
		return ctx.JSON(http.StatusInternalServerError, resp)
	}

	authSellerRequest := &protocol.SellerAuthRequest{}
	err := ctx.Bind(authSellerRequest)
	if err != nil {
		log.Error("failed to bind register user request")
		resp.Status = vcomError.InternalError
		resp.Detail = vcomError.MessageBindFailed
	}

	// TODO check user qualification on redis

	timer := time.NewTimer(time.Duration(config.Get().Api.HandleTimeoutMS) * time.Second)
	defer timer.Stop()

	// channel registration first
	resultCh := make(chan database.InsertQueryResult)
	values := []interface{}{
		authSellerRequest.ChannelName,
	}
	select {
	case customContext.InsertQueryWritePump() <- database.NewInsertQuery(query.InsertSellerChannel, values, resultCh):
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

	// channel registration
	resultCh = make(chan database.InsertQueryResult)
	authProgress := sellerWaitAuthentication
	if authSellerRequest.SellerType == 0 {
		authProgress = sellerAuthenticated
	}
	values = []interface{}{
		authSellerRequest.UniqueId,
		authProgress,
	}

	select {
	case customContext.InsertQueryWritePump() <- database.NewInsertQuery(query.InsertSellerRegistration, values, resultCh):
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

	// TODO 개인회원의 계좌정보 일치 확인 필요.
	// TODO 법인회원 정보 확인.
	// channel auth
	resultCh = make(chan database.InsertQueryResult)
	values = []interface{}{
		authSellerRequest.UniqueId,
		authSellerRequest.SellerType,
		authSellerRequest.CompanyRegistrationNumber,
		authSellerRequest.CompanyOwnerName,
		authSellerRequest.CompanyName,
		authSellerRequest.ChannelName,
		authSellerRequest.ChannelUrl,
		authSellerRequest.ChannelDescription,
		authSellerRequest.BankName,
		authSellerRequest.BankAccountNumber,
	}

	select {
	case customContext.InsertQueryWritePump() <- database.NewInsertQuery(query.InsertSellerAuth, values, resultCh):
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
