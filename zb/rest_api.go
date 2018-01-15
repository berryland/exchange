package zb

import (
	"github.com/buger/jsonparser"
	"github.com/pkg/errors"
	"github.com/valyala/fasthttp"
	"net/url"
	"strconv"
)

const (
	DataApiUrl  = "http://api.zb.com/data/v1/"
	TradeApiUrl = "https://trade.zb.com/api/"
)

type SymbolConfig struct {
	AmountScale byte
	PriceScale  byte
}

func GetSymbols() (*map[string]SymbolConfig, error) {
	resp, err := doGet(DataApiUrl + "markets")
	if err != nil {
		return nil, errors.WithStack(err)
	}

	configs := map[string]SymbolConfig{}
	jsonparser.ObjectEach(resp.Body(), func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
		symbol, _ := jsonparser.ParseString(key)
		amountScale, _ := jsonparser.GetInt(value, "amountScale")
		priceScale, _ := jsonparser.GetInt(value, "priceScale")
		configs[symbol] = SymbolConfig{byte(amountScale), byte(priceScale)}
		return nil
	})
	return &configs, nil
}

type Quote struct {
	Volume float64
	Last   float64
	Sell   float64
	Buy    float64
	High   float64
	Low    float64
	Time   uint64
}

func GetLatestQuote(symbol string) (*Quote, error) {
	u, _ := url.Parse(DataApiUrl + "ticker")
	q := u.Query()
	q.Set("market", symbol)
	u.RawQuery = q.Encode()

	resp, err := doGet(u.String())
	if err != nil {
		return nil, errors.WithStack(err)
	}

	bytes := resp.Body()
	ticker, _, _, _ := jsonparser.Get(bytes, "ticker")
	volString, _ := jsonparser.GetString(ticker, "vol")
	lastString, _ := jsonparser.GetString(ticker, "last")
	sellString, _ := jsonparser.GetString(ticker, "sell")
	buyString, _ := jsonparser.GetString(ticker, "buy")
	highString, _ := jsonparser.GetString(ticker, "high")
	lowString, _ := jsonparser.GetString(ticker, "low")
	dateString, _ := jsonparser.GetString(bytes, "date")

	vol, _ := strconv.ParseFloat(volString, 64)
	last, _ := strconv.ParseFloat(lastString, 64)
	sell, _ := strconv.ParseFloat(sellString, 64)
	buy, _ := strconv.ParseFloat(buyString, 64)
	high, _ := strconv.ParseFloat(highString, 64)
	low, _ := strconv.ParseFloat(lowString, 64)
	date, _ := strconv.ParseUint(dateString, 10, 64)

	return &Quote{Volume: vol, Last: last, Sell: sell, Buy: buy, High: high, Low: low, Time: date}, nil
}

type Kline struct {
	Open   float64
	Close  float64
	High   float64
	Low    float64
	Volume float64
	Time   uint64
}

func GetKlines(symbol string, period string, since uint64, size uint16) (*[]Kline, error) {
	u, _ := url.Parse(DataApiUrl + "kline")
	q := u.Query()
	q.Set("market", symbol)
	q.Set("type", period)
	q.Set("since", strconv.FormatUint(since, 10))
	q.Set("size", strconv.FormatUint(uint64(size), 10))
	u.RawQuery = q.Encode()

	resp, err := doGet(u.String())
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var klines []Kline
	jsonparser.ArrayEach(resp.Body(), func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		time, _ := jsonparser.GetInt(value, "[0]")
		open, _ := jsonparser.GetFloat(value, "[1]")
		high, _ := jsonparser.GetFloat(value, "[2]")
		low, _ := jsonparser.GetFloat(value, "[3]")
		close, _ := jsonparser.GetFloat(value, "[4]")
		volume, _ := jsonparser.GetFloat(value, "[5]")
		klines = append(klines, Kline{Time: uint64(time), Open: open, High: high, Low: low, Close: close, Volume: volume})
	}, "data")

	return &klines, nil
}

func doGet(url string) (*fasthttp.Response, error) {
	req := fasthttp.AcquireRequest()
	req.SetRequestURI(url)
	resp := fasthttp.AcquireResponse()
	client := &fasthttp.Client{}
	err := client.Do(req, resp)
	return resp, errors.WithStack(err)
}
