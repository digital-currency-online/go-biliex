package main

import (
	"errors"
	"strings"
	"strconv"
	"database/sql"
	"gopkg.in/gorp.v2"
	_ "github.com/lib/pq"
	"github.com/jinzhu/copier"
	coinApi "github.com/miguelmota/go-coinmarketcap"
	log "github.com/sirupsen/logrus"
)

type DBWriter struct {
	DBType     string
	DBPath     string
	TableName  string
	TableModel interface{}
	PrimaryKey string
}

func (db DBWriter) Init(initTableFlag bool) (*gorp.DbMap, error) {
	var dbSql *sql.DB
	var dbMap *gorp.DbMap
	var err error

	log.WithFields(log.Fields{
		"path": db.DBPath,
	}).Info("connect to database: ")

	switch db.DBType {
	case "sqlite3":
		dbSql, err = sql.Open("sqlite3", db.DBPath)
		dbMap = &gorp.DbMap{Db: dbSql, Dialect: gorp.SqliteDialect{}}
	case "mysql":
		dbSql, err = sql.Open("mysql", db.DBPath)
		dbMap = &gorp.DbMap{Db: dbSql, Dialect: gorp.MySQLDialect{
			Engine:   "InnoDB",
			Encoding: "UTF8",
		}}
	case "postgres":
		dbSql, err = sql.Open("postgres", db.DBPath)
		dbMap = &gorp.DbMap{Db: dbSql, Dialect: gorp.PostgresDialect{}}
	default:
		err = errors.New("unknown database driver")
	}

	if err != nil {
		return nil, err
	}

	dbMap.AddTableWithName(db.TableModel, db.TableName).SetKeys(true, db.PrimaryKey)

	if initTableFlag {
		log.WithFields(log.Fields{
			"table": db.TableName,
		}).Info("Init db table")

		// drop table
		err = dbMap.DropTablesIfExists()
		if err != nil {
			return nil, err
		}

		// create table
		err = dbMap.CreateTablesIfNotExists()
		if err != nil {
			return nil, err
		}

		// clear table
		err = dbMap.TruncateTables()
		if err != nil {
			return nil, err
		}
	}

	return dbMap, nil
}

func initDB() (*gorp.DbMap, error) {
	dbWriter := DBWriter{
		DBType:     "postgres",
		DBPath:     SqlConn,
		TableName:  SqlTable,
		TableModel: DBCoin{},
		PrimaryKey: "id",
	}
	return dbWriter.Init(false)
}

// batchInsert used for batch insert data
func batchInsert(dbMap *gorp.DbMap, dataList []DBCoin) error {
	log.WithFields(log.Fields{
		"count": len(dataList),
	}).Info("insert data to " + SqlTable)

	// Start a new transaction
	trans, err := dbMap.Begin()
	if err != nil {
		return err
	}

	for _, value := range dataList {
		trans.Insert(&value)
	}

	return trans.Commit()
}

func checkDBDataExists(dbMap *gorp.DbMap, coinData coinApi.Coin) bool {
	dbCoin, err := convertCoin2DB(coinData)
	fatalErr(err, "Convert dbCoin error!")

	var sqlString string
	if strings.HasPrefix(SqlConn, "postgres") {
		sqlString = "select count(*) from " + SqlTable + " where name=$1 and last_update=$2"
	} else {
		sqlString = "select count(*) from " + SqlTable + " where name=? and last_update=?"
	}

	dataCount, err := dbMap.SelectInt(sqlString, coinData.Name, dbCoin.LastUpdated)
	normalErr(err, "check data exist error!")

	return dataCount > 0
}

func saveCoinData(sigChan chan error, coinAllData map[string]coinApi.Coin) {
	var dbCoinList []DBCoin

	dbMap, err := initDB()
	defer dbMap.Db.Close()

	fatalErr(err, "init database error")

	isExist := checkDBDataExists(dbMap, coinAllData["bitcoin"])
	if isExist {
		sigChan <- errors.New("database already exist data")
		return
	}

	for _, coinData := range coinAllData {
		dbCoin, err := convertCoin2DB(coinData)
		if err == nil {
			dbCoinList = append(dbCoinList, dbCoin)
		}
	}
	batchInsert(dbMap, dbCoinList)
	sigChan <- nil
}

func convertCoin2DB(coinData coinApi.Coin) (DBCoin, error) {
	var dbData DBCoin

	if coinData.LastUpdated == "" {
		log.WithFields(log.Fields{
			"name": coinData.Name,
		}).Info("skip empty coin")
		return dbData, errors.New("skip empty coin")
	}

	lastUpdate, err := strconv.Atoi(coinData.LastUpdated)
	if err != nil {
		return dbData, err
	}

	copier.Copy(&dbData, &coinData)
	dbData.LastUpdated = lastUpdate
	return dbData, nil
}
