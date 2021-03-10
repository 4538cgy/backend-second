package sale

import (
	"encoding/json"
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
	"strconv"
	"time"
)

const (
	sellProductUrl = "/api/sale/product"
)

func init() {
	route.AddRoute(route.NewRouteType(sellProductUrl, "POST"), postProduct)
}

func postProduct(ctx echo.Context) error {
	resp := &protocol.PostProductResponse{}
	customContext, ok := ctx.(*context.CustomContext)
	if !ok {
		log.Error("failed to casting echo.Context to api.CustomContext")
		resp.Status = vcomError.InternalError
		resp.Detail = vcomError.MessageUnknownError
		return ctx.JSON(http.StatusInternalServerError, resp)
	}

	uniqueId := ctx.FormValue("unique_id")
	// token := ctx.FormValue("token")
	// TODO token validation check
	title := ctx.FormValue("title")
	categoryJson := ctx.FormValue("category_info_json")
	optionJson := ctx.FormValue("option_json")
	basePrice, err := strconv.Atoi(ctx.FormValue("base_price"))
	if err != nil {
		log.Error("data invalid. err: ", err.Error(), ", base_price: ", ctx.FormValue("base_price"))
		resp.Status = vcomError.InternalError
		resp.Detail = vcomError.MessageUnknownError
		return ctx.JSON(http.StatusInternalServerError, resp)
	}
	baseAmount, err := strconv.Atoi(ctx.FormValue("base_amount"))
	if err != nil {
		log.Error("data invalid. err: ", err.Error(), ", base_amount: ", ctx.FormValue("base_amount"))
		resp.Status = vcomError.InternalError
		resp.Detail = vcomError.MessageUnknownError
		return ctx.JSON(http.StatusInternalServerError, resp)
	}

	form, err := ctx.MultipartForm()
	if err != nil {
		return err
	}
	videos := form.File["files"]
	type (
		videoInfo struct {
			VideoId    string `json:"video_id"`
			VideoUrl   string `json:"video_url"`
			ServeReady int    `json:"serve_ready"`
		}
		videoInfos struct {
			Item []videoInfo `json:"videoInfos"`
		}
	)

	type videoIndices struct {
		VideoIds []string `json:"video_indices"`
	}
	vIndices := videoIndices{VideoIds: make([]string, 0)}
	vinfos := videoInfos{Item: make([]videoInfo, 0)}

	// TODO S3로 옮기자.
	for _, file := range videos {
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
			resp.Detail = vcomError.MessageFileIoFailed
			return ctx.JSON(http.StatusInternalServerError, resp)
		}
		id := util.RandString()
		vinfos.Item = append(vinfos.Item, videoInfo{
			VideoId:  id,
			VideoUrl: "", // TODO url 미리 생성해서 넘기도록 합시다.
		})
		vIndices.VideoIds = append(vIndices.VideoIds, id)
	}
	// TODO 파일 변환 query

	timer := time.NewTimer(time.Duration(config.Get().Api.HandleTimeoutMS) * time.Second)
	defer timer.Stop()

	// video_info first
	for _, vinfo := range vinfos.Item {
		resultCh := make(chan database.CudQueryResult)
		values := []interface{}{
			vinfo.VideoId,
			vinfo.VideoUrl,
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
	// and then product list

	pid := util.RandString()
	resultCh := make(chan database.CudQueryResult)
	values := []interface{}{
		pid,
		categoryJson,
	}
	select {
	case customContext.InsertQueryWritePump() <- database.NewCudTransaction(query.InsertProductCategoryInfo, values, resultCh):
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

	// and finally, product
	videoIds, err := json.Marshal(&vIndices)
	if err != nil {
		resp.Status = vcomError.InternalError
		resp.Detail = vcomError.MessageUnknownError
		return ctx.JSON(http.StatusInternalServerError, resp)
	}

	// channel registration first
	resultCh = make(chan database.CudQueryResult)
	values = []interface{}{
		pid,
		uniqueId,
		string(videoIds),
		title,
		basePrice,
		baseAmount,
		optionJson,
	}
	select {
	case customContext.InsertQueryWritePump() <- database.NewCudTransaction(query.InsertProductSale, values, resultCh):
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
