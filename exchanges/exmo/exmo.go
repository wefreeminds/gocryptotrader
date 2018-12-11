package exmo

import (
	"fmt"
	"log"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/thrasher-/gocryptotrader/common"
	"github.com/thrasher-/gocryptotrader/config"
	"github.com/thrasher-/gocryptotrader/currency/symbol"
	exchange "github.com/thrasher-/gocryptotrader/exchanges"
	"github.com/thrasher-/gocryptotrader/exchanges/request"
	"github.com/thrasher-/gocryptotrader/exchanges/ticker"
)

const (
	exmoAPIURL     = "https://api.exmo.com"
	exmoAPIVersion = "1"

	exmoTrades          = "trades"
	exmoOrderbook       = "order_book"
	exmoTicker          = "ticker"
	exmoPairSettings    = "pair_settings"
	exmoCurrency        = "currency"
	exmoUserInfo        = "user_info"
	exmoOrderCreate     = "order_create"
	exmoOrderCancel     = "order_cancel"
	exmoOpenOrders      = "user_open_orders"
	exmoUserTrades      = "user_trades"
	exmoCancelledOrders = "user_cancelled_orders"
	exmoOrderTrades     = "order_trades"
	exmoRequiredAmount  = "required_amount"
	exmoDepositAddress  = "deposit_address"
	exmoWithdrawCrypt   = "withdraw_crypt"
	exmoGetWithdrawTXID = "withdraw_get_txid"
	exmoExcodeCreate    = "excode_create"
	exmoExcodeLoad      = "excode_load"
	exmoWalletHistory   = "wallet_history"

	// Rate limit: 180 per/minute
	exmoAuthRate   = 180
	exmoUnauthRate = 180
)

// EXMO exchange struct
type EXMO struct {
	exchange.Base
}

// SetDefaults sets the basic defaults for exmo
func (e *EXMO) SetDefaults() {
	e.Name = "EXMO"
	e.Enabled = false
	e.Verbose = false
	e.RESTPollingDelay = 10
	e.APIWithdrawPermissions = exchange.AutoWithdrawCryptoWithSetup
	e.RequestCurrencyPairFormat.Delimiter = "_"
	e.RequestCurrencyPairFormat.Uppercase = true
	e.RequestCurrencyPairFormat.Separator = ","
	e.ConfigCurrencyPairFormat.Delimiter = "_"
	e.ConfigCurrencyPairFormat.Uppercase = true
	e.AssetTypes = []string{ticker.Spot}
	e.SupportsAutoPairUpdating = true
	e.SupportsRESTTickerBatching = true
	e.SupportsRESTAPI = true
	e.SupportsWebsocketAPI = false
	e.Requester = request.New(e.Name,
		request.NewRateLimit(time.Minute, exmoAuthRate),
		request.NewRateLimit(time.Minute, exmoUnauthRate),
		common.NewHTTPClientWithTimeout(exchange.DefaultHTTPTimeout))
	e.APIUrlDefault = exmoAPIURL
	e.APIUrl = e.APIUrlDefault
}

// Setup takes in the supplied exchange configuration details and sets params
func (e *EXMO) Setup(exch config.ExchangeConfig) {
	if !exch.Enabled {
		e.SetEnabled(false)
	} else {
		e.Enabled = true
		e.AuthenticatedAPISupport = exch.AuthenticatedAPISupport
		e.SetAPIKeys(exch.APIKey, exch.APISecret, "", false)
		e.SetHTTPClientTimeout(exch.HTTPTimeout)
		e.SetHTTPClientUserAgent(exch.HTTPUserAgent)
		e.RESTPollingDelay = exch.RESTPollingDelay
		e.Verbose = exch.Verbose
		e.BaseCurrencies = common.SplitStrings(exch.BaseCurrencies, ",")
		e.AvailablePairs = common.SplitStrings(exch.AvailablePairs, ",")
		e.EnabledPairs = common.SplitStrings(exch.EnabledPairs, ",")
		err := e.SetCurrencyPairFormat()
		if err != nil {
			log.Fatal(err)
		}
		err = e.SetAssetTypes()
		if err != nil {
			log.Fatal(err)
		}
		err = e.SetAutoPairDefaults()
		if err != nil {
			log.Fatal(err)
		}
		err = e.SetAPIURL(exch)
		if err != nil {
			log.Fatal(err)
		}
		err = e.SetClientProxyAddress(exch.ProxyAddress)
		if err != nil {
			log.Fatal(err)
		}
	}
}

// GetTrades returns the trades for a symbol or symbols
func (e *EXMO) GetTrades(symbol string) (map[string][]Trades, error) {
	v := url.Values{}
	v.Set("pair", symbol)
	result := make(map[string][]Trades)
	url := fmt.Sprintf("%s/v%s/%s", e.APIUrl, exmoAPIVersion, exmoTrades)

	return result, e.SendHTTPRequest(common.EncodeURLValues(url, v), &result)
}

// GetOrderbook returns the orderbook for a symbol or symbols
func (e *EXMO) GetOrderbook(symbol string) (map[string]Orderbook, error) {
	v := url.Values{}
	v.Set("pair", symbol)
	result := make(map[string]Orderbook)
	url := fmt.Sprintf("%s/v%s/%s", e.APIUrl, exmoAPIVersion, exmoOrderbook)

	return result, e.SendHTTPRequest(common.EncodeURLValues(url, v), &result)
}

// GetTicker returns the ticker for a symbol or symbols
func (e *EXMO) GetTicker(symbol string) (map[string]Ticker, error) {
	v := url.Values{}
	v.Set("pair", symbol)
	result := make(map[string]Ticker)
	url := fmt.Sprintf("%s/v%s/%s", e.APIUrl, exmoAPIVersion, exmoTicker)

	return result, e.SendHTTPRequest(common.EncodeURLValues(url, v), &result)
}

// GetPairSettings returns the pair settings for a symbol or symbols
func (e *EXMO) GetPairSettings() (map[string]PairSettings, error) {
	result := make(map[string]PairSettings)
	url := fmt.Sprintf("%s/v%s/%s", e.APIUrl, exmoAPIVersion, exmoPairSettings)

	return result, e.SendHTTPRequest(url, &result)
}

// GetCurrency returns a list of currencies
func (e *EXMO) GetCurrency() ([]string, error) {
	result := []string{}
	url := fmt.Sprintf("%s/v%s/%s", e.APIUrl, exmoAPIVersion, exmoCurrency)

	return result, e.SendHTTPRequest(url, &result)
}

// GetUserInfo returns the user info
func (e *EXMO) GetUserInfo() (UserInfo, error) {
	var result UserInfo
	err := e.SendAuthenticatedHTTPRequest("POST", exmoUserInfo, url.Values{}, &result)
	return result, err
}

// CreateOrder creates an order
// Params: pair, quantity, price and type
// Type can be buy, sell, market_buy, market_sell, market_buy_total and market_sell_total
func (e *EXMO) CreateOrder(pair, orderType string, price, amount float64) (int64, error) {
	type response struct {
		OrderID int64 `json:"order_id"`
	}

	v := url.Values{}
	v.Set("pair", pair)
	v.Set("type", orderType)
	v.Set("price", strconv.FormatFloat(price, 'f', -1, 64))
	v.Set("quantity", strconv.FormatFloat(amount, 'f', -1, 64))

	var result response
	err := e.SendAuthenticatedHTTPRequest("POST", exmoOrderCreate, v, &result)
	return result.OrderID, err
}

// CancelExistingOrder cancels an order by the orderID
func (e *EXMO) CancelExistingOrder(orderID int64) error {
	v := url.Values{}
	v.Set("order_id", strconv.FormatInt(orderID, 10))
	var result interface{}
	return e.SendAuthenticatedHTTPRequest("POST", exmoOrderCancel, v, &result)
}

// GetOpenOrders returns the users open orders
func (e *EXMO) GetOpenOrders() (map[string]OpenOrders, error) {
	result := make(map[string]OpenOrders)
	err := e.SendAuthenticatedHTTPRequest("POST", exmoOpenOrders, url.Values{}, &result)
	return result, err
}

// GetUserTrades returns the user trades
func (e *EXMO) GetUserTrades(pair, offset, limit string) (map[string][]UserTrades, error) {
	result := make(map[string][]UserTrades)
	v := url.Values{}
	v.Set("pair", pair)

	if offset != "" {
		v.Set("offset", offset)
	}

	if limit != "" {
		v.Set("limit", limit)
	}

	err := e.SendAuthenticatedHTTPRequest("POST", exmoUserTrades, v, &result)
	return result, err
}

// GetCancelledOrders returns a list of cancelled orders
func (e *EXMO) GetCancelledOrders(offset, limit string) ([]CancelledOrder, error) {
	var result []CancelledOrder
	v := url.Values{}

	if offset != "" {
		v.Set("offset", offset)
	}

	if limit != "" {
		v.Set("limit", limit)
	}

	err := e.SendAuthenticatedHTTPRequest("POST", exmoCancelledOrders, v, &result)
	return result, err
}

// GetOrderTrades returns a history of order trade details for the specific orderID
func (e *EXMO) GetOrderTrades(orderID int64) (OrderTrades, error) {
	var result OrderTrades
	v := url.Values{}
	v.Set("order_id", strconv.FormatInt(orderID, 10))

	err := e.SendAuthenticatedHTTPRequest("POST", exmoOrderTrades, v, &result)
	return result, err
}

// GetRequiredAmount calculates the sum of buying a certain amount of currency
// for the particular currency pair
func (e *EXMO) GetRequiredAmount(pair string, amount float64) (RequiredAmount, error) {
	v := url.Values{}
	v.Set("pair", pair)
	v.Set("quantity", strconv.FormatFloat(amount, 'f', -1, 64))
	var result RequiredAmount
	err := e.SendAuthenticatedHTTPRequest("POST", exmoRequiredAmount, v, &result)
	return result, err
}

// GetCryptoDepositAddress returns a list of addresses for cryptocurrency deposits
func (e *EXMO) GetCryptoDepositAddress() (map[string]string, error) {
	result := make(map[string]string)
	err := e.SendAuthenticatedHTTPRequest("POST", exmoDepositAddress, url.Values{}, &result)
	log.Println(reflect.TypeOf(result).String())
	return result, err
}

// WithdrawCryptocurrency withdraws a cryptocurrency from the exchange to the desired address
// NOTE: This API function is available only after request to their tech support team
func (e *EXMO) WithdrawCryptocurrency(currency, address, invoice string, amount float64) (int64, error) {
	type response struct {
		TaskID int64 `json:"task_id,string"`
	}

	v := url.Values{}
	v.Set("currency", currency)
	v.Set("address", address)

	if common.StringToUpper(currency) == "XRP" {
		v.Set(invoice, invoice)
	}

	v.Set("amount", strconv.FormatFloat(amount, 'f', -1, 64))
	var result response
	err := e.SendAuthenticatedHTTPRequest("POST", exmoWithdrawCrypt, v, &result)
	return result.TaskID, err
}

// GetWithdrawTXID gets the result of a withdrawal request
func (e *EXMO) GetWithdrawTXID(taskID int64) (string, error) {
	type response struct {
		Status bool   `json:"status"`
		TXID   string `json:"txid"`
	}

	v := url.Values{}
	v.Set("task_id", strconv.FormatInt(taskID, 10))

	var result response
	err := e.SendAuthenticatedHTTPRequest("POST", exmoGetWithdrawTXID, v, &result)
	return result.TXID, err
}

// ExcodeCreate creates an EXMO coupon
func (e *EXMO) ExcodeCreate(currency string, amount float64) (ExcodeCreate, error) {
	v := url.Values{}
	v.Set("currency", currency)
	v.Set("amount", strconv.FormatFloat(amount, 'f', -1, 64))

	var result ExcodeCreate
	err := e.SendAuthenticatedHTTPRequest("POST", exmoExcodeCreate, v, &result)
	return result, err
}

// ExcodeLoad loads an EXMO coupon
func (e *EXMO) ExcodeLoad(excode string) (ExcodeLoad, error) {
	v := url.Values{}
	v.Set("code", excode)

	var result ExcodeLoad
	err := e.SendAuthenticatedHTTPRequest("POST", exmoExcodeLoad, v, &result)
	return result, err
}

// GetWalletHistory returns the users deposit/withdrawal history
func (e *EXMO) GetWalletHistory(date int64) (WalletHistory, error) {
	v := url.Values{}
	v.Set("date", strconv.FormatInt(date, 10))

	var result WalletHistory
	err := e.SendAuthenticatedHTTPRequest("POST", exmoWalletHistory, v, &result)
	return result, err
}

// SendHTTPRequest sends an unauthenticated HTTP request
func (e *EXMO) SendHTTPRequest(path string, result interface{}) error {
	return e.SendPayload("GET", path, nil, nil, result, false, e.Verbose)
}

// SendAuthenticatedHTTPRequest sends an authenticated HTTP request
func (e *EXMO) SendAuthenticatedHTTPRequest(method, endpoint string, vals url.Values, result interface{}) error {
	if !e.AuthenticatedAPISupport {
		return fmt.Errorf(exchange.WarningAuthenticatedRequestWithoutCredentialsSet, e.Name)
	}

	if e.Nonce.Get() == 0 {
		e.Nonce.Set(time.Now().UnixNano())
	} else {
		e.Nonce.Inc()
	}
	vals.Set("nonce", e.Nonce.String())

	payload := vals.Encode()
	hash := common.GetHMAC(common.HashSHA512, []byte(payload), []byte(e.APISecret))

	if e.Verbose {
		log.Printf("Sending %s request to %s with params %s\n", method, endpoint, payload)
	}

	headers := make(map[string]string)
	headers["Key"] = e.APIKey
	headers["Sign"] = common.HexEncodeToString(hash)
	headers["Content-Type"] = "application/x-www-form-urlencoded"

	path := fmt.Sprintf("%s/v%s/%s", e.APIUrl, exmoAPIVersion, endpoint)

	return e.SendPayload(method, path, headers, strings.NewReader(payload), result, true, e.Verbose)
}

// GetFee returns an estimate of fee based on type of transaction
func (e *EXMO) GetFee(feeBuilder exchange.FeeBuilder) (float64, error) {
	var fee float64
	switch feeBuilder.FeeType {
	case exchange.CryptocurrencyTradeFee:
		fee = e.calculateTradingFee(feeBuilder.PurchasePrice, feeBuilder.Amount)
	case exchange.CryptocurrencyWithdrawalFee:
		fee = getCryptocurrencyWithdrawalFee(feeBuilder.FirstCurrency)
	case exchange.InternationalBankWithdrawalFee:
		fee = getInternationalBankWithdrawalFee(feeBuilder.CurrencyItem, feeBuilder.Amount, feeBuilder.BankTransactionType)
	case exchange.InternationalBankDepositFee:
		fee = getInternationalBankDepositFee(feeBuilder.CurrencyItem, feeBuilder.Amount, feeBuilder.BankTransactionType)
	}

	if fee < 0 {
		fee = 0
	}

	return fee, nil
}

func getCryptocurrencyWithdrawalFee(currency string) float64 {
	return WithdrawalFees[currency]
}

func (e *EXMO) calculateTradingFee(purchasePrice, amount float64) float64 {
	fee := 0.002
	return fee * amount * purchasePrice
}

func calculateTradingFee(purchasePrice, amount float64) float64 {
	fee := 0.002
	return fee * amount * purchasePrice
}

func getInternationalBankWithdrawalFee(currency string, amount float64, bankTransactionType exchange.InternationalBankTransactionType) float64 {
	var fee float64

	switch bankTransactionType {
	case exchange.WireTransfer:
		if currency == symbol.RUB {
			fee = 3200
		} else if currency == symbol.PLN {
			fee = 125
		} else if currency == symbol.TRY {
			fee = 0
		}
	case exchange.PerfectMoney:
		switch currency {
		case symbol.USD:
			fee = 0.01 * amount
		case symbol.EUR:
			fee = 0.0195 * amount
		}
	case exchange.Neteller:
		switch currency {
		case symbol.USD:
			fee = 0.0195 * amount
		case symbol.EUR:
			fee = 0.0195 * amount
		}
	case exchange.AdvCash:
		switch currency {
		case symbol.USD:
			fee = 0.0295 * amount
		case symbol.EUR:
			fee = 0.03 * amount
		case symbol.RUB:
			fee = 0.0195 * amount
		case symbol.UAH:
			fee = 0.0495 * amount
		}
	case exchange.Payeer:
		switch currency {
		case symbol.USD:
			fee = 0.0395 * amount
		case symbol.EUR:
			fee = 0.01 * amount
		case symbol.RUB:
			fee = 0.0595 * amount
		}
	case exchange.Skrill:
		switch currency {
		case symbol.USD:
			fee = 0.0145 * amount
		case symbol.EUR:
			fee = 0.03 * amount
		case symbol.TRY:
			fee = 0
		}
	case exchange.VisaMastercard:
		switch currency {
		case symbol.USD:
			fee = 0.06 * amount
		case symbol.EUR:
			fee = 0.06 * amount
		case symbol.PLN:
			fee = 0.06 * amount
		}
	}

	return fee
}

func getInternationalBankDepositFee(currency string, amount float64, bankTransactionType exchange.InternationalBankTransactionType) float64 {
	var fee float64
	switch bankTransactionType {
	case exchange.WireTransfer:
		if currency == symbol.RUB {
			fee = 1600
		} else if currency == symbol.PLN {
			fee = 30
		} else if currency == symbol.TRY {
			fee = 0
		}
	case exchange.Neteller:
		switch currency {
		case symbol.USD:
			fee = (0.035 * amount) + 0.29
		case symbol.EUR:
			fee = (0.035 * amount) + 0.25
		}
	case exchange.AdvCash:
		switch currency {
		case symbol.USD:
			fee = 0.0295 * amount
		case symbol.EUR:
			fee = 0.01 * amount
		case symbol.RUB:
			fee = 0.0495 * amount
		case symbol.UAH:
			fee = 0.01 * amount
		}
	case exchange.Payeer:
		switch currency {
		case symbol.USD:
			fee = 0.0195 * amount
		case symbol.EUR:
			fee = 0.0295 * amount
		case symbol.RUB:
			fee = 0.0345 * amount
		}
	case exchange.Skrill:
		switch currency {
		case symbol.USD:
			fee = (0.0495 * amount) + 0.36
		case symbol.EUR:
			fee = (0.0295 * amount) + 0.29
		case symbol.PLN:
			fee = (0.035 * amount) + 1.21
		case symbol.TRY:
			fee = 0
		}
	}

	return fee
}
