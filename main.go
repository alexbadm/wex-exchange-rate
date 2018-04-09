package main

import (
	"log"
	"strings"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/thrasher-/gocryptotrader/exchanges/wex"
)

type Specification struct {
	Debug      bool
	Port       int
	Users      []string
	Rate       float32
	Timeout    string `deafult:"2s"`
	ColorCodes map[string]int
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
		log.Print("Set timeout to '2s'")
		tickerInterval = 2 * time.Second
	}

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

	// log.Print(allPairsString)
	// log.Printf("pairs: %d\n", len(info.Pairs))

	for true {
		tickers, err := api.GetTicker(allPairsString)
		if err != nil {
			log.Printf("error: %v", err)
		}
		log.Printf("tickers: %+v\n", tickers)
		time.Sleep(tickerInterval)
	}
}
