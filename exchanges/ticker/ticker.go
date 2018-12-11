package ticker

import (
	"errors"
	"strconv"
	"sync"
	"time"

	"github.com/thrasher-/gocryptotrader/exchanges/assets"

	"github.com/thrasher-/gocryptotrader/common"
	"github.com/thrasher-/gocryptotrader/currency/pair"
)

// Const values for the ticker package
const (
	ErrTickerForExchangeNotFound = "Ticker for exchange does not exist."
	ErrPrimaryCurrencyNotFound   = "Error primary currency for ticker not found."
	ErrSecondaryCurrencyNotFound = "Error secondary currency for ticker not found."
)

// Vars for the ticker package
var (
	Tickers []Ticker
	m       sync.Mutex
)

// Price struct stores the currency pair and pricing information
type Price struct {
	Pair         pair.CurrencyPair `json:"Pair"`
	LastUpdated  time.Time         `json:"LastUpdated"`
	CurrencyPair string            `json:"CurrencyPair"`
	Last         float64           `json:"Last"`
	High         float64           `json:"High"`
	Low          float64           `json:"Low"`
	Bid          float64           `json:"Bid"`
	Ask          float64           `json:"Ask"`
	Volume       float64           `json:"Volume"`
	PriceATH     float64           `json:"PriceATH"`
}

// Ticker struct holds the ticker information for a currency pair and type
type Ticker struct {
	Price        map[pair.CurrencyItem]map[pair.CurrencyItem]map[assets.AssetType]Price
	ExchangeName string
}

// PriceToString returns the string version of a stored price field
func (t *Ticker) PriceToString(p pair.CurrencyPair, priceType string, tickerType assets.AssetType) string {
	priceType = common.StringToLower(priceType)

	switch priceType {
	case "last":
		return strconv.FormatFloat(t.Price[p.FirstCurrency][p.SecondCurrency][tickerType].Last, 'f', -1, 64)
	case "high":
		return strconv.FormatFloat(t.Price[p.FirstCurrency][p.SecondCurrency][tickerType].High, 'f', -1, 64)
	case "low":
		return strconv.FormatFloat(t.Price[p.FirstCurrency][p.SecondCurrency][tickerType].Low, 'f', -1, 64)
	case "bid":
		return strconv.FormatFloat(t.Price[p.FirstCurrency][p.SecondCurrency][tickerType].Bid, 'f', -1, 64)
	case "ask":
		return strconv.FormatFloat(t.Price[p.FirstCurrency][p.SecondCurrency][tickerType].Ask, 'f', -1, 64)
	case "volume":
		return strconv.FormatFloat(t.Price[p.FirstCurrency][p.SecondCurrency][tickerType].Volume, 'f', -1, 64)
	case "ath":
		return strconv.FormatFloat(t.Price[p.FirstCurrency][p.SecondCurrency][tickerType].PriceATH, 'f', -1, 64)
	default:
		return ""
	}
}

// GetTicker checks and returns a requested ticker if it exists
func GetTicker(exchange string, p pair.CurrencyPair, tickerType assets.AssetType) (Price, error) {
	ticker, err := GetTickerByExchange(exchange)
	if err != nil {
		return Price{}, err
	}

	if !FirstCurrencyExists(exchange, p.FirstCurrency) {
		return Price{}, errors.New(ErrPrimaryCurrencyNotFound)
	}

	if !SecondCurrencyExists(exchange, p) {
		return Price{}, errors.New(ErrSecondaryCurrencyNotFound)
	}

	return ticker.Price[p.FirstCurrency][p.SecondCurrency][tickerType], nil
}

// GetTickerByExchange returns an exchange Ticker
func GetTickerByExchange(exchange string) (*Ticker, error) {
	m.Lock()
	defer m.Unlock()
	for _, y := range Tickers {
		if y.ExchangeName == exchange {
			return &y, nil
		}
	}
	return nil, errors.New(ErrTickerForExchangeNotFound)
}

// FirstCurrencyExists checks to see if the first currency of the Price map
// exists
func FirstCurrencyExists(exchange string, currency pair.CurrencyItem) bool {
	m.Lock()
	defer m.Unlock()
	for _, y := range Tickers {
		if y.ExchangeName == exchange {
			if _, ok := y.Price[currency]; ok {
				return true
			}
		}
	}
	return false
}

// SecondCurrencyExists checks to see if the second currency of the Price map
// exists
func SecondCurrencyExists(exchange string, p pair.CurrencyPair) bool {
	m.Lock()
	defer m.Unlock()
	for _, y := range Tickers {
		if y.ExchangeName == exchange {
			if _, ok := y.Price[p.FirstCurrency]; ok {
				if _, ok := y.Price[p.FirstCurrency][p.SecondCurrency]; ok {
					return true
				}
			}
		}
	}
	return false
}

// CreateNewTicker creates a new Ticker
func CreateNewTicker(exchangeName string, p pair.CurrencyPair, tickerNew Price, tickerType assets.AssetType) Ticker {
	m.Lock()
	defer m.Unlock()
	ticker := Ticker{}
	ticker.ExchangeName = exchangeName
	ticker.Price = make(map[pair.CurrencyItem]map[pair.CurrencyItem]map[assets.AssetType]Price)
	a := make(map[pair.CurrencyItem]map[assets.AssetType]Price)
	b := make(map[assets.AssetType]Price)
	b[tickerType] = tickerNew
	a[p.SecondCurrency] = b
	ticker.Price[p.FirstCurrency] = a
	Tickers = append(Tickers, ticker)
	return ticker
}

// ProcessTicker processes incoming tickers, creating or updating the Tickers
// list
func ProcessTicker(exchangeName string, p pair.CurrencyPair, tickerNew Price, tickerType assets.AssetType) {
	if tickerNew.Pair.Pair() == "" {
		// set Pair if not set
		tickerNew.Pair = p
	}

	tickerNew.CurrencyPair = p.Pair().String()
	tickerNew.LastUpdated = time.Now()

	ticker, err := GetTickerByExchange(exchangeName)
	if err != nil {
		CreateNewTicker(exchangeName, p, tickerNew, tickerType)
		return
	}

	if FirstCurrencyExists(exchangeName, p.FirstCurrency) {
		m.Lock()
		a := make(map[assets.AssetType]Price)
		a[tickerType] = tickerNew
		ticker.Price[p.FirstCurrency][p.SecondCurrency] = a
		m.Unlock()
		return
	}

	m.Lock()
	a := make(map[pair.CurrencyItem]map[assets.AssetType]Price)
	b := make(map[assets.AssetType]Price)
	b[tickerType] = tickerNew
	a[p.SecondCurrency] = b
	ticker.Price[p.FirstCurrency] = a
	m.Unlock()
}
