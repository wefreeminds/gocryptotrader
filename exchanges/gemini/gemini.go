package gemini

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"

	"github.com/thrasher-/gocryptotrader/common"
	"github.com/thrasher-/gocryptotrader/exchanges"
)

const (
	geminiAPIURL        = "https://api.gemini.com"
	geminiSandboxAPIURL = "https://api.sandbox.gemini.com"
	geminiAPIVersion    = "1"

	geminiSymbols            = "symbols"
	geminiTicker             = "pubticker"
	geminiAuction            = "auction"
	geminiAuctionHistory     = "history"
	geminiOrderbook          = "book"
	geminiTrades             = "trades"
	geminiOrders             = "orders"
	geminiOrderNew           = "order/new"
	geminiOrderCancel        = "order/cancel"
	geminiOrderCancelSession = "order/cancel/session"
	geminiOrderCancelAll     = "order/cancel/all"
	geminiOrderStatus        = "order/status"
	geminiMyTrades           = "mytrades"
	geminiBalances           = "balances"
	geminiTradeVolume        = "tradevolume"
	geminiDeposit            = "deposit"
	geminiNewAddress         = "newAddress"
	geminiWithdraw           = "withdraw/"
	geminiHeartbeat          = "heartbeat"
	geminiVolume             = "notionalvolume"

	// gemini limit rates
	geminiAuthRate   = 600
	geminiUnauthRate = 120

	// Too many requests returns this
	geminiRateError = "429"

	// Assigned API key roles on creation
	geminiRoleTrader      = "trader"
	geminiRoleFundManager = "fundmanager"
)

var (
	// Session manager
	Session map[int]*Gemini
)

// Gemini is the overarching type across the Gemini package, create multiple
// instances with differing APIkeys for segregation of roles for authenticated
// requests & sessions by appending new sessions to the Session map using
// AddSession, if sandbox test is needed append a new session with with the same
// API keys and change the IsSandbox variable to true.
type Gemini struct {
	exchange.Base
	Role              string
	RequiresHeartBeat bool
}

// AddSession adds a new session to the gemini base
func AddSession(g *Gemini, sessionID int, apiKey, apiSecret, role string, needsHeartbeat, isSandbox bool) error {
	if Session == nil {
		Session = make(map[int]*Gemini)
	}

	_, ok := Session[sessionID]
	if ok {
		return errors.New("sessionID already being used")
	}

	g.API.Credentials.Key = apiKey
	g.API.Credentials.Secret = apiSecret
	g.Role = role
	g.RequiresHeartBeat = needsHeartbeat
	g.API.Endpoints.URL = geminiAPIURL

	if isSandbox {
		g.API.Endpoints.URL = geminiSandboxAPIURL
	}

	Session[sessionID] = g

	return nil
}

// GetSymbols returns all available symbols for trading
func (g *Gemini) GetSymbols() ([]string, error) {
	symbols := []string{}
	path := fmt.Sprintf("%s/v%s/%s", g.API.Endpoints.URL, geminiAPIVersion, geminiSymbols)

	return symbols, g.SendHTTPRequest(path, &symbols)
}

// GetTicker returns information about recent trading activity for the symbol
func (g *Gemini) GetTicker(currencyPair string) (Ticker, error) {

	type TickerResponse struct {
		Ask     float64 `json:"ask,string"`
		Bid     float64 `json:"bid,string"`
		Last    float64 `json:"last,string"`
		Volume  map[string]interface{}
		Message string `json:"message"`
	}

	ticker := Ticker{}
	resp := TickerResponse{}
	path := fmt.Sprintf("%s/v%s/%s/%s", g.API.Endpoints.URL, geminiAPIVersion, geminiTicker, currencyPair)

	err := g.SendHTTPRequest(path, &resp)
	if err != nil {
		return ticker, err
	}

	if resp.Message != "" {
		return ticker, errors.New(resp.Message)
	}

	ticker.Ask = resp.Ask
	ticker.Bid = resp.Bid
	ticker.Last = resp.Last

	ticker.Volume.Currency, _ = strconv.ParseFloat(resp.Volume[currencyPair[0:3]].(string), 64)

	if common.StringContains(currencyPair, "USD") {
		ticker.Volume.USD, _ = strconv.ParseFloat(resp.Volume["USD"].(string), 64)
	} else {
		if resp.Volume["ETH"] != nil {
			ticker.Volume.ETH, _ = strconv.ParseFloat(resp.Volume["ETH"].(string), 64)
		}

		if resp.Volume["BTC"] != nil {
			ticker.Volume.BTC, _ = strconv.ParseFloat(resp.Volume["BTC"].(string), 64)
		}
	}

	time, _ := resp.Volume["timestamp"].(float64)
	ticker.Volume.Timestamp = int64(time)

	return ticker, nil
}

// GetOrderbook returns the current order book, as two arrays, one of bids, and
// one of asks
//
// params - limit_bids or limit_asks [OPTIONAL] default 50, 0 returns all Values
// Type is an integer ie "params.Set("limit_asks", 30)"
func (g *Gemini) GetOrderbook(currencyPair string, params url.Values) (Orderbook, error) {
	path := common.EncodeURLValues(fmt.Sprintf("%s/v%s/%s/%s", g.API.Endpoints.URL, geminiAPIVersion, geminiOrderbook, currencyPair), params)
	orderbook := Orderbook{}

	return orderbook, g.SendHTTPRequest(path, &orderbook)
}

// GetTrades eturn the trades that have executed since the specified timestamp.
// Timestamps are either seconds or milliseconds since the epoch (1970-01-01).
//
// currencyPair - example "btcusd"
// params --
// since, timestamp [optional]
// limit_trades	integer	Optional. The maximum number of trades to return.
// include_breaks	boolean	Optional. Whether to display broken trades. False by
// default. Can be '1' or 'true' to activate
func (g *Gemini) GetTrades(currencyPair string, params url.Values) ([]Trade, error) {
	path := common.EncodeURLValues(fmt.Sprintf("%s/v%s/%s/%s", g.API.Endpoints.URL, geminiAPIVersion, geminiTrades, currencyPair), params)
	trades := []Trade{}

	return trades, g.SendHTTPRequest(path, &trades)
}

// GetAuction returns auction information
func (g *Gemini) GetAuction(currencyPair string) (Auction, error) {
	path := fmt.Sprintf("%s/v%s/%s/%s", g.API.Endpoints.URL, geminiAPIVersion, geminiAuction, currencyPair)
	auction := Auction{}

	return auction, g.SendHTTPRequest(path, &auction)
}

// GetAuctionHistory returns the auction events, optionally including
// publications of indicative prices, since the specific timestamp.
//
// currencyPair - example "btcusd"
// params -- [optional]
//          since - [timestamp] Only returns auction events after the specified
// timestamp.
//          limit_auction_results - [integer] The maximum number of auction
// events to return.
//          include_indicative - [bool] Whether to include publication of
// indicative prices and quantities.
func (g *Gemini) GetAuctionHistory(currencyPair string, params url.Values) ([]AuctionHistory, error) {
	path := common.EncodeURLValues(fmt.Sprintf("%s/v%s/%s/%s/%s", g.API.Endpoints.URL, geminiAPIVersion, geminiAuction, currencyPair, geminiAuctionHistory), params)
	auctionHist := []AuctionHistory{}

	return auctionHist, g.SendHTTPRequest(path, &auctionHist)
}

func (g *Gemini) isCorrectSession(role string) error {
	if g.Role != role {
		return errors.New("incorrect role for APIKEY cannot use this function")
	}
	return nil
}

// NewOrder Only limit orders are supported through the API at present.
// returns order ID if successful
func (g *Gemini) NewOrder(symbol string, amount, price float64, side, orderType string) (int64, error) {
	if err := g.isCorrectSession(geminiRoleTrader); err != nil {
		return 0, err
	}

	request := make(map[string]interface{})
	request["symbol"] = symbol
	request["amount"] = strconv.FormatFloat(amount, 'f', -1, 64)
	request["price"] = strconv.FormatFloat(price, 'f', -1, 64)
	request["side"] = side
	request["type"] = orderType

	response := Order{}
	err := g.SendAuthenticatedHTTPRequest("POST", geminiOrderNew, request, &response)
	if err != nil {
		return 0, err
	}
	return response.OrderID, nil
}

// CancelExistingOrder will cancel an order. If the order is already canceled, the
// message will succeed but have no effect.
func (g *Gemini) CancelExistingOrder(OrderID int64) (Order, error) {
	request := make(map[string]interface{})
	request["order_id"] = OrderID

	response := Order{}
	err := g.SendAuthenticatedHTTPRequest("POST", geminiOrderCancel, request, &response)
	if err != nil {
		return Order{}, err
	}
	if response.Message != "" {
		return response, errors.New(response.Message)
	}

	return response, nil
}

// CancelExistingOrders will cancel all outstanding orders created by all
// sessions owned by this account, including interactive orders placed through
// the UI. If sessions = true will only cancel the order that is called on this
// session asssociated with the APIKEY
func (g *Gemini) CancelExistingOrders(CancelBySession bool) (OrderResult, error) {
	response := OrderResult{}
	path := geminiOrderCancelAll
	if CancelBySession {
		path = geminiOrderCancelSession
	}

	err := g.SendAuthenticatedHTTPRequest("POST", path, nil, &response)
	if err != nil {
		return response, err
	}
	if response.Message != "" {
		return response, errors.New(response.Message)
	}
	return response, nil
}

// GetOrderStatus returns the status for an order
func (g *Gemini) GetOrderStatus(orderID int64) (Order, error) {
	request := make(map[string]interface{})
	request["order_id"] = orderID

	response := Order{}

	err := g.SendAuthenticatedHTTPRequest("POST", geminiOrderStatus, request, &response)
	if err != nil {
		return response, err
	}

	if response.Message != "" {
		return response, errors.New(response.Message)
	}
	return response, nil
}

// GetOrders returns active orders in the market
func (g *Gemini) GetOrders() ([]Order, error) {
	var response struct {
		orders  []Order
		Message string `json:"message"`
	}

	err := g.SendAuthenticatedHTTPRequest("POST", geminiOrders, nil, &response)
	if err != nil {
		return response.orders, err
	}
	if response.Message != "" {
		return response.orders, errors.New(response.Message)
	}
	return response.orders, nil
}

// GetTradeHistory returns an array of trades that have been on the exchange
//
// currencyPair - example "btcusd"
// timestamp - [optional] Only return trades on or after this timestamp.
func (g *Gemini) GetTradeHistory(currencyPair string, timestamp int64) ([]TradeHistory, error) {
	response := []TradeHistory{}
	request := make(map[string]interface{})
	request["symbol"] = currencyPair

	if timestamp != 0 {
		request["timestamp"] = timestamp
	}

	return response,
		g.SendAuthenticatedHTTPRequest("POST", geminiMyTrades, request, &response)
}

// GetNotionalVolume returns  the volume in price currency that has been traded across all pairs over a period of 30 days
func (g *Gemini) GetNotionalVolume() (NotionalVolume, error) {
	response := NotionalVolume{}

	return response,
		g.SendAuthenticatedHTTPRequest("POST", geminiVolume, nil, &response)
}

// GetTradeVolume returns a multi-arrayed volume response
func (g *Gemini) GetTradeVolume() ([][]TradeVolume, error) {
	response := [][]TradeVolume{}

	return response,
		g.SendAuthenticatedHTTPRequest("POST", geminiTradeVolume, nil, &response)
}

// GetBalances returns available balances in the supported currencies
func (g *Gemini) GetBalances() ([]Balance, error) {
	response := []Balance{}

	return response,
		g.SendAuthenticatedHTTPRequest("POST", geminiBalances, nil, &response)
}

// GetCryptoDepositAddress returns a deposit address
func (g *Gemini) GetCryptoDepositAddress(depositAddlabel, currency string) (DepositAddress, error) {
	response := DepositAddress{}

	err := g.SendAuthenticatedHTTPRequest("POST", geminiDeposit+"/"+currency+"/"+geminiNewAddress, nil, &response)
	if err != nil {
		return response, err
	}
	if response.Message != "" {
		return response, errors.New(response.Message)
	}
	return response, nil
}

// WithdrawCrypto withdraws crypto currency to a whitelisted address
func (g *Gemini) WithdrawCrypto(address, currency string, amount float64) (WithdrawalAddress, error) {
	response := WithdrawalAddress{}
	request := make(map[string]interface{})
	request["address"] = address
	request["amount"] = strconv.FormatFloat(amount, 'f', -1, 64)

	err := g.SendAuthenticatedHTTPRequest("POST", geminiWithdraw+currency, nil, &response)
	if err != nil {
		return response, err
	}
	if response.Message != "" {
		return response, errors.New(response.Message)
	}
	return response, nil
}

// PostHeartbeat sends a maintenance heartbeat to the exchange for all heartbeat
// maintaned sessions
func (g *Gemini) PostHeartbeat() (string, error) {
	type Response struct {
		Result  string `json:"result"`
		Message string `json:"message"`
	}
	response := Response{}

	err := g.SendAuthenticatedHTTPRequest("POST", geminiHeartbeat, nil, &response)
	if err != nil {
		return response.Result, err
	}
	if response.Message != "" {
		return response.Result, errors.New(response.Message)
	}
	return response.Result, nil
}

// SendHTTPRequest sends an unauthenticated request
func (g *Gemini) SendHTTPRequest(path string, result interface{}) error {
	return g.SendPayload("GET", path, nil, nil, result, false, g.Verbose)
}

// SendAuthenticatedHTTPRequest sends an authenticated HTTP request to the
// exchange and returns an error
func (g *Gemini) SendAuthenticatedHTTPRequest(method, path string, params map[string]interface{}, result interface{}) (err error) {
	if !g.API.AuthenticatedSupport {
		return fmt.Errorf(exchange.WarningAuthenticatedRequestWithoutCredentialsSet, g.Name)
	}

	headers := make(map[string]string)
	request := make(map[string]interface{})
	request["request"] = fmt.Sprintf("/v%s/%s", geminiAPIVersion, path)
	request["nonce"] = g.Nonce.GetValue(g.Name, false)

	if params != nil {
		for key, value := range params {
			request[key] = value
		}
	}

	PayloadJSON, err := common.JSONEncode(request)
	if err != nil {
		return errors.New("SendAuthenticatedHTTPRequest: Unable to JSON request")
	}

	if g.Verbose {
		log.Printf("Request JSON: %s\n", PayloadJSON)
	}

	PayloadBase64 := common.Base64Encode(PayloadJSON)
	hmac := common.GetHMAC(common.HashSHA512_384, []byte(PayloadBase64), []byte(g.API.Credentials.Secret))

	headers["X-GEMINI-APIKEY"] = g.API.Credentials.Key
	headers["X-GEMINI-PAYLOAD"] = PayloadBase64
	headers["X-GEMINI-SIGNATURE"] = common.HexEncodeToString(hmac)

	return g.SendPayload(method, g.API.Endpoints.URL+"/v1/"+path, headers, strings.NewReader(""), result, true, g.Verbose)
}

// GetFee returns an estimate of fee based on type of transaction
func (g *Gemini) GetFee(feeBuilder exchange.FeeBuilder) (float64, error) {
	var fee float64
	switch feeBuilder.FeeType {
	case exchange.CryptocurrencyTradeFee:
		notionVolume, err := g.GetNotionalVolume()
		if err != nil {
			return 0, err
		}
		fee = calculateTradingFee(notionVolume, feeBuilder.PurchasePrice, feeBuilder.Amount, feeBuilder.IsMaker)
	case exchange.CryptocurrencyWithdrawalFee:
		// TODO: no free transactions after 10; Need database to know how many trades have been done
		// Could do via trade history, but would require analysis of response and dates to determine level of fee
	case exchange.InternationalBankWithdrawalFee:
		fee = 0
	}
	if fee < 0 {
		fee = 0
	}

	return fee, nil
}

func calculateTradingFee(notionVolume NotionalVolume, purchasePrice, amount float64, isMaker bool) float64 {
	var volumeFee float64
	if isMaker {
		volumeFee = (float64(notionVolume.MakerFee) / 100)
	} else {
		volumeFee = (float64(notionVolume.TakerFee) / 100)
	}

	return volumeFee * amount * purchasePrice
}
