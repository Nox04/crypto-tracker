package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"strings"

	"github.com/leekchan/accounting"
	"github.com/spf13/pflag"
	"github.com/tidwall/gjson"
)

const endpoint = "https://pro-api.coinmarketcap.com/v2/cryptocurrency/quotes/latest"

type output struct {
	Text    string `json:"text"`
	Tooltip string `json:"tooltip"`
}

func requestData(key string, symbols string) (string, error) {
	client := http.Client{}
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return "", errors.New("error creating request")
	}

	q := url.Values{}
	q.Add("symbol", symbols)

	req.Header.Set("Accepts", "application/json")
	req.Header.Add("X-CMC_PRO_API_KEY", key)
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return "", errors.New("error sending request")
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.New("error reading response body")
	}

	return string(respBody), nil
}

func main() {

	var mainSymbol, key, symbolsToTrack string

	// Default values
	symbols := "BTC,DOT,BNB,ETH,FLOW"
	mainSymbolToShow := "BTC"

	pflag.StringVarP(&mainSymbol, "mainsymbol", "m", "", "main symbol to display")
	pflag.StringVarP(&key, "key", "k", "", "coinmarketcap api key")
	pflag.StringVarP(&symbolsToTrack, "symbols", "s", "", "symbols to track, comma separated")
	pflag.Parse()

	if key == "" {
		log.Fatalln("no key provided. Please use -k or --key to provide your coinmarketcap api key")
	}

	if symbolsToTrack != "" {
		symbols = symbolsToTrack
	}

	if mainSymbol != "" {
		mainSymbolToShow = mainSymbol
	}

	symbolsArray := strings.Split(symbols, ",")
	ac := accounting.Accounting{Symbol: "$", Precision: 2}
	rawData, err := requestData(key, symbols)
	if err != nil {
		log.Fatalln(err)
	}

	var mainText, tooltipText string

	for i, symbol := range symbolsArray {
		price := gjson.Get(rawData, "data."+symbol+".0.quote.USD.price").Float()
		price = math.Ceil(price*100) / 100
		if symbol == mainSymbolToShow {
			mainText = fmt.Sprintf("%s: %s", mainSymbolToShow, ac.FormatMoney(price))
		}
		tooltipText += fmt.Sprintf("%s: %s", symbol, ac.FormatMoney(price))
		if i != len(symbolsArray)-1 {
			tooltipText += "\n"
		}
	}

	if mainText == "" {
		price := gjson.Get(rawData, "data."+symbolsArray[0]+".0.quote.USD.price").Float()
		price = math.Ceil(price*100) / 100
		mainText = fmt.Sprintf("%s: %s", symbolsArray[0], ac.FormatMoney(price))
	}

	o := output{mainText, tooltipText}
	result, err := json.Marshal(&o)
	if err != nil {
		log.Fatalln("error marshalling output")
	}

	fmt.Println(string(result))
}
