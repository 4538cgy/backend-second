package seller

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
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	sellerAuthUrl = "/api/sale/auth"
)

type sellerAuthType int

const (
	sellerWaitAuthentication = sellerAuthType(0)
	sellerAuthenticated      = sellerAuthType(1)
)

func init() {
	route.AddRoute(route.NewRouteType(sellerAuthUrl, "POST"), authSeller)
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

	var err error
	authSellerRequest := &protocol.SellerAuthRequest{}
	authSellerRequest.UniqueId = ctx.FormValue("unique_id")
	authSellerRequest.Token = ctx.FormValue("token")
	authSellerRequest.SellerType, err = strconv.Atoi(ctx.FormValue("seller_type"))
	if err != nil {
		log.Error("failed to bind register user request")
		resp.Status = vcomError.InternalError
		resp.Detail = vcomError.MessageBindFailed
		return ctx.JSON(http.StatusInternalServerError, resp)
	}
	authSellerRequest.CompanyRegistrationNumber = ctx.FormValue("company_registration_number")
	authSellerRequest.CompanyOwnerName = ctx.FormValue("owner_name")
	authSellerRequest.CompanyName = ctx.FormValue("company_name")
	authSellerRequest.ChannelName = ctx.FormValue("channel_name")
	authSellerRequest.ChannelUrl = ctx.FormValue("channel_url")
	authSellerRequest.ChannelDescription = ctx.FormValue("channel_description")
	authSellerRequest.BankName = ctx.FormValue("bank_name")
	authSellerRequest.BankAccountNumber = ctx.FormValue("bank_account_number")
	file, err := ctx.FormFile("file")
	if err != nil {
		return err
	}
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	filePath := fmt.Sprintf(config.Get().Api.SellerUploadFilePath + "/" + authSellerRequest.UniqueId + ".pdf")
	// Destination
	dst, err := os.Create(filePath) // TODO s3 나 특정 위치로 파일을 옮길 수 있어야 함.
	if err != nil {
		return err
	}
	defer dst.Close()
	// Copy
	if _, err = io.Copy(dst, src); err != nil {
		return err
	}
	// TODO check user qualification on redis

	timer := time.NewTimer(time.Duration(config.Get().Api.HandleTimeoutMS) * time.Second)
	defer timer.Stop()

	// channel registration first
	resultCh := make(chan database.CudQueryResult)
	values := []interface{}{
		authSellerRequest.ChannelName,
	}
	select {
	case customContext.InsertQueryWritePump() <- database.NewCudTransaction(query.InsertSellerChannel, values, resultCh):
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
	resultCh = make(chan database.CudQueryResult)
	authProgress := sellerWaitAuthentication
	if authSellerRequest.SellerType == 0 {
		authProgress = sellerAuthenticated
	}
	values = []interface{}{
		authSellerRequest.UniqueId,
		authProgress,
	}

	select {
	case customContext.InsertQueryWritePump() <- database.NewCudTransaction(query.InsertSellerRegistration, values, resultCh):
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
	resultCh = make(chan database.CudQueryResult)
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
		filePath,
	}

	select {
	case customContext.InsertQueryWritePump() <- database.NewCudTransaction(query.InsertSellerAuth, values, resultCh):
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
