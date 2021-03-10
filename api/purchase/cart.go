package purchase

import (
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
	cartUrl = "/api/purchase/cart"
)

func init() {
	route.AddRoute(route.NewRouteType(cartUrl, "POST"), addCartItem)
	route.AddRoute(route.NewRouteType(cartUrl, "DELETE"), deleteCartItem)
}

func addCartItem(ctx echo.Context) error {
	resp := &protocol.CartItemAddResponse{}
	customContext, ok := ctx.(*context.CustomContext)
	if !ok {
		log.Error("failed to casting echo.Context to api.CustomContext")
		resp.Status = vcomError.InternalError
		resp.Detail = vcomError.MessageUnknownError
		return ctx.JSON(http.StatusInternalServerError, resp)
	}

	cartItemAddRequest := &protocol.CartItemAddRequest{}
	err := ctx.Bind(cartItemAddRequest)
	if err != nil {
		log.Error("failed to bind register user request")
		resp.Status = vcomError.InternalError
		resp.Detail = vcomError.MessageBindFailed
		return ctx.JSON(http.StatusInternalServerError, resp)
	}

	timer := time.NewTimer(time.Duration(config.Get().Api.HandleTimeoutMS) * time.Second)
	defer timer.Stop()

	cart_id := util.RandString()
	resultCh := make(chan database.CudQueryResult)
	values := []interface{}{
		cart_id,
		cartItemAddRequest.UniqueId,
		cartItemAddRequest.ProductId,
		cartItemAddRequest.SelectedJson,
	}
	select {
	case customContext.InsertQueryWritePump() <- database.NewCudTransaction(query.InsertCart, values, resultCh):
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
		log.Error("database operation timeout.")
		resp.Status = vcomError.ApiOperationResponseTimeout
		resp.Detail = vcomError.MessageOperationTimeout
		return ctx.JSON(http.StatusInternalServerError, resp)
	}

	resp.Status = vcomError.QueryResultOk
	return ctx.JSON(http.StatusOK, resp)
}

func deleteCartItem(ctx echo.Context) error {
	resp := &protocol.CartItemRemoveResponse{}
	customContext, ok := ctx.(*context.CustomContext)
	if !ok {
		log.Error("failed to casting echo.Context to api.CustomContext")
		resp.Status = vcomError.InternalError
		resp.Detail = vcomError.MessageUnknownError
		return ctx.JSON(http.StatusInternalServerError, resp)
	}

	cartItemRemoveRequest := &protocol.CartItemRemoveRequest{}
	err := ctx.Bind(cartItemRemoveRequest)
	if err != nil {
		log.Error("failed to bind register user request")
		resp.Status = vcomError.InternalError
		resp.Detail = vcomError.MessageBindFailed
		return ctx.JSON(http.StatusInternalServerError, resp)
	}

	timer := time.NewTimer(time.Duration(config.Get().Api.HandleTimeoutMS) * time.Second)
	defer timer.Stop()

	resultCh := make(chan database.CudQueryResult)
	values := []interface{}{
		cartItemRemoveRequest.CartId,
		cartItemRemoveRequest.UniqueId,
	}
	select {
	case customContext.InsertQueryWritePump() <- database.NewCudTransaction(query.DeleteCart, values, resultCh):
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
		log.Error("database operation timeout.")
		resp.Status = vcomError.ApiOperationResponseTimeout
		resp.Detail = vcomError.MessageOperationTimeout
		return ctx.JSON(http.StatusInternalServerError, resp)
	}

	resp.Status = vcomError.QueryResultOk
	return ctx.JSON(http.StatusOK, resp)
}
