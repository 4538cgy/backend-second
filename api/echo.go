package api

import (
	"github.com/labstack/echo/v4"
	"sync"
)

func newEcho() *echo.Echo {
	return echo.New()
}

var e *echo.Echo
var eLock sync.Mutex

func Echo() *echo.Echo {
	eLock.Lock()
	defer eLock.Unlock()
	if e == nil {
		e = newEcho()
	}
	return e
}