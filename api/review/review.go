package review

import (
	"encoding/json"
	"github.com/4538cgy/backend-second/api/context"
	"github.com/4538cgy/backend-second/api/route"
	"github.com/4538cgy/backend-second/api/types"
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
	"strconv"
	"time"
)

const (
	reviewUrl = "/api/review"
)

func init() {
	route.AddRoute(route.NewRouteType(reviewUrl, "POST"), postReview)
}

func postReview(ctx echo.Context) error {
	resp := &protocol.ProductPostResponse{}
	customContext, ok := ctx.(*context.CustomContext)
	if !ok {
		log.Error("failed to casting echo.Context to api.CustomContext")
		resp.Status = vcomError.InternalError
		resp.Detail = vcomError.MessageUnknownError
		return ctx.JSON(http.StatusInternalServerError, resp)
	}

	reviewId := util.RandString()
	uniqueId := ctx.FormValue("unique_id")
	token := ctx.FormValue("token")
	log.Info("token: ", token)
	// TODO token validation check
	productId := ctx.FormValue("product_id")
	thumbsUpAndDown := util.RandString()
	log.Info("ThumbsUpAndDown ID: ", thumbsUpAndDown) // TODO Redis 등록
	bodyMessage := ctx.FormValue("body")
	starScore, err := strconv.Atoi(ctx.FormValue("star")) // 별점
	if err != nil {
		log.Error("data invalid. err: ", err.Error(), ",  startScore: ", ctx.FormValue("star"))
		resp.Status = vcomError.InternalError
		resp.Detail = vcomError.MessageUnknownError
		return ctx.JSON(http.StatusInternalServerError, resp)
	}

	form, err := ctx.MultipartForm()
	if err != nil {
		return err
	}
	medias := form.File["files"]
	mediaIndices := types.MediaIndices{MediaIds: make([]string, 0)}
	mediaInfos := types.MediaInfos{Item: make([]types.MediaInfo, 0)}

	// TODO S3로 옮기자. 이미지의 경우 s3로 바로 올리고, 동영상의 경우 hls 변환 후 s3로 올리자.
	for _, file := range medias {
		// Source
		src, err := file.Open()
		if err != nil {
			return err
		}
		defer src.Close()

		// Destination
		dst, err := os.Create(file.Filename)
		if err != nil {
			return err
		}
		defer dst.Close()

		// Copy
		if _, err = io.Copy(dst, src); err != nil {
			resp.Status = vcomError.InternalError
			resp.Detail = vcomError.MessageIOFailed
			return ctx.JSON(http.StatusInternalServerError, resp)
		}
		id := util.RandString()
		mediaInfos.Item = append(mediaInfos.Item, types.MediaInfo{
			Kind:     types.VideoType.String(), // TODO 확장자 보고 type 설정해서 넣읍시다.
			MediaId:  id,
			MediaUrl: "", // TODO url 미리 생성해서 넘기도록 합시다.
		})
		mediaIndices.MediaIds = append(mediaIndices.MediaIds, id)
	}
	// TODO 파일 변환 query

	timer := time.NewTimer(time.Duration(config.Get().Api.HandleTimeoutMS) * time.Second)
	defer timer.Stop()

	// video_info first
	for _, media := range mediaInfos.Item {
		resultCh := make(chan database.CudQueryResult)
		values := []interface{}{
			media.MediaId,
			media.MediaUrl,
		}
		select {
		case customContext.InsertQueryWritePump() <- database.NewCudTransaction(query.InsertVideoList, values, resultCh):
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

	}

	mediaInfoJson, err := json.Marshal(&mediaIndices)
	if err != nil {
		resp.Status = vcomError.InternalError
		resp.Detail = vcomError.MessageUnknownError
		return ctx.JSON(http.StatusInternalServerError, resp)
	}
	// channel registration first
	resultCh := make(chan database.CudQueryResult)
	values := []interface{}{
		reviewId,
		productId,
		uniqueId,
		thumbsUpAndDown,
		bodyMessage,
		string(mediaInfoJson),
		starScore,
	}
	select {
	case customContext.InsertQueryWritePump() <- database.NewCudTransaction(query.InsertReview, values, resultCh):
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
