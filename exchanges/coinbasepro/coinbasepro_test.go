package coinbasepro

import (
	"testing"
	"time"

	"github.com/thrasher-/gocryptotrader/config"
	"github.com/thrasher-/gocryptotrader/currency/pair"
	"github.com/thrasher-/gocryptotrader/currency/symbol"
	exchange "github.com/thrasher-/gocryptotrader/exchanges"
)

var c CoinbasePro

// Please supply your APIKeys here for better testing
const (
	apiKey                  = ""
	apiSecret               = ""
	clientID                = "" //passphrase you made at API CREATION
	canManipulateRealOrders = false
)

func TestSetDefaults(t *testing.T) {
	c.SetDefaults()
	c.Requester.SetRateLimit(false, time.Second, 1)
}

func TestSetup(t *testing.T) {
	cfg := config.GetConfig()
	cfg.LoadConfig("../../testdata/configtest.json")
	gdxConfig, err := cfg.GetExchangeConfig("CoinbasePro")
	if err != nil {
		t.Error("Test Failed - coinbasepro Setup() init error")
	}
	gdxConfig.API.Credentials.Key = apiKey
	gdxConfig.API.Credentials.Secret = apiSecret
	gdxConfig.API.AuthenticatedSupport = true
	c.Setup(gdxConfig)
}

func TestGetProducts(t *testing.T) {
	_, err := c.GetProducts()
	if err != nil {
		t.Errorf("Test failed - Coinbase, GetProducts() Error: %s", err)
	}
}

func TestGetTicker(t *testing.T) {
	_, err := c.GetTicker("BTC-USD")
	if err != nil {
		t.Error("Test failed - GetTicker() error", err)
	}
}

func TestGetTrades(t *testing.T) {
	_, err := c.GetTrades("BTC-USD")
	if err != nil {
		t.Error("Test failed - GetTrades() error", err)
	}
}

func TestGetHistoricRates(t *testing.T) {
	_, err := c.GetHistoricRates("BTC-USD", 0, 0, 0)
	if err != nil {
		t.Error("Test failed - GetHistoricRates() error", err)
	}
}

func TestGetStats(t *testing.T) {
	_, err := c.GetStats("BTC-USD")
	if err != nil {
		t.Error("Test failed - GetStats() error", err)
	}
}

func TestGetCurrencies(t *testing.T) {
	_, err := c.GetCurrencies()
	if err != nil {
		t.Error("Test failed - GetCurrencies() error", err)
	}
}

func TestGetServerTime(t *testing.T) {
	_, err := c.GetServerTime()
	if err != nil {
		t.Error("Test failed - GetServerTime() error", err)
	}
}

func TestAuthRequests(t *testing.T) {

	if c.ValidateAPICredentials() {

		_, err := c.GetAccounts()
		if err == nil {
			t.Error("Test failed - GetAccounts() error", err)
		}

		_, err = c.GetAccount("234cb213-ac6f-4ed8-b7b6-e62512930945")
		if err == nil {
			t.Error("Test failed - GetAccount() error", err)
		}

		_, err = c.GetAccountHistory("234cb213-ac6f-4ed8-b7b6-e62512930945")
		if err == nil {
			t.Error("Test failed - GetAccountHistory() error", err)
		}

		_, err = c.GetHolds("234cb213-ac6f-4ed8-b7b6-e62512930945")
		if err == nil {
			t.Error("Test failed - GetHolds() error", err)
		}

		_, err = c.PlaceLimitOrder("", 0, 0, "buy", "", "", "BTC-USD", "", false)
		if err == nil {
			t.Error("Test failed - PlaceLimitOrder() error", err)
		}

		_, err = c.PlaceMarketOrder("", 1, 0, "buy", "BTC-USD", "")
		if err == nil {
			t.Error("Test failed - PlaceMarketOrder() error", err)
		}

		err = c.CancelExistingOrder("1337")
		if err == nil {
			t.Error("Test failed - CancelExistingOrder() error", err)
		}

		_, err = c.CancelAllExistingOrders("BTC-USD")
		if err == nil {
			t.Error("Test failed - CancelAllExistingOrders() error", err)
		}

		_, err = c.GetOrders([]string{"open", "done"}, "BTC-USD")
		if err == nil {
			t.Error("Test failed - GetOrders() error", err)
		}

		_, err = c.GetOrder("1337")
		if err == nil {
			t.Error("Test failed - GetOrders() error", err)
		}

		_, err = c.GetFills("1337", "BTC-USD")
		if err == nil {
			t.Error("Test failed - GetFills() error", err)
		}
		_, err = c.GetFills("", "")
		if err == nil {
			t.Error("Test failed - GetFills() error", err)
		}

		_, err = c.GetFundingRecords("rejected")
		if err == nil {
			t.Error("Test failed - GetFundingRecords() error", err)
		}

		// 	_, err := c.RepayFunding("1", "BTC")
		// 	if err != nil {
		// 		t.Error("Test failed - RepayFunding() error", err)
		// 	}

		_, err = c.MarginTransfer(1, "withdraw", "45fa9e3b-00ba-4631-b907-8a98cbdf21be", "BTC")
		if err == nil {
			t.Error("Test failed - MarginTransfer() error", err)
		}

		_, err = c.GetPosition()
		if err == nil {
			t.Error("Test failed - GetPosition() error", err)
		}

		_, err = c.ClosePosition(false)
		if err == nil {
			t.Error("Test failed - ClosePosition() error", err)
		}

		_, err = c.GetPayMethods()
		if err == nil {
			t.Error("Test failed - GetPayMethods() error", err)
		}

		_, err = c.DepositViaPaymentMethod(1, "BTC", "1337")
		if err == nil {
			t.Error("Test failed - DepositViaPaymentMethod() error", err)
		}

		_, err = c.DepositViaCoinbase(1, "BTC", "1337")
		if err == nil {
			t.Error("Test failed - DepositViaCoinbase() error", err)
		}

		_, err = c.WithdrawViaPaymentMethod(1, "BTC", "1337")
		if err == nil {
			t.Error("Test failed - WithdrawViaPaymentMethod() error", err)
		}

		// 	_, err := c.WithdrawViaCoinbase(1, "BTC", "c13cd0fc-72ca-55e9-843b-b84ef628c198")
		// 	if err != nil {
		// 		t.Error("Test failed - WithdrawViaCoinbase() error", err)
		// 	}

		_, err = c.WithdrawCrypto(1, "BTC", "1337")
		if err == nil {
			t.Error("Test failed - WithdrawViaCoinbase() error", err)
		}

		_, err = c.GetCoinbaseAccounts()
		if err == nil {
			t.Error("Test failed - GetCoinbaseAccounts() error", err)
		}

		_, err = c.GetReportStatus("1337")
		if err == nil {
			t.Error("Test failed - GetReportStatus() error", err)
		}

		_, err = c.GetTrailingVolume()
		if err == nil {
			t.Error("Test failed - GetTrailingVolume() error", err)
		}
	}
}

func setFeeBuilder() exchange.FeeBuilder {
	return exchange.FeeBuilder{
		Amount:         1,
		Delimiter:      "",
		FeeType:        exchange.CryptocurrencyTradeFee,
		FirstCurrency:  symbol.BTC,
		SecondCurrency: symbol.LTC,
		IsMaker:        false,
		PurchasePrice:  1,
	}
}

func TestGetFee(t *testing.T) {
	c.SetDefaults()
	TestSetup(t)

	var feeBuilder = setFeeBuilder()

	if apiKey != "" || apiSecret != "" {
		// CryptocurrencyTradeFee Basic
		if resp, err := c.GetFee(feeBuilder); resp != float64(0.003) || err != nil {
			t.Error(err)
			t.Errorf("Test Failed - GetFee() error. Expected: %f, Recieved: %f", float64(0.003), resp)
		}

		// CryptocurrencyTradeFee High quantity
		feeBuilder = setFeeBuilder()
		feeBuilder.Amount = 1000
		feeBuilder.PurchasePrice = 1000
		if resp, err := c.GetFee(feeBuilder); resp != float64(3000) || err != nil {
			t.Errorf("Test Failed - GetFee() error. Expected: %f, Recieved: %f", float64(3000), resp)
			t.Error(err)
		}

		// CryptocurrencyTradeFee IsMaker
		feeBuilder = setFeeBuilder()
		feeBuilder.IsMaker = true
		if resp, err := c.GetFee(feeBuilder); resp != float64(0) || err != nil {
			t.Errorf("Test Failed - GetFee() error. Expected: %f, Recieved: %f", float64(0.01), resp)
			t.Error(err)
		}

		// CryptocurrencyTradeFee Negative purchase price
		feeBuilder = setFeeBuilder()
		feeBuilder.PurchasePrice = -1000
		if resp, err := c.GetFee(feeBuilder); resp != float64(0) || err != nil {
			t.Errorf("Test Failed - GetFee() error. Expected: %f, Recieved: %f", float64(0), resp)
			t.Error(err)
		}
	}

	// CryptocurrencyWithdrawalFee Basic
	feeBuilder = setFeeBuilder()
	feeBuilder.FeeType = exchange.CryptocurrencyWithdrawalFee
	if resp, err := c.GetFee(feeBuilder); resp != float64(0) || err != nil {
		t.Errorf("Test Failed - GetFee() error. Expected: %f, Recieved: %f", float64(0), resp)
		t.Error(err)
	}

	// CyptocurrencyDepositFee Basic
	feeBuilder = setFeeBuilder()
	feeBuilder.FeeType = exchange.CyptocurrencyDepositFee
	if resp, err := c.GetFee(feeBuilder); resp != float64(0) || err != nil {
		t.Errorf("Test Failed - GetFee() error. Expected: %f, Recieved: %f", float64(0), resp)
		t.Error(err)
	}

	// InternationalBankDepositFee Basic
	feeBuilder = setFeeBuilder()
	feeBuilder.FeeType = exchange.InternationalBankDepositFee
	feeBuilder.CurrencyItem = symbol.EUR
	if resp, err := c.GetFee(feeBuilder); resp != float64(0.15) || err != nil {
		t.Errorf("Test Failed - GetFee() error. Expected: %f, Recieved: %f", float64(0), resp)
		t.Error(err)
	}

	// InternationalBankWithdrawalFee Basic
	feeBuilder = setFeeBuilder()
	feeBuilder.FeeType = exchange.InternationalBankWithdrawalFee
	feeBuilder.CurrencyItem = symbol.USD
	if resp, err := c.GetFee(feeBuilder); resp != float64(25) || err != nil {
		t.Errorf("Test Failed - GetFee() error. Expected: %f, Recieved: %f", float64(0), resp)
		t.Error(err)
	}
}

func TestCalculateTradingFee(t *testing.T) {
	t.Parallel()
	// uppercase
	var volume = []Volume{
		Volume{
			ProductID: "BTC_USD",
			Volume:    100,
		},
	}

	if resp := c.calculateTradingFee(volume, "btc", "_", "usd", 1, 1, false); resp != float64(0.003) {
		t.Errorf("Test Failed - GetFee() error. Expected: %f, Recieved: %f", float64(0.003), resp)
	}

	// lowercase
	volume = []Volume{
		Volume{
			ProductID: "btc_usd",
			Volume:    100,
		},
	}

	if resp := c.calculateTradingFee(volume, "btc", "_", "usd", 1, 1, false); resp != float64(0.003) {
		t.Errorf("Test Failed - GetFee() error. Expected: %f, Recieved: %f", float64(0.003), resp)
	}

	// mixedCase
	volume = []Volume{
		Volume{
			ProductID: "btc_USD",
			Volume:    100,
		},
	}

	if resp := c.calculateTradingFee(volume, "btc", "_", "usd", 1, 1, false); resp != float64(0.003) {
		t.Errorf("Test Failed - GetFee() error. Expected: %f, Recieved: %f", float64(0.003), resp)
	}

	// medium volume
	volume = []Volume{
		Volume{
			ProductID: "btc_USD",
			Volume:    10000001,
		},
	}

	if resp := c.calculateTradingFee(volume, "btc", "_", "usd", 1, 1, false); resp != float64(0.002) {
		t.Errorf("Test Failed - GetFee() error. Expected: %f, Recieved: %f", float64(0.002), resp)
	}

	// high volume
	volume = []Volume{
		Volume{
			ProductID: "btc_USD",
			Volume:    100000010000,
		},
	}

	if resp := c.calculateTradingFee(volume, "btc", "_", "usd", 1, 1, false); resp != float64(0.001) {
		t.Errorf("Test Failed - GetFee() error. Expected: %f, Recieved: %f", float64(0.001), resp)
	}

	// no match
	volume = []Volume{
		Volume{
			ProductID: "btc_beeteesee",
			Volume:    100000010000,
		},
	}

	if resp := c.calculateTradingFee(volume, "btc", "_", "usd", 1, 1, false); resp != float64(0) {
		t.Errorf("Test Failed - GetFee() error. Expected: %f, Recieved: %f", float64(0), resp)
	}

	// taker
	volume = []Volume{
		Volume{
			ProductID: "btc_USD",
			Volume:    100000010000,
		},
	}

	if resp := c.calculateTradingFee(volume, "btc", "_", "usd", 1, 1, true); resp != float64(0) {
		t.Errorf("Test Failed - GetFee() error. Expected: %f, Recieved: %f", float64(0), resp)
	}
}

func TestFormatWithdrawPermissions(t *testing.T) {
	// Arrange
	c.SetDefaults()
	expectedResult := exchange.AutoWithdrawCryptoWithAPIPermissionText + " & " + exchange.AutoWithdrawFiatWithAPIPermissionText
	// Act
	withdrawPermissions := c.FormatWithdrawPermissions()
	// Assert
	if withdrawPermissions != expectedResult {
		t.Errorf("Expected: %s, Recieved: %s", expectedResult, withdrawPermissions)
	}
}

// Any tests below this line have the ability to impact your orders on the exchange. Enable canManipulateRealOrders to run them
// ----------------------------------------------------------------------------------------------------------------------------
func isRealOrderTestEnabled() bool {
	if !c.ValidateAPICredentials() || !canManipulateRealOrders {
		return false
	}
	return true
}

func TestSubmitOrder(t *testing.T) {
	c.SetDefaults()
	TestSetup(t)

	if !isRealOrderTestEnabled() {
		t.Skip()
	}

	var p = pair.CurrencyPair{
		Delimiter:      "-",
		FirstCurrency:  symbol.BTC,
		SecondCurrency: symbol.LTC,
	}

	response, err := c.SubmitOrder(p, exchange.Buy, exchange.Limit, 1, 1, "clientId")
	if err != nil || !response.IsOrderPlaced {
		t.Errorf("Order failed to be placed: %v", err)
	}
}

func TestCancelExchangeOrder(t *testing.T) {
	// Arrange
	c.SetDefaults()
	TestSetup(t)

	if !isRealOrderTestEnabled() {
		t.Skip()
	}

	c.Verbose = true
	currencyPair := pair.NewCurrencyPair(symbol.LTC, symbol.BTC)

	var orderCancellation = exchange.OrderCancellation{
		OrderID:       "1",
		WalletAddress: "1F5zVDgNjorJ51oGebSvNCrSAHpwGkUdDB",
		AccountID:     "1",
		CurrencyPair:  currencyPair,
	}

	// Act
	err := c.CancelOrder(orderCancellation)

	// Assert
	if err != nil {
		t.Errorf("Could not cancel order: %s", err)
	}
}
