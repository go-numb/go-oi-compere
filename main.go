package main

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/dustin/go-humanize"

	"github.com/go-numb/go-bitmex"

	"github.com/gocarina/gocsv"

	"github.com/okcoin-okex/open-api-v3-sdk/okex-go-sdk-api"
)

func main() {
	done := make(chan bool)

	go getOIByOkex()
	go getOIByBitmex()

	<-done
}

func getOIByOkex() {
	var config okex.Config
	config.Endpoint = "https://www.okex.com/"
	config.ApiKey = ""
	config.SecretKey = ""
	config.Passphrase = ""
	config.TimeoutSecond = 10
	config.IsPrint = false
	config.I18n = okex.ENGLISH
	c := okex.NewClient(config)

	f, err := os.OpenFile("./okex.oi.csv", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0755)
	if err != nil {
		return
	}
	defer f.Close()

	w := gocsv.DefaultCSVWriter(f)

	symbol := "BTC-USD-191227"
	layout := "2006-01-02T15:04:05.000Z"

	w.Write(set())

	for {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		select {
		case <-ticker.C:
			in, err := c.GetFuturesInstrumentOpenInterest(symbol)
			if err != nil {
				continue
			}

			price, _ := c.GetFuturesInstrumentTrades(symbol, nil)
			if len(price) < 1 {
				continue
			}

			var m map[string]string
			m, ok := price[0].(map[string]string)
			if !ok {
				continue
			}

			out := []string{
				in.InstrumentId,
				m["price"],
				fmt.Sprintf("%d", in.Amount),
				fmt.Sprintf("%s", in.Timestamp),
				fmt.Sprintf("%s", time.Now().Format(layout)),
			}

			w.Write(out)
			w.Flush()
		}
	}

}

func getOIByBitmex() {
	c := bitmex.NewAPIClient(bitmex.NewConfiguration())
	ctx := context.Background()

	h20 := &bitmex.InstrumentGetOpts{}
	h20.Symbol.Set("XBTH20")
	h20.Count.Set(1)
	h20.Reverse.Set(true)
	z19 := &bitmex.InstrumentGetOpts{}
	z19.Symbol.Set("XBTZ19")
	z19.Count.Set(1)
	z19.Reverse.Set(true)

	fh20, err := os.OpenFile("./mexH20.oi.csv", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0755)
	if err != nil {
		return
	}
	defer fh20.Close()

	wh20 := gocsv.DefaultCSVWriter(fh20)

	fz19, err := os.OpenFile("./mexZ19.oi.csv", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0755)
	if err != nil {
		return
	}
	defer fz19.Close()

	wz19 := gocsv.DefaultCSVWriter(fz19)

	layout := "2006-01-02T15:04:05.000Z"

	wh20.Write(set())
	wz19.Write(set())

	var wg sync.WaitGroup

	for {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		select {
		case <-ticker.C:
			wg.Add(1)
			go func() {
				defer wg.Done()
				in, _, err := c.InstrumentApi.InstrumentGet(ctx, h20)
				if err != nil {
					return
				}

				if len(in) < 1 {
					return
				}

				out := []string{
					in[0].Symbol,
					humanize.Commaf(in[0].MidPrice),
					fmt.Sprintf("%d", in[0].OpenInterest),
					fmt.Sprintf("%s", in[0].Timestamp.Format(layout)),
					fmt.Sprintf("%s", time.Now().Format(layout)),
				}

				wh20.Write(out)
				wh20.Flush()
			}()

			wg.Add(1)
			go func() {
				defer wg.Done()
				in, _, err := c.InstrumentApi.InstrumentGet(ctx, z19)
				if err != nil {
					return
				}

				if len(in) < 1 {
					return
				}

				out := []string{
					in[0].Symbol,
					humanize.Commaf(in[0].MidPrice),
					fmt.Sprintf("%d", in[0].OpenInterest),
					fmt.Sprintf("%s", in[0].Timestamp.Format(layout)),
					fmt.Sprintf("%s", time.Now().Format(layout)),
				}

				wz19.Write(out)
				wz19.Flush()
			}()

			wg.Wait()

		}
	}

}

func set() []string {
	return []string{
		"symbol",
		"price",
		"oi_vol",
		"api_time",
		"revice_time",
	}
}
