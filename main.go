package main

import (
	"log"
	"strings"
	"time"

	"fmt"
	"net/http"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/kelseyhightower/envconfig"
	"github.com/thrasher-/gocryptotrader/exchanges/wex"
)

type Specification struct {
	Port       string `default:"4000"`
	Timeout    string `deafult:"2s"`
	Dbuser     string `default:"postgres"`
	Dbpassword string `default:"postgres"`
	Dbhost     string `default:"127.0.0.1"`
	Dbname     string `default:"wex"`
	Dbport     string `default:"32768"`
}

var storage *Storage

func Index(w http.ResponseWriter, r *http.Request) {
	if storage == nil || storage.db == nil {
		w.Write([]byte("storage not started"))
		return
	}
	data, err := storage.GetLastAverage(10 * time.Minute)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	dataStrings := make([]string, 0, len(data))
	for _, d := range data {
		dataStrings = append(dataStrings, d.Marshal())
	}
	json := "[" + strings.Join(dataStrings, ",") + "]"
	w.Write([]byte(json))

	/* Alternative approach below */
	// d, err := json.Marshal(data)
	// if err != nil {
	// 	w.WriteHeader(http.StatusInternalServerError)
	// 	return
	// }
	// w.Write(d)
}

type LastPricesAvg struct {
	Pair    string  `db:"pair" json:"pair"`
	LastAvg float64 `db:"last_avg" json:"last_avg"`
}

func (lp *LastPricesAvg) Marshal() string {
	return fmt.Sprintf(`["%s",%f]`, lp.Pair, lp.LastAvg)
}

func main() {
	var s Specification
	err := envconfig.Process("wex", &s)
	if err != nil {
		log.Fatal(err.Error())
	}

	tickerInterval, err := time.ParseDuration(s.Timeout)
	if err != nil {
		log.Printf("Warning: failed to parse Timeout environment variable: %v", err.Error())
	}
	var defaultTimeout = 2 * time.Second
	if tickerInterval < defaultTimeout {
		log.Print("Set timeout to '2s'")
		tickerInterval = defaultTimeout
	}

	storage = new(Storage)
	storage.db, err = sqlx.Connect("postgres", fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=disable",
		s.Dbuser, s.Dbpassword, s.Dbhost, s.Dbport, s.Dbname))
	if err != nil {
		log.Fatal(err.Error())
	}

	http.HandleFunc("/", Index)
	go http.ListenAndServe(":"+s.Port, nil)

	api := wex.WEX{}
	api.SetDefaults()
	info, err := api.GetInfo()
	if err != nil {
		log.Panicf("Failed to get all currency pairs. Error: %v", err)
	}

	pairs := make([]string, 0, len(info.Pairs))
	for pair := range info.Pairs {
		pairs = append(pairs, pair)
	}
	allPairsString := strings.Join(pairs, "-")

	var ticker map[string]wex.Ticker
	for true {
		ticker, err = api.GetTicker(allPairsString)
		if err != nil {
			log.Printf("Failed to get ticker: %v", err)
		}
		log.Print("tick")
		// log.Printf("ticker: %+v\n", ticker)
		err = storage.SaveTicker(ticker)
		if err != nil {
			log.Printf("Storage save ticker error: %v", err)
		}
		time.Sleep(tickerInterval)
	}
}
