package bitstamp

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/thrasher-/gocryptotrader/common"
	"github.com/thrasher-/gocryptotrader/currency/symbol"
	"github.com/thrasher-/gocryptotrader/exchanges"
)

const (
	bitstampAPIURL                = "https://www.bitstamp.net/api"
	bitstampAPIVersion            = "2"
	bitstampAPITicker             = "ticker"
	bitstampAPITickerHourly       = "ticker_hour"
	bitstampAPIOrderbook          = "order_book"
	bitstampAPITransactions       = "transactions"
	bitstampAPIEURUSD             = "eur_usd"
	bitstampAPIBalance            = "balance"
	bitstampAPIUserTransactions   = "user_transactions"
	bitstampAPIOpenOrders         = "open_orders"
	bitstampAPIOrderStatus        = "order_status"
	bitstampAPICancelOrder        = "cancel_order"
	bitstampAPICancelAllOrders    = "cancel_all_orders"
	bitstampAPIBuy                = "buy"
	bitstampAPISell               = "sell"
	bitstampAPIMarket             = "market"
	bitstampAPIWithdrawalRequests = "withdrawal_requests"
	bitstampAPIBitcoinWithdrawal  = "bitcoin_withdrawal"
	bitstampAPILTCWithdrawal      = "ltc_withdrawal"
	bitstampAPIETHWithdrawal      = "eth_withdrawal"
	bitstampAPIBitcoinDeposit     = "bitcoin_deposit_address"
	bitstampAPILitecoinDeposit    = "ltc_address"
	bitstampAPIEthereumDeposit    = "eth_address"
	bitstampAPIUnconfirmedBitcoin = "unconfirmed_btc"
	bitstampAPITransferToMain     = "transfer-to-main"
	bitstampAPITransferFromMain   = "transfer-from-main"
	bitstampAPIXrpWithdrawal      = "xrp_withdrawal"
	bitstampAPIXrpDeposit         = "xrp_address"
	bitstampAPIReturnType         = "string"
	bitstampAPITradingPairsInfo   = "trading-pairs-info"

	bitstampAuthRate   = 600
	bitstampUnauthRate = 600
)

// Bitstamp is the overarching type across the bitstamp package
type Bitstamp struct {
	exchange.Base
	Balance       Balances
	WebsocketConn WebsocketConn
}

// GetFee returns an estimate of fee based on type of transaction
func (b *Bitstamp) GetFee(feeBuilder exchange.FeeBuilder) (float64, error) {
	var fee float64

	switch feeBuilder.FeeType {
	case exchange.CryptocurrencyTradeFee:
		var err error

		b.Balance, err = b.GetBalance()
		if err != nil {
			return 0, err
		}
		fee = b.CalculateTradingFee(feeBuilder.FirstCurrency+feeBuilder.SecondCurrency, feeBuilder.PurchasePrice, feeBuilder.Amount)
	case exchange.CyptocurrencyDepositFee:
		fee = 0
	case exchange.InternationalBankDepositFee:
		fee = getInternationalBankDepositFee(feeBuilder.Amount)
	case exchange.InternationalBankWithdrawalFee:
		fee = getInternationalBankWithdrawalFee(feeBuilder.Amount)
	}
	if fee < 0 {
		fee = 0
	}
	return fee, nil
}

// getInternationalBankWithdrawalFee returns international withdrawal fee
func getInternationalBankWithdrawalFee(amount float64) float64 {
	fee := amount * 0.0009

	if fee < 15 {
		return 15
	}
	return fee
}

// getInternationalBankDepositFee returns international deposit fee
func getInternationalBankDepositFee(amount float64) float64 {
	fee := amount * 0.0005

	if fee < 7.5 {
		return 7.5
	}
	if fee > 300 {
		return 300
	}
	return fee
}

// CalculateTradingFee returns fee on a currency pair
func (b *Bitstamp) CalculateTradingFee(currency string, purchasePrice float64, amount float64) float64 {
	var fee float64

	switch currency {
	case symbol.BTC + symbol.USD:
		fee = b.Balance.BTCUSDFee
	case symbol.BTC + symbol.EUR:
		fee = b.Balance.BTCEURFee
	case symbol.XRP + symbol.EUR:
		fee = b.Balance.XRPEURFee
	case symbol.XRP + symbol.USD:
		fee = b.Balance.XRPUSDFee
	case symbol.EUR + symbol.USD:
		fee = b.Balance.EURUSDFee
	default:
		fee = 0
	}
	return fee * purchasePrice * amount
}

// GetTicker returns ticker information
func (b *Bitstamp) GetTicker(currency string, hourly bool) (Ticker, error) {
	response := Ticker{}
	tickerEndpoint := bitstampAPITicker

	if hourly {
		tickerEndpoint = bitstampAPITickerHourly
	}

	path := fmt.Sprintf(
		"%s/v%s/%s/%s/",
		b.API.Endpoints.URL,
		bitstampAPIVersion,
		tickerEndpoint,
		common.StringToLower(currency),
	)
	return response, b.SendHTTPRequest(path, &response)
}

// GetOrderbook Returns a JSON dictionary with "bids" and "asks". Each is a list
// of open orders and each order is represented as a list holding the price and
//the amount.
func (b *Bitstamp) GetOrderbook(currency string) (Orderbook, error) {
	type response struct {
		Timestamp int64      `json:"timestamp,string"`
		Bids      [][]string `json:"bids"`
		Asks      [][]string `json:"asks"`
	}
	resp := response{}

	path := fmt.Sprintf(
		"%s/v%s/%s/%s/",
		b.API.Endpoints.URL,
		bitstampAPIVersion,
		bitstampAPIOrderbook,
		common.StringToLower(currency),
	)

	err := b.SendHTTPRequest(path, &resp)
	if err != nil {
		return Orderbook{}, err
	}

	orderbook := Orderbook{}
	orderbook.Timestamp = resp.Timestamp

	for _, x := range resp.Bids {
		price, err := strconv.ParseFloat(x[0], 64)
		if err != nil {
			log.Println(err)
			continue
		}
		amount, err := strconv.ParseFloat(x[1], 64)
		if err != nil {
			log.Println(err)
			continue
		}
		orderbook.Bids = append(orderbook.Bids, OrderbookBase{price, amount})
	}

	for _, x := range resp.Asks {
		price, err := strconv.ParseFloat(x[0], 64)
		if err != nil {
			log.Println(err)
			continue
		}
		amount, err := strconv.ParseFloat(x[1], 64)
		if err != nil {
			log.Println(err)
			continue
		}
		orderbook.Asks = append(orderbook.Asks, OrderbookBase{price, amount})
	}

	return orderbook, nil
}

// GetTradingPairs returns a list of trading pairs which Bitstamp
// currently supports
func (b *Bitstamp) GetTradingPairs() ([]TradingPair, error) {
	var result []TradingPair

	path := fmt.Sprintf("%s/v%s/%s",
		b.API.Endpoints.URL,
		bitstampAPIVersion,
		bitstampAPITradingPairsInfo)

	return result, b.SendHTTPRequest(path, &result)
}

// GetTransactions returns transaction information
// value paramater ["time"] = "minute", "hour", "day" will collate your
// response into time intervals. Implementation of value in test code.
func (b *Bitstamp) GetTransactions(currencyPair string, values url.Values) ([]Transactions, error) {
	transactions := []Transactions{}
	path := common.EncodeURLValues(
		fmt.Sprintf(
			"%s/v%s/%s/%s/",
			b.API.Endpoints.URL,
			bitstampAPIVersion,
			bitstampAPITransactions,
			common.StringToLower(currencyPair),
		),
		values,
	)

	return transactions, b.SendHTTPRequest(path, &transactions)
}

// GetEURUSDConversionRate returns the conversion rate between Euro and USD
func (b *Bitstamp) GetEURUSDConversionRate() (EURUSDConversionRate, error) {
	rate := EURUSDConversionRate{}
	path := fmt.Sprintf("%s/%s", b.API.Endpoints.URL, bitstampAPIEURUSD)

	return rate, b.SendHTTPRequest(path, &rate)
}

// GetBalance returns full balance of currency held on the exchange
func (b *Bitstamp) GetBalance() (Balances, error) {
	balance := Balances{}
	path := fmt.Sprintf("%s/%s", b.API.Endpoints.URL, bitstampAPIBalance)

	return balance, b.SendHTTPRequest(path, &balance)
}

// GetUserTransactions returns an array of transactions
func (b *Bitstamp) GetUserTransactions(currencyPair string) ([]UserTransactions, error) {
	type Response struct {
		Date    string      `json:"datetime"`
		TransID int64       `json:"id"`
		Type    int         `json:"type,string"`
		USD     interface{} `json:"usd"`
		EUR     float64     `json:"eur"`
		XRP     float64     `json:"xrp"`
		BTC     interface{} `json:"btc"`
		BTCUSD  interface{} `json:"btc_usd"`
		Fee     float64     `json:"fee,string"`
		OrderID int64       `json:"order_id"`
	}
	response := []Response{}

	if currencyPair != "" {
		if err := b.SendAuthenticatedHTTPRequest(bitstampAPIUserTransactions, true, url.Values{}, &response); err != nil {
			return nil, err
		}
	} else {
		if err := b.SendAuthenticatedHTTPRequest(bitstampAPIUserTransactions+"/"+currencyPair, true, url.Values{}, &response); err != nil {
			return nil, err
		}
	}

	transactions := []UserTransactions{}

	for _, y := range response {
		tx := UserTransactions{}
		tx.Date = y.Date
		tx.TransID = y.TransID
		tx.Type = y.Type

		/* Hack due to inconsistent JSON values... */
		varType := reflect.TypeOf(y.USD).String()
		if varType == bitstampAPIReturnType {
			tx.USD, _ = strconv.ParseFloat(y.USD.(string), 64)
		} else {
			tx.USD = y.USD.(float64)
		}

		tx.EUR = y.EUR
		tx.XRP = y.XRP

		varType = reflect.TypeOf(y.BTC).String()
		if varType == bitstampAPIReturnType {
			tx.BTC, _ = strconv.ParseFloat(y.BTC.(string), 64)
		} else {
			tx.BTC = y.BTC.(float64)
		}

		varType = reflect.TypeOf(y.BTCUSD).String()
		if varType == bitstampAPIReturnType {
			tx.BTCUSD, _ = strconv.ParseFloat(y.BTCUSD.(string), 64)
		} else {
			tx.BTCUSD = y.BTCUSD.(float64)
		}

		tx.Fee = y.Fee
		tx.OrderID = y.OrderID
		transactions = append(transactions, tx)
	}

	return transactions, nil
}

// GetOpenOrders returns all open orders on the exchange
func (b *Bitstamp) GetOpenOrders(currencyPair string) ([]Order, error) {
	resp := []Order{}
	path := fmt.Sprintf(
		"%s/%s", bitstampAPIOpenOrders, common.StringToLower(currencyPair),
	)

	return resp, b.SendAuthenticatedHTTPRequest(path, true, nil, &resp)
}

// GetOrderStatus returns an the status of an order by its ID
func (b *Bitstamp) GetOrderStatus(OrderID int64) (OrderStatus, error) {
	resp := OrderStatus{}
	req := url.Values{}
	req.Add("id", strconv.FormatInt(OrderID, 10))

	return resp,
		b.SendAuthenticatedHTTPRequest(bitstampAPIOrderStatus, false, req, &resp)
}

// CancelExistingOrder cancels order by ID
func (b *Bitstamp) CancelExistingOrder(OrderID int64) (bool, error) {
	result := false
	var req = url.Values{}
	req.Add("id", strconv.FormatInt(OrderID, 10))

	return result,
		b.SendAuthenticatedHTTPRequest(bitstampAPICancelOrder, true, req, &result)
}

// CancelAllExistingOrders cancels all open orders on the exchange
func (b *Bitstamp) CancelAllExistingOrders() (bool, error) {
	result := false

	return result,
		b.SendAuthenticatedHTTPRequest(bitstampAPICancelAllOrders, false, nil, &result)
}

// PlaceOrder places an order on the exchange.
func (b *Bitstamp) PlaceOrder(currencyPair string, price float64, amount float64, buy, market bool) (Order, error) {
	var req = url.Values{}
	req.Add("amount", strconv.FormatFloat(amount, 'f', -1, 64))
	req.Add("price", strconv.FormatFloat(price, 'f', -1, 64))
	response := Order{}
	orderType := bitstampAPIBuy

	if !buy {
		orderType = bitstampAPISell
	}

	path := fmt.Sprintf("%s/%s", orderType, common.StringToLower(currencyPair))

	if market {
		path = fmt.Sprintf("%s/%s/%s", orderType, bitstampAPIMarket, common.StringToLower(currencyPair))
	}

	return response,
		b.SendAuthenticatedHTTPRequest(path, true, req, &response)
}

// GetWithdrawalRequests returns withdrawal requests for the account
// timedelta - positive integer with max value 50000000 which returns requests
// from number of seconds ago to now.
func (b *Bitstamp) GetWithdrawalRequests(timedelta int64) ([]WithdrawalRequests, error) {
	resp := []WithdrawalRequests{}
	if timedelta > 50000000 || timedelta < 0 {
		return resp, errors.New("time delta exceeded, max: 50000000 min: 0")
	}

	value := url.Values{}
	value.Set("timedelta", strconv.FormatInt(timedelta, 10))

	if timedelta == 0 {
		value = url.Values{}
	}

	return resp,
		b.SendAuthenticatedHTTPRequest(bitstampAPIWithdrawalRequests, false, value, &resp)
}

// CryptoWithdrawal withdraws a cryptocurrency into a supplied wallet, returns ID
// amount - The amount you want withdrawn
// address - The wallet address of the cryptocurrency
// symbol - the type of crypto ie "ltc", "btc", "eth"
// destTag - only for XRP  default to ""
// instant - only for bitcoins
func (b *Bitstamp) CryptoWithdrawal(amount float64, address, symbol, destTag string, instant bool) (string, error) {
	var req = url.Values{}
	req.Add("amount", strconv.FormatFloat(amount, 'f', -1, 64))
	req.Add("address", address)

	type response struct {
		ID string `json:"id"`
	}
	resp := response{}

	switch common.StringToLower(symbol) {
	case "btc":
		if instant {
			req.Add("instant", "1")
		} else {
			req.Add("instant", "0")
		}
		return resp.ID,
			b.SendAuthenticatedHTTPRequest(bitstampAPIBitcoinWithdrawal, false, req, &resp)
	case "ltc":
		return resp.ID,
			b.SendAuthenticatedHTTPRequest(bitstampAPILTCWithdrawal, true, req, &resp)
	case "eth":
		return resp.ID,
			b.SendAuthenticatedHTTPRequest(bitstampAPIETHWithdrawal, true, req, &resp)
	case "xrp":
		if destTag != "" {
			req.Add("destination_tag", destTag)
		}
		return resp.ID,
			b.SendAuthenticatedHTTPRequest(bitstampAPIXrpWithdrawal, true, req, &resp)
	}
	return resp.ID,
		errors.New("incorrect symbol")
}

// GetCryptoDepositAddress returns a depositing address by crypto
// crypto - example "btc", "ltc", "eth", or "xrp"
func (b *Bitstamp) GetCryptoDepositAddress(crypto string) (string, error) {
	type response struct {
		Address string `json:"address"`
	}
	resp := response{}

	switch common.StringToLower(crypto) {
	case "btc":
		return resp.Address,
			b.SendAuthenticatedHTTPRequest(bitstampAPIBitcoinDeposit, false, nil, &resp.Address)
	case "ltc":
		return resp.Address,
			b.SendAuthenticatedHTTPRequest(bitstampAPILitecoinDeposit, true, nil, &resp)
	case "eth":
		return resp.Address,
			b.SendAuthenticatedHTTPRequest(bitstampAPIEthereumDeposit, true, nil, &resp)
	case "xrp":
		return resp.Address,
			b.SendAuthenticatedHTTPRequest(bitstampAPIXrpDeposit, true, nil, &resp)
	}

	return resp.Address, errors.New("incorrect cryptocurrency string")
}

// GetUnconfirmedBitcoinDeposits returns unconfirmed transactions
func (b *Bitstamp) GetUnconfirmedBitcoinDeposits() ([]UnconfirmedBTCTransactions, error) {
	response := []UnconfirmedBTCTransactions{}

	return response,
		b.SendAuthenticatedHTTPRequest(bitstampAPIUnconfirmedBitcoin, false, nil, &response)
}

// TransferAccountBalance transfers funds from either a main or sub account
// amount - to transfers
// currency - which currency to transfer
// subaccount - name of account
// toMain - bool either to or from account
func (b *Bitstamp) TransferAccountBalance(amount float64, currency, subAccount string, toMain bool) (bool, error) {
	var req = url.Values{}
	req.Add("amount", strconv.FormatFloat(amount, 'f', -1, 64))
	req.Add("currency", currency)
	req.Add("subAccount", subAccount)

	path := bitstampAPITransferToMain
	if !toMain {
		path = bitstampAPITransferFromMain
	}

	err := b.SendAuthenticatedHTTPRequest(path, true, req, nil)
	if err != nil {
		return false, err
	}

	return true, nil
}

// SendHTTPRequest sends an unauthenticated HTTP request
func (b *Bitstamp) SendHTTPRequest(path string, result interface{}) error {
	return b.SendPayload("GET", path, nil, nil, result, false, b.Verbose)
}

// SendAuthenticatedHTTPRequest sends an authenticated request
func (b *Bitstamp) SendAuthenticatedHTTPRequest(path string, v2 bool, values url.Values, result interface{}) (err error) {
	if !b.API.AuthenticatedSupport {
		return fmt.Errorf(exchange.WarningAuthenticatedRequestWithoutCredentialsSet, b.Name)
	}

	if b.Nonce.Get() == 0 {
		b.Nonce.Set(time.Now().UnixNano())
	} else {
		b.Nonce.Inc()
	}

	if values == nil {
		values = url.Values{}
	}

	values.Set("key", b.API.Credentials.Key)
	values.Set("nonce", b.Nonce.String())
	hmac := common.GetHMAC(common.HashSHA256, []byte(b.Nonce.String()+b.API.Credentials.ClientID+b.API.Credentials.Key), []byte(b.API.Credentials.Secret))
	values.Set("signature", common.StringToUpper(common.HexEncodeToString(hmac)))

	if v2 {
		path = fmt.Sprintf("%s/v%s/%s/", b.API.Endpoints.URL, bitstampAPIVersion, path)
	} else {
		path = fmt.Sprintf("%s/%s/", b.API.Endpoints.URL, path)
	}

	if b.Verbose {
		log.Println("Sending POST request to " + path)
	}

	headers := make(map[string]string)
	headers["Content-Type"] = "application/x-www-form-urlencoded"

	encodedValues := values.Encode()
	readerValues := strings.NewReader(encodedValues)
	return b.SendPayload("POST", path, headers, readerValues, result, true, b.Verbose)
}
