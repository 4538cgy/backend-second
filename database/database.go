package database

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/4538cgy/backend-second/config"
	"github.com/4538cgy/backend-second/log"
	_ "github.com/go-sql-driver/mysql"
)

const driverMySql = "mysql"
const dbQueryPumpChannelBufferSize = 128

type SelectQueryResult struct {
	Err  error
	Rows *sql.Rows
}

type selectQueryTx struct {
	selectQuery string
	resultCh    chan<- SelectQueryResult
}

func NewSelectQuery(query string, ch chan<- SelectQueryResult) selectQueryTx {
	return selectQueryTx{
		selectQuery: query,
		resultCh:    ch,
	}
}

type InsertQueryResult struct {
	Err    error
	Result sql.Result
}

type insertQueryTx struct {
	insertQuery string
	args        []interface{}
	resultCh    chan<- InsertQueryResult
}

func NewInsertQuery(query string, args []interface{}, ch chan<- InsertQueryResult) insertQueryTx {
	return insertQueryTx{
		insertQuery: query,
		args:        args,
		resultCh:    ch,
	}
}

type manager struct {
	conf               *config.Config
	db                 *sql.DB
	selectQueryWriteCh chan selectQueryTx
	insertQueryWriteCh chan insertQueryTx
}

type Manager interface {
	Connect() error
	DSN() string
	SelectQueryWritePump() chan<- selectQueryTx
	InsertQueryWritePump() chan<- insertQueryTx
}

func NewDBManager(cfg *config.Config) (Manager, error) {
	switch cfg.Database.Driver {
	case driverMySql:
	default:
		return nil, errors.New(fmt.Sprintf("wrong database driver name: %s", cfg.Database.Driver))
	}
	manager := manager{
		conf:               cfg,
		selectQueryWriteCh: make(chan selectQueryTx, dbQueryPumpChannelBufferSize),
		insertQueryWriteCh: make(chan insertQueryTx, dbQueryPumpChannelBufferSize),
	}
	return &manager, nil
}

func (m *manager) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		m.conf.Database.Id,
		m.conf.Database.Password,
		m.conf.Database.IpAddress,
		m.conf.Database.Port,
		m.conf.Database.DbName)
}

func (m *manager) Connect() error {
	db, err := sql.Open(m.conf.Database.Driver, m.DSN())
	if err != nil {
		return err
	}
	m.db = db

	for index := 0; index < m.conf.Database.MaxOpenConnection; index++ {
		go m.readPump()
		go m.writePump()
	}

	return nil
}

func (m *manager) SelectQueryWritePump() chan<- selectQueryTx {
	return m.selectQueryWriteCh
}

func (m *manager) InsertQueryWritePump() chan<- insertQueryTx {
	return m.insertQueryWriteCh
}

func (m *manager) readPump() {
	for {
		select {
		case query, ok := <-m.selectQueryWriteCh:
			if !ok {
				log.Info("unexpected channel close")
				return
			}

			log.Debug("selectQuery: ", query)
			res, err := m.db.Query(query.selectQuery)
			if err != nil {
				query.resultCh <- SelectQueryResult{
					Err: err,
				}
				continue
			}
			log.Debug("Query success. res: ", res)
			query.resultCh <- SelectQueryResult{
				Rows: res,
			}
		}
	}
}

func (m *manager) writePump() {
	for {
		select {
		case query, ok := <-m.insertQueryWriteCh:
			if !ok {
				log.Info("unexpected channel closed.")
			}

			log.Debug("insertQuery: ", query)
			stmt, err := m.db.Prepare(query.insertQuery)
			if err != nil {
				query.resultCh <- InsertQueryResult{
					Err: err,
				}
				continue
			}
			res, err := stmt.Exec(query.args...)
			if err != nil {
				query.resultCh <- InsertQueryResult{
					Err: err,
				}
				continue
			}
			log.Debug("Query success. res: ", res)
			query.resultCh <- InsertQueryResult{
				Result: res,
			}
		}
	}
}
