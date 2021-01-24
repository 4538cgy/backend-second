package database

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/4538cgy/backend-second/config"
	"github.com/4538cgy/backend-second/log"
	"github.com/4538cgy/backend-second/util"
	_ "github.com/go-sql-driver/mysql"
)

const driverMySql = "mysql"
const dbQueryPumpChannelBufferSize = 128

type queryTx struct {
	query  string
	result chan<- error
}

func NewQuery(query string, ch chan<- error) *queryTx {
	return &queryTx{
		query:  query,
		result: ch,
	}
}

type manager struct {
	conf           *config.Config
	db             *sql.DB
	shardedWriteCh map[int]chan string
	writeCh        chan string
}

type Manager interface {
	Connect() error
	DSN() string
	SerializedWritePump(shardedKey string) chan<- string
	WritePump() chan<- string
}

func NewDBManager(cfg *config.Config) (Manager, error) {
	switch cfg.Database.Driver {
	case driverMySql:
	default:
		return nil, errors.New(fmt.Sprintf("wrong database driver name: %s", cfg.Database.Driver))
	}
	manager := manager{
		conf:           cfg,
		writeCh:        make(chan string, dbQueryPumpChannelBufferSize),
		shardedWriteCh: map[int]chan string{},
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
		go m.run()
		go m.syncRun(index)
	}

	return nil
}

func (m *manager) WritePump() chan<- string {
	return m.writeCh
}

func (m *manager) SerializedWritePump(shardedKey string) chan<- string {
	value := util.StringToValue(shardedKey) % m.conf.Database.MaxOpenConnection
	return m.shardedWriteCh[value]
}

func (m *manager) run() {
	for {
		select {
		case query, ok := <-m.writeCh:
			if !ok {
				log.Info("channel closed.")
				return
			}

			log.Debug("query: ", query)
		}
	}
}

func (m *manager) syncRun(bucket int) {
	log.Info("sync run exec. name -> ", bucket)
	for {
		select {
		case query, ok := <-m.shardedWriteCh[bucket]:
			if !ok {
				log.Info("channel is closed")
			}

			log.Debug("query: ", query)
		}
	}
}
