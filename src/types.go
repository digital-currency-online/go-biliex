package main

import "fmt"

type DBCoin struct {
	ID               int     `db:"id, primarykey, autoincrement"`
	Name             string  `db:"name"`
	Symbol           string  `db:"symbol"`
	Rank             int     `db:"rank"`
	PriceUsd         float64 `db:"price_usd"`
	PriceBtc         float64 `db:"price_btc"`
	Usd24hVolume     float64 `db:"volume_usd_24h"`
	MarketCapUsd     float64 `db:"market_cap_usd"`
	AvailableSupply  float64 `db:"available_supply"`
	TotalSupply      float64 `db:"total_supply"`
	PercentChange1h  float64 `db:"percent_change_1h"`
	PercentChange24h float64 `db:"percent_change_24h"`
	PercentChange7d  float64 `db:"percent_change_7d"`
	LastUpdated      int     `db:"last_update"`
}

func (dbCoin DBCoin) String() string {
	return fmt.Sprintf("name: %v, rank: %v, price: %v, volume:%v, update: %v",
		dbCoin.Name, dbCoin.Rank, dbCoin.PriceUsd, dbCoin.Usd24hVolume, dbCoin.LastUpdated)
}

type coinChan struct {
	Error error
	Data  interface{}
}
