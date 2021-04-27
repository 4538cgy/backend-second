package context

import (
	"github.com/4538cgy/backend-second/api/firebase"
	"github.com/4538cgy/backend-second/api/session"
	"github.com/4538cgy/backend-second/database"
	"github.com/labstack/echo/v4"
)

type CustomContext struct {
	echo.Context
	database.Manager
	firebase.Firebase
	session.Session
}
