package main

import (
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/thrasher-/gocryptotrader/exchanges/wex"
)

type Storage struct {
	db *sqlx.DB
}

func (s *Storage) SaveTicker(ticker map[string]wex.Ticker) (err error) {
	for pair, tick := range ticker {
		_, err = s.db.Exec(`INSERT INTO ticker (pair, high, low, avg, vol, vol_cur, last, buy, sell, updated)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			ON CONFLICT (pair, updated) DO NOTHING`,
			pair, tick.High, tick.Low, tick.Avg, tick.Vol, tick.VolumeCurrent, tick.Last, tick.Buy, tick.Sell, tick.Updated)
	}
	return
}

func (s *Storage) GetLastAverage(d time.Duration) (prices []*LastPricesAvg, err error) {
	now := time.Now().Unix()
	rows, err := s.db.Queryx("SELECT pair, AVG(last) as last_avg FROM ticker WHERE updated > $1 GROUP BY pair", (now-int64(d))/1000)
	if err != nil {
		return
	}
	for rows.Next() {
		var p LastPricesAvg
		err = rows.StructScan(&p)
		prices = append(prices, &p)
	}
	return
}
