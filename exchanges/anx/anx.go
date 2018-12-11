package anx

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/thrasher-/gocryptotrader/common"
	"github.com/thrasher-/gocryptotrader/currency/symbol"
	"github.com/thrasher-/gocryptotrader/exchanges"
)

const (
	anxAPIURL          = "https://anxpro.com/"
	anxAPIVersion      = "3"
	anxAPIKey          = "apiKey"
	anxCurrencies      = "currencyStatic"
	anxDataToken       = "dataToken"
	anxOrderNew        = "order/new"
	anxCancel          = "order/cancel"
	anxOrderInfo       = "order/info"
	anxSend            = "send"
	anxSubaccountNew   = "subaccount/new"
	anxReceieveAddress = "receive"
	anxCreateAddress   = "receive/create"
	anxTicker          = "money/ticker"
	anxDepth           = "money/depth/full"
	anxAccount         = "account"

	// ANX rate limites for authenticated and unauthenticated requests
	anxAuthRate   = 0
	anxUnauthRate = 0
)

// ANX is the overarching type across the alphapoint package
type ANX struct {
	exchange.Base
}

// GetCurrencies returns a list of supported currencies (both fiat
// and cryptocurrencies)
func (a *ANX) GetCurrencies() (CurrenciesStore, error) {
	var result CurrenciesStaticResponse
	path := fmt.Sprintf("%sapi/3/%s", a.API.Endpoints.URL, anxCurrencies)

	err := a.SendHTTPRequest(path, &result)
	if err != nil {
		return CurrenciesStore{}, err
	}

	return result.CurrenciesResponse, nil
}

// GetTicker returns the current ticker
func (a *ANX) GetTicker(currency string) (Ticker, error) {
	var ticker Ticker
	path := fmt.Sprintf("%sapi/2/%s/%s", a.API.Endpoints.URL, currency, anxTicker)

	return ticker, a.SendHTTPRequest(path, &ticker)
}

// GetDepth returns current orderbook depth.
func (a *ANX) GetDepth(currency string) (Depth, error) {
	var depth Depth
	path := fmt.Sprintf("%sapi/2/%s/%s", a.API.Endpoints.URL, currency, anxDepth)

	return depth, a.SendHTTPRequest(path, &depth)
}

// GetAPIKey returns a new generated API key set.
func (a *ANX) GetAPIKey(username, password, otp, deviceID string) (string, string, error) {
	request := make(map[string]interface{})
	request["nonce"] = strconv.FormatInt(time.Now().UnixNano(), 10)[0:13]
	request["username"] = username
	request["password"] = password

	if otp != "" {
		request["otp"] = otp
	}

	request["deviceId"] = deviceID

	type APIKeyResponse struct {
		APIKey     string `json:"apiKey"`
		APISecret  string `json:"apiSecret"`
		ResultCode string `json:"resultCode"`
		Timestamp  int64  `json:"timestamp"`
	}
	var response APIKeyResponse

	err := a.SendAuthenticatedHTTPRequest(anxAPIKey, request, &response)
	if err != nil {
		return "", "", err
	}

	if response.ResultCode != "OK" {
		return "", "", errors.New("Response code is not OK: " + response.ResultCode)
	}

	return response.APIKey, response.APISecret, nil
}

// GetDataToken returns token data
func (a *ANX) GetDataToken() (string, error) {
	request := make(map[string]interface{})

	type DataTokenResponse struct {
		ResultCode string `json:"resultCode"`
		Timestamp  int64  `json:"timestamp"`
		Token      string `json:"token"`
		UUID       string `json:"uuid"`
	}
	var response DataTokenResponse

	err := a.SendAuthenticatedHTTPRequest(anxDataToken, request, &response)
	if err != nil {
		return "", err
	}

	if response.ResultCode != "OK" {
		return "", errors.New("Response code is not OK: %s" + response.ResultCode)
	}
	return response.Token, nil
}

// NewOrder sends a new order request to the exchange.
func (a *ANX) NewOrder(orderType string, buy bool, tradedCurrency string, tradedCurrencyAmount float64, settlementCurrency string, settlementCurrencyAmount float64, limitPriceSettlement float64,
	replace bool, replaceUUID string, replaceIfActive bool) (string, error) {

	request := make(map[string]interface{})
	var order Order
	order.OrderType = orderType
	order.BuyTradedCurrency = buy

	if buy {
		order.TradedCurrencyAmount = tradedCurrencyAmount
	} else {
		order.SettlementCurrencyAmount = settlementCurrencyAmount
	}

	order.TradedCurrency = tradedCurrency
	order.SettlementCurrency = settlementCurrency
	order.LimitPriceInSettlementCurrency = limitPriceSettlement

	if replace {
		order.ReplaceExistingOrderUUID = replaceUUID
		order.ReplaceOnlyIfActive = replaceIfActive
	}

	request["order"] = order

	type OrderResponse struct {
		OrderID    string `json:"orderId"`
		Timestamp  int64  `json:"timestamp,string"`
		ResultCode string `json:"resultCode"`
	}
	var response OrderResponse

	err := a.SendAuthenticatedHTTPRequest(anxOrderNew, request, &response)
	if err != nil {
		return "", err
	}

	if response.ResultCode != "OK" {
		return "", errors.New("Response code is not OK: " + response.ResultCode)
	}
	return response.OrderID, nil
}

// CancelOrderByIDs cancels orders, requires already knowing order IDs
// There is no existing API call to retrieve orderIds
func (a *ANX) CancelOrderByIDs(orderIds []string) (err error) {
	request := make(map[string]interface{})
	request["orderIds"] = orderIds
	type OrderCancelResponse struct {
		Order      OrderResponse `json:"order"`
		ResultCode string        `json:"resultCode"`
		UUID       int64         `json:"uuid"`
		ErrorCode  int64         `json:"errorCode"`
	}
	var response OrderCancelResponse
	err = a.SendAuthenticatedHTTPRequest(anxCancel, request, &response)
	if response.ResultCode != "OK" {
		log.Printf("Response code is not OK: %s\n", response.ResultCode)
		return errors.New(response.ResultCode)
	}
	return err
}

// OrderInfo returns information about a specific order
func (a *ANX) OrderInfo(orderID string) (OrderResponse, error) {
	request := make(map[string]interface{})
	request["orderId"] = orderID

	type OrderInfoResponse struct {
		Order      OrderResponse `json:"order"`
		ResultCode string        `json:"resultCode"`
		Timestamp  int64         `json:"timestamp"`
	}
	var response OrderInfoResponse

	err := a.SendAuthenticatedHTTPRequest(anxOrderInfo, request, &response)

	if err != nil {
		return OrderResponse{}, err
	}

	if response.ResultCode != "OK" {
		log.Printf("Response code is not OK: %s\n", response.ResultCode)
		return OrderResponse{}, errors.New(response.ResultCode)
	}
	return response.Order, nil
}

// Send withdraws a currency to an address
func (a *ANX) Send(currency, address, otp, amount string) (string, error) {
	request := make(map[string]interface{})
	request["ccy"] = currency
	request["amount"] = amount
	request["address"] = address

	if otp != "" {
		request["otp"] = otp
	}

	type SendResponse struct {
		TransactionID string `json:"transactionId"`
		ResultCode    string `json:"resultCode"`
		Timestamp     int64  `json:"timestamp"`
	}
	var response SendResponse

	err := a.SendAuthenticatedHTTPRequest(anxSend, request, &response)

	if err != nil {
		return "", err
	}

	if response.ResultCode != "OK" {
		log.Printf("Response code is not OK: %s\n", response.ResultCode)
		return "", errors.New(response.ResultCode)
	}
	return response.TransactionID, nil
}

// CreateNewSubAccount generates a new sub account
func (a *ANX) CreateNewSubAccount(currency, name string) (string, error) {
	request := make(map[string]interface{})
	request["ccy"] = currency
	request["customRef"] = name

	type SubaccountResponse struct {
		SubAccount string `json:"subAccount"`
		ResultCode string `json:"resultCode"`
		Timestamp  int64  `json:"timestamp"`
	}
	var response SubaccountResponse

	err := a.SendAuthenticatedHTTPRequest(anxSubaccountNew, request, &response)

	if err != nil {
		return "", err
	}

	if response.ResultCode != "OK" {
		log.Printf("Response code is not OK: %s\n", response.ResultCode)
		return "", errors.New(response.ResultCode)
	}
	return response.SubAccount, nil
}

// GetDepositAddressByCurrency returns a deposit address for a specific currency
func (a *ANX) GetDepositAddressByCurrency(currency, name string, new bool) (string, error) {
	request := make(map[string]interface{})
	request["ccy"] = currency

	if name != "" {
		request["subAccount"] = name
	}

	type AddressResponse struct {
		Address    string `json:"address"`
		SubAccount string `json:"subAccount"`
		ResultCode string `json:"resultCode"`
		Timestamp  int64  `json:"timestamp"`
	}
	var response AddressResponse

	path := anxReceieveAddress
	if new {
		path = anxCreateAddress
	}

	err := a.SendAuthenticatedHTTPRequest(path, request, &response)

	if err != nil {
		return "", err
	}

	if response.ResultCode != "OK" {
		log.Printf("Response code is not OK: %s\n", response.ResultCode)
		return "", errors.New(response.ResultCode)
	}

	return response.Address, nil
}

// SendHTTPRequest sends an unauthenticated HTTP request
func (a *ANX) SendHTTPRequest(path string, result interface{}) error {
	return a.SendPayload("GET", path, nil, nil, result, false, a.Verbose)
}

// SendAuthenticatedHTTPRequest sends a authenticated HTTP request
func (a *ANX) SendAuthenticatedHTTPRequest(path string, params map[string]interface{}, result interface{}) error {
	if !a.API.AuthenticatedSupport {
		return fmt.Errorf(exchange.WarningAuthenticatedRequestWithoutCredentialsSet, a.Name)
	}

	if a.Nonce.Get() == 0 {
		a.Nonce.Set(time.Now().UnixNano())
	} else {
		a.Nonce.Inc()
	}

	request := make(map[string]interface{})
	request["nonce"] = a.Nonce.String()[0:13]
	path = fmt.Sprintf("api/%s/%s", anxAPIVersion, path)

	if params != nil {
		for key, value := range params {
			request[key] = value
		}
	}

	PayloadJSON, err := common.JSONEncode(request)

	if err != nil {
		return errors.New("SendAuthenticatedHTTPRequest: Unable to JSON request")
	}

	if a.Verbose {
		log.Printf("Request JSON: %s\n", PayloadJSON)
	}

	hmac := common.GetHMAC(common.HashSHA512, []byte(path+string("\x00")+string(PayloadJSON)), []byte(a.API.Credentials.Secret))
	headers := make(map[string]string)
	headers["Rest-Key"] = a.API.Credentials.Key
	headers["Rest-Sign"] = common.Base64Encode([]byte(hmac))
	headers["Content-Type"] = "application/json"

	return a.SendPayload("POST", a.API.Endpoints.URL+path, headers, bytes.NewBuffer(PayloadJSON), result, true, a.Verbose)
}

// GetFee returns an estimate of fee based on type of transaction
func (a *ANX) GetFee(feeBuilder exchange.FeeBuilder) (float64, error) {
	var fee float64

	switch feeBuilder.FeeType {
	case exchange.CryptocurrencyTradeFee:
		fee = a.calculateTradingFee(feeBuilder.PurchasePrice, feeBuilder.Amount, feeBuilder.IsMaker)
	case exchange.CryptocurrencyWithdrawalFee:
		fee = getCryptocurrencyWithdrawalFee(feeBuilder.FirstCurrency)
	case exchange.InternationalBankWithdrawalFee:
		fee = getInternationalBankWithdrawalFee(feeBuilder.CurrencyItem, feeBuilder.Amount)
	}
	if fee < 0 {
		fee = 0
	}
	return fee, nil
}

func (a *ANX) calculateTradingFee(purchasePrice, amount float64, isMaker bool) float64 {
	var fee float64

	if isMaker {
		fee = 0.01 * amount * purchasePrice
	} else {
		fee = 0.02 * amount * purchasePrice
	}

	return fee
}

func getCryptocurrencyWithdrawalFee(currency string) float64 {
	return WithdrawalFees[currency]
}

func getInternationalBankWithdrawalFee(currency string, amount float64) float64 {
	var fee float64

	if currency == symbol.HKD {
		fee = 250 + (WithdrawalFees[currency] * amount)
	}
	//TODO, other fiat currencies require consultation with ANXPRO
	return fee
}

// GetAccountInformation retrieves details including API permissions
func (a *ANX) GetAccountInformation() (AccountInformation, error) {
	var response AccountInformation
	err := a.SendAuthenticatedHTTPRequest(anxOrderInfo, nil, &response)
	if err != nil {
		return response, err
	}

	if response.ResultCode != "OK" {
		log.Printf("Response code is not OK: %s\n", response.ResultCode)
		return response, errors.New(response.ResultCode)
	}
	return response, nil
}

// CheckAPIWithdrawPermission checks if the API key is allowed to withdraw
func (a *ANX) CheckAPIWithdrawPermission() (bool, error) {
	accountInfo, err := a.GetAccountInformation()

	if err != nil {
		return false, err
	}

	var apiAllowsWithdraw bool

	for _, a := range accountInfo.Rights {
		if a == "withdraw" {
			apiAllowsWithdraw = true
		}
	}

	if !apiAllowsWithdraw {
		log.Printf("API key is missing withdrawal permissions")
	}

	return apiAllowsWithdraw, nil
}
