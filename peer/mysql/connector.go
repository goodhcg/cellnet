package mysql

import (
	"database/sql"
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/peer"
	"github.com/go-sql-driver/mysql"
	"sync"
)

type mysqlConnector struct {
	peer.CorePeerProperty
	peer.CoreContextSet
	peer.CoreSQLParameter

	db      *sql.DB
	dbGuard sync.RWMutex
}

func (self *mysqlConnector) IsReady() bool {
	return self.dbConn() != nil
}

func (self *mysqlConnector) Raw() interface{} {

	return self.dbConn()
}

func (self *mysqlConnector) dbConn() *sql.DB {
	self.dbGuard.RLock()
	defer self.dbGuard.RUnlock()
	return self.db
}

func (self *mysqlConnector) TypeName() string {
	return "mysql.Connector"
}

func (self *mysqlConnector) Start() cellnet.Peer {

	self.tryConnect()

	return self
}

func (self *mysqlConnector) tryConnect() {

	config, err := mysql.ParseDSN(self.Address())

	if err != nil {
		log.Errorf("Invalid mysql DSN: %s, %s\n", self.Address(), err.Error())
		return
	}

	log.Infof("Connect to mysql database: %s/%s...", config.Addr, config.DBName)

	db, err := sql.Open("mysql", self.Address())
	if err != nil {
		log.Errorf("Open mysql database error: %s\n", err)
		return
	}

	db.SetMaxOpenConns(int(self.PoolConnCount))
	db.SetMaxIdleConns(int(self.PoolConnCount / 2))

	err = db.Ping()
	if err != nil {
		log.Errorln(err)
		return
	}

	self.dbGuard.Lock()
	self.db = db
	self.dbGuard.Unlock()

	if config != nil {
		log.SetColor("green").Infof("Connected to mysql %s/%s", config.Addr, config.DBName)
	}
}

func (self *mysqlConnector) Stop() {

	db := self.dbConn()
	if db != nil {
		db.Close()
	}

}

func init() {

	peer.RegisterPeerCreator(func() cellnet.Peer {
		self := &mysqlConnector{}
		self.CoreSQLParameter.Init()

		return self
	})
}
