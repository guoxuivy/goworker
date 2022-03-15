package db

import (
	"cxe/helper"
	"database/sql"
	"net/url"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"go.mongodb.org/mongo-driver/mongo"
)

type Database struct {
	Mongo *mongo.Client
	// Mysql
	Db *sql.DB
	// mysql 当前事务连接
	Tx *TX
}

func (db *Database) GetCollection(collectName string) *mongo.Collection {
	dbName := "cxe"
	u, err := url.Parse(helper.Config.Database.Mongodb.Dsn)
	if err == nil {
		dbName = strings.Trim(u.Path, "/")
	}
	return db.Mongo.Database(dbName).Collection(collectName)
}

var DB *Database

func init() {
	DB = &Database{
		Mongo: MgoConnect(helper.Config.Database.Mongodb.Dsn),
		Db:    MysqlConnect(helper.Config.Database.Mysql.Dsn),
	}
}

// mysql 事务连接句柄
type TX struct {
	*sql.Tx
	// 当前事务层级
	TransactionLevel int8
	// 发生的错误信息
	Err error
}
