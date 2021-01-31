package route

import (
	"github.com/4538cgy/backend-second/log"
	"github.com/labstack/echo/v4"
	"sync"
)

type routeType struct {
	routeUri  string
	routeType string
}

var routeLock = sync.Mutex{}
var apiMap = map[routeType]func(echo.Context) error{}

func NewRouteType(uri, registerType string) routeType {
	return routeType{uri, registerType}
}

func AddRoute(route routeType, fun func(echo.Context) error) {
	routeLock.Lock()
	defer routeLock.Unlock()
	apiMap[route] = fun
	log.Debug("Add route... ", route.routeUri)
}

func Range(register func(routeType, routeUri string, handler func(echo.Context) error) bool) {
	routeLock.Lock()
	defer routeLock.Unlock()
	for rt, fun := range apiMap {
		if !register(rt.routeType, rt.routeUri, fun) {
			break
		}
	}
}
