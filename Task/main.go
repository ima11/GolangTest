package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var (
	socketUrl               = "wss://api.hitbtc.com/api/2/ws"
	symbolUrl               = "https://api.hitbtc.com/api/2/public/symbol"
	InMemorySavedCurrencies []Result
	Results                 []Result
	requiredSymbol          = []string{"BTCUSD", "ETHBTC"}
	port                    = ":9999"
)

type Result struct {
	ID          string `json:"id"`
	FullName    string `json:"fullName"`
	Ask         string `json:"ask"`
	Bid         string `json:"bid"`
	Last        string `json:"last"`
	Open        string `json:"open"`
	Low         string `json:"low"`
	High        string `json:"high"`
	FeeCurrency string `json:"feeCurrency"`
}

type Symbols struct {
	ID           string `json:"id"`
	BaseCurrency string `json:"baseCurrency"`
	FeeCurrency  string `json:"feeCurrency"`
}
type CurrencyResponse struct {
	Result ResponseResult `json:"result"`
}
type TickerResponse struct {
	Params TickerParams `json:"params"`
}

type TickerParams struct {
	Ask  string `json:"ask"`
	Bid  string `json:"bid"`
	Last string `json:"last"`
	Open string `json:"open"`
	Low  string `json:"low"`
	High string `json:"high"`
}

type ResponseResult struct {
	ID       string `json:"id"`
	FullName string `json:"fullName"`
}

type AllResult struct {
	Currencies []Result `json:"currencies"`
}

type Request struct {
	Method string `json:"method"`
	Params Param  `json:"params"`
	ID     int    `json:"id"`
}
type Param struct {
	Currency string `json:"currency,omitempty"`
	Symbol   string `json:"symbol,omitempty"`
}

func main() {
	mux := mux.NewRouter()
	mux.HandleFunc("/currency/all", GetAllCurrency)
	mux.HandleFunc("/currency/{symbol}", GetSymbolCurrency)
	fmt.Println("Server started at port 9999")
	log.Fatal(http.ListenAndServe(port, mux))
}

func init() {
	symbolResp, err := http.Get(symbolUrl)
	if err != nil {
		fmt.Println("Error in fetching symbol", err)
		return
	}
	symbolBytes, err := ioutil.ReadAll(symbolResp.Body)
	if err != nil {
		fmt.Println("Error in reading response body", err)
		return
	}

	var allsymbols, filteredSymbols []Symbols

	json.Unmarshal(symbolBytes, &allsymbols)

	for _, v := range requiredSymbol {
		for _, value := range allsymbols {
			if value.ID == v {
				filteredSymbols = append(filteredSymbols, value)
			}
		}
	}

	go LoadCurrency(filteredSymbols)
}

func LoadCurrency(allsymbols []Symbols) {

	fmt.Println("Supported symbols", requiredSymbol)

	for {
		for _, v := range allsymbols {
			currresp := readBasicCurrency(v.BaseCurrency)
			tickerresp := readTickerCurrency(v.ID)

			result := Result{
				ID:          currresp.Result.ID,
				FullName:    currresp.Result.FullName,
				Ask:         tickerresp.Params.Ask,
				Bid:         tickerresp.Params.Bid,
				Last:        tickerresp.Params.Last,
				Open:        tickerresp.Params.Open,
				Low:         tickerresp.Params.Low,
				High:        tickerresp.Params.High,
				FeeCurrency: v.FeeCurrency,
			}
			Results = append(Results, result)
		}
		InMemorySavedCurrencies = Results
		Results = []Result{}
	}
}

func GetAllCurrency(res http.ResponseWriter, req *http.Request) {
	allresult := AllResult{Currencies: InMemorySavedCurrencies}
	by, err := json.Marshal(allresult)
	if err != nil {
		fmt.Fprintln(res, "Error in marshalling ", err)
	}
	res.Header().Add("Content-type", "application/json")
	fmt.Fprint(res, string(by))
}

func GetSymbolCurrency(res http.ResponseWriter, req *http.Request) {
	symbol := mux.Vars(req)
	symbolResp, err := http.Get(symbolUrl)
	if err != nil {
		fmt.Println("Error in fetching symbol", err)
		return
	}
	symbolBytes, err := ioutil.ReadAll(symbolResp.Body)
	if err != nil {
		fmt.Println("Error in reading response body", err)
		return
	}

	var symbols []Symbols
	json.Unmarshal(symbolBytes, &symbols)

	exist := contains(symbols, symbol["symbol"])
	if !exist {
		fmt.Fprint(res, "Symbol doesnot exist")
		return
	}
	var result Result
	for i, v := range InMemorySavedCurrencies {
		if v.ID+v.FeeCurrency == symbol["symbol"] {
			result = InMemorySavedCurrencies[i]
		}
	}
	r, err := json.Marshal(result)
	if err != nil {
		fmt.Println("Error in marshalling results")
		fmt.Fprint(res, "Error in marshalling results")
	}

	res.Header().Add("Content-type", "application/json")

	fmt.Fprint(res, string(r))
}

func readBasicCurrency(baseCurrency string) CurrencyResponse {
	conn, _, err := websocket.DefaultDialer.Dial(socketUrl, nil)
	if err != nil {
		log.Fatal("Error connecting to Websocket Server:", err)
	}
	defer conn.Close()

	m := Request{Method: "getCurrency", Params: Param{Currency: baseCurrency}, ID: 123}
	reqBytes, err := json.Marshal(m)
	err = conn.WriteMessage(websocket.TextMessage, reqBytes)
	if err != nil {
		log.Println("Error during writing to websocket:", err)
		return CurrencyResponse{}
	}

	_, msg, err := conn.ReadMessage()
	if err != nil {
		log.Println("Error in receive:", err)
		return CurrencyResponse{}
	}
	fmt.Println(string(msg))
	var currresp CurrencyResponse
	err = json.Unmarshal(msg, &currresp)
	if err != nil {
		log.Println("Error in unmarshalling basic details", err)
		return CurrencyResponse{}
	}
	return currresp
}

func readTickerCurrency(id string) TickerResponse {
	conn, _, err := websocket.DefaultDialer.Dial(socketUrl, nil)
	if err != nil {
		log.Fatal("Error connecting to Websocket Server:", err)
	}
	defer conn.Close()

	m := Request{Method: "subscribeTicker", Params: Param{Symbol: id}, ID: 123}
	reqBytes, err := json.Marshal(m)
	err = conn.WriteMessage(websocket.TextMessage, reqBytes)
	if err != nil {
		log.Println("Error during writing to websocket:", err)
		return TickerResponse{}
	}

	_, _, err = conn.NextReader()
	if err != nil {
		log.Println("Error in receive:", err)
		return TickerResponse{}
	}
	_, msg, err := conn.ReadMessage()
	if err != nil {
		log.Println("Error in receive:", err)
		return TickerResponse{}
	}
	fmt.Println(string(msg))
	var tickerresp TickerResponse
	err = json.Unmarshal(msg, &tickerresp)
	if err != nil {
		log.Println("Error in unmarshalling basic details", err)
		return TickerResponse{}
	}
	return tickerresp
}

func contains(s []Symbols, str string) bool {
	for _, v := range s {
		if v.ID == str {
			return true
		}
	}
	return false
}
