package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/canhlinh/cointracker/config"

	socket "github.com/canhlinh/tradingview-scraper/v2"
)

var TickerDuration = time.Minute * 30

type Price struct {
	LastTimePrice float64
	CurrentPrice  float64
}

func main() {
	ticker := time.NewTicker(time.Minute)
	symbols := map[string]Price{}
	look := &sync.Mutex{}

	tradingviewsocket, err := socket.Connect(
		func(symbol string, data *socket.QuoteData) {
			if data.Price != nil {
				currentPrice := *data.Price
				look.Lock()
				price := symbols[symbol]
				lastTimePrice := price.LastTimePrice
				if lastTimePrice == 0 {
					lastTimePrice = currentPrice
				}

				symbols[symbol] = Price{
					CurrentPrice:  currentPrice,
					LastTimePrice: lastTimePrice,
				}
				look.Unlock()
				log.Printf("%v %v\n", symbol, currentPrice)
			}

			select {
			case <-ticker.C:
				m := time.Now().Minute()
				if m == 0 || m == 30 {
					ticker.Reset(TickerDuration)

					look.Lock()
					price := symbols[symbol]
					if price.CurrentPrice > 0 {
						if price.CurrentPrice > price.LastTimePrice {
							priceChanged := (price.CurrentPrice - price.LastTimePrice) / price.CurrentPrice * 100
							if math.Abs(priceChanged) >= config.Config().PercentPriceChangedAlert {
								sendTelegramMessage(symbol, priceChanged, price.CurrentPrice)
							}
						}
					}
					symbols[symbol] = Price{
						CurrentPrice:  price.CurrentPrice,
						LastTimePrice: price.CurrentPrice,
					}
					look.Unlock()
				}
			default:
				// do nothing
			}
		},
		func(err error, context string) {
			log.Panicf("error -> %#v context-> %v \n", err.Error(), context)
		},
	)
	if err != nil {
		log.Panicf("Error while initializing the trading view socket -> \n" + err.Error())
	}
	for _, symbol := range config.Config().Symbols {
		tradingviewsocket.AddSymbol(symbol)
		symbols[symbol] = Price{}
	}

	for i := 0; i < 1; {
		time.Sleep(time.Second * 1)
	}
}

type TelegramMessage struct {
	ChatId int    `json:"chat_id"`
	Text   string `json:"text"`
}

func sendTelegramMessage(symbol string, priceChanged float64, currentPrice float64) {
	log.Println("sending telegram message", symbol, priceChanged, currentPrice)

	priceChangedFormat := fmt.Sprintf("%.2f%%", priceChanged)
	if priceChanged > 0 {
		priceChangedFormat = "+" + priceChangedFormat
	}
	msg := fmt.Sprintf("%s %s (%v$) in the last 1 hour", symbol, priceChangedFormat, currentPrice)

	buf := &bytes.Buffer{}
	json.NewEncoder(buf).Encode(&TelegramMessage{
		ChatId: config.Config().BotChatID,
		Text:   msg,
	})

	res, err := http.Post(config.BotAPI(), "application/json", buf)
	if err != nil {
		fmt.Println(err)
	}
	if res.StatusCode != 200 {
		fmt.Println(res.Status)
	}
}
