package main

import (
	coinApi "github.com/miguelmota/go-coinmarketcap"
	log "github.com/sirupsen/logrus"
)

var SqlConn = "postgres://postgres:123456@localhost/bilirest?sslmode=disable"
var SqlTable = "apicoinmarket_coin"

func fetchCoinData(sig chan coinChan) {
	coinAllData, err := coinApi.GetAllCoinData(0)
	fatalErr(err, "Get coin data error!")

	log.WithFields(log.Fields{
		"count":       len(coinAllData),
		"last_update": coinAllData["bitcoin"].LastUpdated,
	}).Info("Get coin data success")

	sig <- coinChan{
		Error: nil,
		Data:  coinAllData,
	}
}

func main() {
	dataChan := make(chan coinChan, 1)
	errChan := make(chan error, 1)

	go fetchCoinData(dataChan)
	dataRes := <-dataChan
	close(dataChan)

	go saveCoinData(errChan, dataRes.Data.(map[string]coinApi.Coin))
	saveRes := <-errChan
	if saveRes != nil {
		normalErr(saveRes, "Save coin data error!")
	} else {
		log.Info("Save coin data success!")
	}
}
