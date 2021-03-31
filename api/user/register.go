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
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const (
	checkEmailUri   = "/api/user/email"
	checkUserIdUri  = "/api/user/id"
	registerUserUri = "/api/user"

	paramEmail = "email"
	paramUser  = "id"
)

func init() {
	route.AddRoute(route.NewRouteType(registerUserUri, "POST"), registerUser)
	route.AddRoute(route.NewRouteType(checkEmailUri, "GET"), checkEmail)
	route.AddRoute(route.NewRouteType(checkUserIdUri, "GET"), checkUserId)
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

	uniqueId := ctx.FormValue("unique_id")
	userId := ctx.FormValue("user_id")
	emailAddress := ctx.FormValue("email")
	cellPhoneNumber := ctx.FormValue("cell_phone_number")
	dayOfBirth := ctx.FormValue("day_of_birth")
	auth := ctx.FormValue("auth")
	meta := ctx.FormValue("meta")
	timer := time.NewTimer(time.Duration(config.Get().Api.HandleTimeoutMS) * time.Second)
	defer timer.Stop()

	// email insert first
	resultCh := make(chan database.CudQueryResult)
	values := []interface{}{
		emailAddress,
	}
	select {
	case customContext.InsertQueryWritePump() <- database.NewCudTransaction(query.InsertEmail, values, resultCh):
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

	// user id insert
	resultCh = make(chan database.CudQueryResult)
	values = []interface{}{
		userId,
	}
	select {
	case customContext.InsertQueryWritePump() <- database.NewCudTransaction(query.InsertUserID, values, resultCh):
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

	// save file
	file, err := ctx.FormFile("file")
	if err != nil {
		log.Error("FormFile failed. err: ", err)
		return err
	}
	src, err := file.Open()
	if err != nil {
		log.Error("File Open failed. err: ", err)
		return err
	}
	defer src.Close()

	ext := filepath.Ext(file.Filename)
	filePath := fmt.Sprintf(config.Get().Asset.UserProfileImageSavePath + "/" + uniqueId + ext)
	// Destination
	dst, err := os.Create(filePath) // TODO s3 나 특정 위치로 파일을 옮길 수 있어야 함.
	if err != nil {
		log.Error("File Create failed. err: ", err)
		return err
	}
	defer dst.Close()
	// Copy
	if _, err = io.Copy(dst, src); err != nil {
		log.Error("File Copy failed. err: ", err)
		return err
	}

	profileImagePath := "/asset/profile/" + uniqueId + ext
	log.Info("file saved: ", filePath)
	// user insert
	resultCh = make(chan database.CudQueryResult)
	values = []interface{}{
		uniqueId,
		userId,
		dayOfBirth,
		cellPhoneNumber,
		profileImagePath,
		emailAddress,
		auth,
		meta,
	}

	select {
	case customContext.InsertQueryWritePump() <- database.NewCudTransaction(query.InsertUser, values, resultCh):
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

	resp.Token = util.RandString()

	// insert session
	values = []interface{}{
		resp.Token,
		uniqueId,
	}
	resultCh = make(chan database.CudQueryResult)
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
	case customContext.SelectQueryWritePump() <- database.NewSelectTransaction(
		fmt.Sprintf("select email from vcommerce.emails where email='%s' limit 1", emailAddress),
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

func checkUserId(ctx echo.Context) error {
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
	case customContext.SelectQueryWritePump() <- database.NewSelectTransaction(
		fmt.Sprintf("select user_id from vcommerce.userids where user_id='%s' limit 1", userId),
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
