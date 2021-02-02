package admin

import (
	"github.com/4538cgy/backend-second/log"
	"github.com/labstack/echo/v4"
	"net/http"
)

func init() {
	go func() {
		e := echo.New()
		e.GET("/health", func(context echo.Context) error {
			return context.String(http.StatusOK, "")
		})
		log.Fatal(echo.New().Start(":8888"))
	}()
}
