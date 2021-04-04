package api

import (
	"fmt"
	_ "github.com/4538cgy/backend-second/api/auth"
	"github.com/4538cgy/backend-second/api/context"
	"github.com/4538cgy/backend-second/api/firebase"
	"github.com/4538cgy/backend-second/api/route"
	_ "github.com/4538cgy/backend-second/api/sale"
	_ "github.com/4538cgy/backend-second/api/seller"
	"github.com/4538cgy/backend-second/api/session"
	_ "github.com/4538cgy/backend-second/api/user"
	"github.com/4538cgy/backend-second/config"
	"github.com/4538cgy/backend-second/database"
	"github.com/4538cgy/backend-second/log"
	"github.com/labstack/echo/v4"
	"net/http"
)

type apiManager struct {
	echo           *echo.Echo
	config         *config.Config
	dbManager      database.Manager
	fbManager      firebase.Firebase
	sessionHandler session.Handler
}

func StartAPI(cfg *config.Config, dbManager database.Manager) {
	fbManager, err := firebase.NewManager(cfg)
	if err != nil {
		log.Fatal("firebase manager create failed!!! ", err.Error())
	}

	sessionHandler := session.NewSessionHandler(dbManager)

	api := &apiManager{
		echo:           echo.New(),
		config:         cfg,
		dbManager:      dbManager,
		fbManager:      fbManager,
		sessionHandler: sessionHandler,
	}

	api.echo.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cc := &context.CustomContext{
				Context:  c,
				Manager:  dbManager,
				Firebase: fbManager,
				Handler:  sessionHandler,
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
			log.Panic("wrong route types: ", routeType)
		}
		return true
	})

	api.echo.GET("/health", func(context echo.Context) error {
		return context.String(http.StatusOK, "")
	})

	fs := http.FileServer(http.Dir(cfg.Asset.UserProfileImageSavePath))
	api.echo.GET("/assets/profile/*", echo.WrapHandler(http.StripPrefix("/assets/profile", fs)))

	go func() {
		address := fmt.Sprintf(":%d", api.config.Echo.Port)
		log.Fatal(api.echo.Start(address))
	}()

}
