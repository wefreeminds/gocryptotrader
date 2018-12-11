package bithumb

import (
	"bytes"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"time"

	"github.com/thrasher-/gocryptotrader/common"
	"github.com/thrasher-/gocryptotrader/currency/symbol"
	exchange "github.com/thrasher-/gocryptotrader/exchanges"
)

const (
	apiURL = "https://api.bithumb.com"

	noError = "0000"

	// Public API
	requestsPerSecondPublicAPI = 20

	publicTicker             = "/public/ticker/"
	publicOrderBook          = "/public/orderbook/"
	publicTransactionHistory = "/public/transaction_history/"

	// Private API
	requestsPerSecondPrivateAPI = 10

	privateAccInfo     = "/info/account"
	privateAccBalance  = "/info/balance"
	privateWalletAdd   = "/info/wallet_address"
	privateTicker      = "/info/ticker"
	privateOrders      = "/info/orders"
	privateUserTrans   = "/info/user_transactions"
	privatePlaceTrade  = "/trade/place"
	privateOrderDetail = "/info/order_detail"
	privateCancelTrade = "/trade/cancel"
	privateBTCWithdraw = "/trade/btc_withdrawal"
	privateKRWDeposit  = "/trade/krw_deposit"
	privateKRWWithdraw = "/trade/krw_withdrawal"
	privateMarketBuy   = "/trade/market_buy"
	privateMarketSell  = "/trade/market_sell"

	bithumbAuthRate   = 10
	bithumbUnauthRate = 20
)

// Bithumb is the overarching type across the Bithumb package
type Bithumb struct {
	exchange.Base
}

// GetTradablePairs returns a list of tradable currencies
func (b *Bithumb) GetTradablePairs() ([]string, error) {
	result, err := b.GetAllTickers()
	if err != nil {
		return nil, err
	}

	var currencies []string
	for x := range result {
		currencies = append(currencies, x)
	}
	return currencies, nil
}

// GetTicker returns ticker information
//
// symbol e.g. "btc"
func (b *Bithumb) GetTicker(symbol string) (Ticker, error) {
	response := Ticker{}
	path := fmt.Sprintf("%s%s%s", b.API.Endpoints.URL, publicTicker, common.StringToUpper(symbol))

	err := b.SendHTTPRequest(path, &response)
	if err != nil {
		return response, err
	}

	if response.Status != noError {
		return response, errors.New(response.Message)
	}

	return response, nil
}

// GetAllTickers returns all ticker information
func (b *Bithumb) GetAllTickers() (map[string]Ticker, error) {
	type Response struct {
		ActionStatus
		Data map[string]interface{}
	}

	response := Response{}
	path := fmt.Sprintf("%s%s%s", b.API.Endpoints.URL, publicTicker, "all")

	err := b.SendHTTPRequest(path, &response)
	if err != nil {
		return nil, err
	}

	if response.Status != noError {
		return nil, errors.New(response.Message)
	}

	result := make(map[string]Ticker)
	for k, v := range response.Data {
		if k == "date" {
			continue
		}

		if reflect.TypeOf(v).String() != "map[string]interface {}" {
			continue
		}

		data := v.(map[string]interface{})
		var t Ticker
		t.AveragePrice, _ = strconv.ParseFloat(data["average_price"].(string), 64)
		t.BuyPrice, _ = strconv.ParseFloat(data["buy_price"].(string), 64)
		t.ClosingPrice, _ = strconv.ParseFloat(data["closing_price"].(string), 64)
		t.MaxPrice, _ = strconv.ParseFloat(data["max_price"].(string), 64)
		t.MinPrice, _ = strconv.ParseFloat(data["min_price"].(string), 64)
		t.OpeningPrice, _ = strconv.ParseFloat(data["opening_price"].(string), 64)
		t.SellPrice, _ = strconv.ParseFloat(data["sell_price"].(string), 64)
		t.UnitsTraded, _ = strconv.ParseFloat(data["units_traded"].(string), 64)
		t.Volume1Day, _ = strconv.ParseFloat(data["volume_1day"].(string), 64)
		t.Volume7Day, _ = strconv.ParseFloat(data["volume_7day"].(string), 64)
		result[k] = t

	}
	return result, nil
}

// GetOrderBook returns current orderbook
//
// symbol e.g. "btc"
func (b *Bithumb) GetOrderBook(symbol string) (Orderbook, error) {
	response := Orderbook{}
	path := fmt.Sprintf("%s%s%s", b.API.Endpoints.URL, publicOrderBook, common.StringToUpper(symbol))

	err := b.SendHTTPRequest(path, &response)
	if err != nil {
		return response, err
	}

	if response.Status != noError {
		return response, errors.New(response.Message)
	}

	return response, nil
}

// GetTransactionHistory returns recent transactions
//
// symbol e.g. "btc"
func (b *Bithumb) GetTransactionHistory(symbol string) (TransactionHistory, error) {
	response := TransactionHistory{}
	path := fmt.Sprintf("%s%s%s", b.API.Endpoints.URL, publicTransactionHistory, common.StringToUpper(symbol))

	err := b.SendHTTPRequest(path, &response)
	if err != nil {
		return response, err
	}

	if response.Status != noError {
		return response, errors.New(response.Message)
	}

	return response, nil
}

// GetAccountInformation returns account information
func (b *Bithumb) GetAccountInformation() (Account, error) {
	response := Account{}

	err := b.SendAuthenticatedHTTPRequest(privateAccInfo, nil, &response)
	if err != nil {
		return response, err
	}

	if response.Status != noError {
		return response, errors.New(response.Message)
	}
	return response, nil
}

// GetAccountBalance returns customer wallet information
func (b *Bithumb) GetAccountBalance() (Balance, error) {
	response := Balance{}

	err := b.SendAuthenticatedHTTPRequest(privateAccBalance, nil, &response)
	if err != nil {
		return response, err
	}

	if response.Status != noError {
		return response, errors.New(response.Message)
	}
	return response, nil
}

// GetWalletAddress returns customer wallet address
//
// currency e.g. btc, ltc or "", will default to btc without currency specified
func (b *Bithumb) GetWalletAddress(currency string) (WalletAddressRes, error) {
	response := WalletAddressRes{}
	params := url.Values{}
	params.Set("currency", common.StringToUpper(currency))

	err := b.SendAuthenticatedHTTPRequest(privateWalletAdd, params, &response)
	if err != nil {
		return response, err
	}

	if response.Status != noError {
		return response, errors.New(response.Message)
	}
	return response, nil
}

// GetLastTransaction returns customer last transaction
func (b *Bithumb) GetLastTransaction() (LastTransactionTicker, error) {
	response := LastTransactionTicker{}

	err := b.SendAuthenticatedHTTPRequest(privateTicker, nil, &response)
	if err != nil {
		return response, err
	}

	if response.Status != noError {
		return response, errors.New(response.Message)
	}
	return response, nil
}

// GetOrders returns order list
//
// orderID: order number registered for purchase/sales
// transactionType: transaction type(bid : purchase, ask : sell)
// count: Value : 1 ~1000 (default : 100)
// after: YYYY-MM-DD hh:mm:ss's UNIX Timestamp
// (2014-11-28 16:40:01 = 1417160401000)
func (b *Bithumb) GetOrders(orderID, transactionType, count, after, currency string) (Orders, error) {
	response := Orders{}

	params := url.Values{}
	params.Set("order_id", orderID)
	params.Set("type", transactionType)
	params.Set("count", count)
	params.Set("after", after)
	params.Set("currency", common.StringToUpper(currency))

	err := b.SendAuthenticatedHTTPRequest(privateOrders, params, &response)
	if err != nil {
		return response, err
	}

	if response.Status != noError {
		return response, errors.New(response.Message)
	}
	return response, nil
}

// GetUserTransactions returns customer transactions
func (b *Bithumb) GetUserTransactions() (UserTransactions, error) {
	response := UserTransactions{}

	err := b.SendAuthenticatedHTTPRequest(privateUserTrans, nil, &response)
	if err != nil {
		return response, err
	}

	if response.Status != noError {
		return response, errors.New(response.Message)
	}
	return response, nil
}

// PlaceTrade executes a trade order
//
// orderCurrency: BTC, ETH, DASH, LTC, ETC, XRP, BCH, XMR, ZEC, QTUM, BTG, EOS
// (default value: BTC)
// transactionType: Transaction type(bid : purchase, ask : sales)
// units: Order quantity
// price: Transaction amount per currency
func (b *Bithumb) PlaceTrade(orderCurrency, transactionType string, units float64, price int64) (OrderPlace, error) {
	response := OrderPlace{}

	params := url.Values{}
	params.Set("order_currency", common.StringToUpper(orderCurrency))
	params.Set("Payment_currency", "KRW")
	params.Set("type", common.StringToUpper(transactionType))
	params.Set("units", strconv.FormatFloat(units, 'f', -1, 64))
	params.Set("price", strconv.FormatInt(price, 10))

	err := b.SendAuthenticatedHTTPRequest(privatePlaceTrade, params, &response)
	if err != nil {
		return response, err
	}

	if response.Status != noError {
		return response, errors.New(response.Message)
	}
	return response, nil
}

// GetOrderDetails returns specific order details
//
// orderID: Order number registered for purchase/sales
// transactionType: Transaction type(bid : purchase, ask : sales)
// currency: BTC, ETH, DASH, LTC, ETC, XRP, BCH, XMR, ZEC, QTUM, BTG, EOS
// (default value: BTC)
func (b *Bithumb) GetOrderDetails(orderID, transactionType, currency string) (OrderDetails, error) {
	response := OrderDetails{}

	params := url.Values{}
	params.Set("order_id", common.StringToUpper(orderID))
	params.Set("type", common.StringToUpper(transactionType))
	params.Set("currency", common.StringToUpper(currency))

	err := b.SendAuthenticatedHTTPRequest(privateOrderDetail, params, &response)
	if err != nil {
		return response, err
	}

	if response.Status != noError {
		return response, errors.New(response.Message)
	}
	return response, nil
}

// CancelTrade cancels a customer purchase/sales transaction
// transactionType: Transaction type(bid : purchase, ask : sales)
// orderID: Order number registered for purchase/sales
// currency: BTC, ETH, DASH, LTC, ETC, XRP, BCH, XMR, ZEC, QTUM, BTG, EOS
// (default value: BTC)
func (b *Bithumb) CancelTrade(transactionType, orderID, currency string) (ActionStatus, error) {
	response := ActionStatus{}

	params := url.Values{}
	params.Set("order_id", common.StringToUpper(orderID))
	params.Set("type", common.StringToUpper(transactionType))
	params.Set("currency", common.StringToUpper(currency))

	err := b.SendAuthenticatedHTTPRequest(privateCancelTrade, nil, &response)
	if err != nil {
		return response, err
	}

	if response.Status != noError {
		return response, errors.New(response.Message)
	}
	return response, nil
}

// WithdrawCrypto withdraws a customer currency to an address
//
// address: Currency withdrawing address
// destination: Currency withdrawal Destination Tag (when withdraw XRP) OR
// Currency withdrawal Payment Id (when withdraw XMR)
// currency: BTC, ETH, DASH, LTC, ETC, XRP, BCH, XMR, ZEC, QTUM
// (default value: BTC)
// units: Quantity to withdraw currency
func (b *Bithumb) WithdrawCrypto(address, destination, currency string, units float64) (ActionStatus, error) {
	response := ActionStatus{}

	params := url.Values{}
	params.Set("address", address)
	params.Set("destination", destination)
	params.Set("currency", common.StringToUpper(currency))
	params.Set("units", strconv.FormatFloat(units, 'f', -1, 64))

	err := b.SendAuthenticatedHTTPRequest(privateBTCWithdraw, params, &response)
	if err != nil {
		return response, err
	}

	if response.Status != noError {
		return response, errors.New(response.Message)
	}
	return response, nil
}

// RequestKRWDepositDetails returns Bithumb banking details for deposit
// information
func (b *Bithumb) RequestKRWDepositDetails() (KRWDeposit, error) {
	response := KRWDeposit{}

	err := b.SendAuthenticatedHTTPRequest(privateKRWDeposit, nil, &response)
	if err != nil {
		return response, err
	}

	if response.Status != noError {
		return response, errors.New(response.Message)
	}
	return response, nil
}

// RequestKRWWithdraw allows a customer KRW withdrawal request
//
// bank: Bankcode with bank name e.g. (bankcode)_(bankname)
// account: Withdrawing bank account number
// price: 	Withdrawing amount
func (b *Bithumb) RequestKRWWithdraw(bank, account string, price int64) (ActionStatus, error) {
	response := ActionStatus{}

	params := url.Values{}
	params.Set("bank", bank)
	params.Set("account", account)
	params.Set("price", strconv.FormatInt(price, 10))

	err := b.SendAuthenticatedHTTPRequest(privateKRWWithdraw, params, &response)
	if err != nil {
		return response, err
	}

	if response.Status != noError {
		return response, errors.New(response.Message)
	}
	return response, nil
}

// MarketBuyOrder initiates a buy order through available order books
//
// currency: BTC, ETH, DASH, LTC, ETC, XRP, BCH, XMR, ZEC, QTUM, BTG, EOS
// (default value: BTC)
// units: Order quantity
func (b *Bithumb) MarketBuyOrder(currency string, units float64) (MarketBuy, error) {
	response := MarketBuy{}

	params := url.Values{}
	params.Set("currency", common.StringToUpper(currency))
	params.Set("units", strconv.FormatFloat(units, 'f', -1, 64))

	err := b.SendAuthenticatedHTTPRequest(privateMarketBuy, params, &response)
	if err != nil {
		return response, err
	}

	if response.Status != noError {
		return response, errors.New(response.Message)
	}
	return response, nil
}

// MarketSellOrder initiates a sell order through available order books
//
// currency: BTC, ETH, DASH, LTC, ETC, XRP, BCH, XMR, ZEC, QTUM, BTG, EOS
// (default value: BTC)
// units: Order quantity
func (b *Bithumb) MarketSellOrder(currency string, units float64) (MarketSell, error) {
	response := MarketSell{}

	params := url.Values{}
	params.Set("currency", common.StringToUpper(currency))
	params.Set("units", strconv.FormatFloat(units, 'f', -1, 64))

	err := b.SendAuthenticatedHTTPRequest(privateMarketSell, params, &response)
	if err != nil {
		return response, err
	}

	if response.Status != noError {
		return response, errors.New(response.Message)
	}
	return response, nil
}

// SendHTTPRequest sends an unauthenticated HTTP request
func (b *Bithumb) SendHTTPRequest(path string, result interface{}) error {
	return b.SendPayload("GET", path, nil, nil, result, false, b.Verbose)
}

// SendAuthenticatedHTTPRequest sends an authenticated HTTP request to bithumb
func (b *Bithumb) SendAuthenticatedHTTPRequest(path string, params url.Values, result interface{}) error {
	if !b.API.AuthenticatedSupport {
		return fmt.Errorf(exchange.WarningAuthenticatedRequestWithoutCredentialsSet, b.Name)
	}

	if params == nil {
		params = url.Values{}
	}

	if b.Nonce.Get() == 0 {
		b.Nonce.Set(time.Now().UnixNano() / int64(time.Millisecond))
	} else {
		b.Nonce.Inc()
	}

	params.Set("endpoint", path)
	payload := params.Encode()
	hmacPayload := path + string(0) + payload + string(0) + b.Nonce.String()
	hmac := common.GetHMAC(common.HashSHA512, []byte(hmacPayload), []byte(b.API.Credentials.Secret))
	hmacStr := common.HexEncodeToString(hmac)

	headers := make(map[string]string)
	headers["Api-Key"] = b.API.Credentials.Key
	headers["Api-Sign"] = common.Base64Encode([]byte(hmacStr))
	headers["Api-Nonce"] = b.Nonce.String()
	headers["Content-Type"] = "application/x-www-form-urlencoded"

	return b.SendPayload("POST", b.API.Endpoints.URL+path, headers, bytes.NewBufferString(payload), result, true, b.Verbose)
}

// GetFee returns an estimate of fee based on type of transaction
func (b *Bithumb) GetFee(feeBuilder exchange.FeeBuilder) (float64, error) {
	var fee float64

	switch feeBuilder.FeeType {
	case exchange.CryptocurrencyTradeFee:
		fee = calculateTradingFee(feeBuilder.PurchasePrice, feeBuilder.Amount)
	case exchange.CyptocurrencyDepositFee:
		fee = getDepositFee(feeBuilder.FirstCurrency, feeBuilder.Amount)
	case exchange.CryptocurrencyWithdrawalFee:
		fee = getWithdrawalFee(feeBuilder.FirstCurrency)
	case exchange.InternationalBankWithdrawalFee:
		fee = getWithdrawalFee(feeBuilder.CurrencyItem)
	}
	if fee < 0 {
		fee = 0
	}
	return fee, nil
}

// calculateTradingFee returns fee when performing a trade
func calculateTradingFee(purchasePrice float64, amount float64) float64 {
	fee := 0.0015

	return fee * amount * purchasePrice
}

// getDepositFee returns fee on a currency when depositing small amounts to bithumb
func getDepositFee(currency string, amount float64) float64 {
	var fee float64

	switch currency {
	case symbol.BTC:
		if amount <= 0.005 {
			fee = 0.001
		}
	case symbol.LTC:
		if amount <= 0.3 {
			fee = 0.01
		}
	case symbol.DASH:
		if amount <= 0.04 {
			fee = 0.01
		}
	case symbol.BCH:
		if amount <= 0.03 {
			fee = 0.001
		}
	case symbol.ZEC:
		if amount <= 0.02 {
			fee = 0.001
		}
	case symbol.BTG:
		if amount <= 0.15 {
			fee = 0.001
		}
	}

	return fee
}

// getWithdrawalFee returns fee on a currency when withdrawing out of bithumb
func getWithdrawalFee(currency string) float64 {
	return WithdrawalFees[currency]
}
