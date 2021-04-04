package session

import (
	"errors"
	"fmt"
	"github.com/4538cgy/backend-second/database"
	"github.com/4538cgy/backend-second/log"
	"github.com/4538cgy/backend-second/query"
	"github.com/4538cgy/backend-second/util"
	"time"
)

type Handler interface {
	InitHandler() error
	InsertSession(uid string, timer *time.Timer) (string, error)            // sessionToken, error
	ValidateSession(sessionToken string, timer *time.Timer) (string, error) // unique_id, error
	UpdateSession(sessionToken string) (string, error)                      // newSessionToken, error
}

type sessionHandler struct {
	dbManager database.Manager
}

func NewSessionHandler(dbManager database.Manager) Handler {
	return &sessionHandler{
		dbManager: dbManager,
	}
}

func (s *sessionHandler) InitHandler() error {
	return nil
}

func (s *sessionHandler) InsertSession(uid string, timer *time.Timer) (string, error) {
	serverSessiontoken := util.RandString()
	values := []interface{}{
		serverSessiontoken,
		uid,
	}
	resultCh := make(chan database.CudQueryResult)
	select {
	case s.dbManager.InsertQueryWritePump() <- database.NewCudTransaction(query.InsertSession, values, resultCh):
	case <-timer.C:
		log.Error("failed to exec query")
		return "", errors.New("session insertion timeout")
	}

	select {
	case res := <-resultCh:
		if res.Err != nil {
			log.Errorf("database operation failed. err: %s", res.Err)
			return "", errors.New("session result failed")
		}

	case <-timer.C:
		log.Error("database operation timeout.")
		return "", errors.New("rollback needed")
	}
	return serverSessiontoken, nil
}

func (s *sessionHandler) ValidateSession(sessionToken string, timer *time.Timer) (string, error) {
	query := fmt.Sprintf("SELECT unique_id FROM vcommerce.session WHERE token='%s' LIMIT 1", sessionToken)
	result := make(chan database.SelectQueryResult)
	select {
	case s.dbManager.SelectQueryWritePump() <- database.NewSelectTransaction(
		query,
		result,
	):
	case <-timer.C:
		return "", errors.New("validate session query timeout")
	}

	var uniqueId string
	// receive result
	select {
	case res := <-result:
		if res.Err != nil {
			return "", errors.New("session result failed")
		}
		if res.Rows.Next() {
			res.Rows.Scan(&sessionToken)
		}
	case <-timer.C:
		return "", errors.New("rollback needed")
	}

	if uniqueId == "" {
		return "", errors.New("nothing found")
	}
	return uniqueId, nil
}

func (s *sessionHandler) UpdateSession(sessionToken string) (string, error) {

	return "", nil
}
