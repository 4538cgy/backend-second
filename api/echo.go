package api

import (
	"fmt"
	"github.com/4538cgy/backend-second/api/context"
	"github.com/4538cgy/backend-second/api/route"
	_ "github.com/4538cgy/backend-second/api/user"
	"github.com/4538cgy/backend-second/config"
	"github.com/4538cgy/backend-second/database"
	"github.com/4538cgy/backend-second/log"
	"github.com/labstack/echo/v4"
)

type apiManager struct {
	echo      *echo.Echo
	config    *config.Config
	dbManager database.Manager
}

func StartAPI(cfg *config.Config, dbManager database.Manager) {
	api := &apiManager{
		echo:      echo.New(),
		config:    cfg,
		dbManager: dbManager,
	}

	api.echo.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cc := &context.CustomContext{
				Context: c,
				Manager: dbManager,
			}
			return next(cc)
		}
	})

	route.Range(func(routeType, routeUri string, fun func(echo.Context) error) bool {
		switch routeType {
		case "GET":
			log.Infof("GET: %s", routeUri)
			api.echo.GET(routeUri, fun)
		case "POST":
			log.Infof("POST: %s", routeUri)
			api.echo.POST(routeUri, fun)
		case "PUT":
			log.Infof("PUT: %s", routeUri)
			api.echo.PUT(routeUri, fun)
		case "DELETE":
			log.Infof("DELETE: %s", routeUri)
			api.echo.DELETE(routeUri, fun)
		default:
			log.Panic("wrong route type: ", routeType)
		}
		return true
	})

	go func() {
		address := fmt.Sprintf(":%d", api.config.Echo.Port)
		log.Fatal(api.echo.Start(address))
	}()

}
