package bitfinex

import (
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/thrasher-/gocryptotrader/common"
	"github.com/thrasher-/gocryptotrader/config"
	"github.com/thrasher-/gocryptotrader/currency/pair"
	"github.com/thrasher-/gocryptotrader/currency/symbol"
	exchange "github.com/thrasher-/gocryptotrader/exchanges"
)

// Please supply your own keys here to do better tests
const (
	testAPIKey              = ""
	testAPISecret           = ""
	canManipulateRealOrders = false
)

var b Bitfinex

func TestSetup(t *testing.T) {
	b.SetDefaults()
	cfg := config.GetConfig()
	cfg.LoadConfig("../../testdata/configtest.json")
	bfxConfig, err := cfg.GetExchangeConfig("Bitfinex")
	if err != nil {
		t.Error("Test Failed - Bitfinex Setup() init error")
	}
	b.Setup(bfxConfig)
	b.API.Credentials.Key = testAPIKey
	b.API.Credentials.Secret = testAPISecret
	if !b.Enabled || b.API.AuthenticatedSupport ||
		b.Verbose || b.Websocket.IsEnabled() || len(b.BaseCurrencies) < 1 ||
		len(b.CurrencyPairs.Spot.Available) < 1 || len(b.CurrencyPairs.Spot.Enabled) < 1 {
		t.Error("Test Failed - Bitfinex Setup values not set correctly")
	}
	b.API.AuthenticatedSupport = true
	// custom rate limit for testing
	b.Requester.SetRateLimit(true, time.Millisecond*300, 1)
	b.Requester.SetRateLimit(false, time.Millisecond*300, 1)
}

func TestGetPlatformStatus(t *testing.T) {
	t.Parallel()

	result, err := b.GetPlatformStatus()
	if err != nil {
		t.Errorf("TestGetPlatformStatus error: %s", err)
	}

	if result != bitfinexOperativeMode && result != bitfinexMaintenanceMode {
		t.Errorf("TestGetPlatformStatus unexpected response code")
	}
}

func TestGetLatestSpotPrice(t *testing.T) {
	t.Parallel()
	_, err := b.GetLatestSpotPrice("BTCUSD")
	if err != nil {
		t.Error("Bitfinex GetLatestSpotPrice error: ", err)
	}
}

func TestGetTicker(t *testing.T) {
	t.Parallel()
	_, err := b.GetTicker("BTCUSD")
	if err != nil {
		t.Error("BitfinexGetTicker init error: ", err)
	}

	_, err = b.GetTicker("wigwham")
	if err == nil {
		t.Error("Test Failed - GetTicker() error")
	}
}

func TestGetTickerV2(t *testing.T) {
	t.Parallel()
	_, err := b.GetTickerV2("tBTCUSD")
	if err != nil {
		t.Errorf("GetTickerV2 error: %s", err)
	}

	_, err = b.GetTickerV2("fUSD")
	if err != nil {
		t.Errorf("GetTickerV2 error: %s", err)
	}
}

func TestGetTickersV2(t *testing.T) {
	t.Parallel()
	_, err := b.GetTickersV2("tBTCUSD,fUSD")
	if err != nil {
		t.Errorf("GetTickersV2 error: %s", err)
	}
}

func TestGetStats(t *testing.T) {
	t.Parallel()
	_, err := b.GetStats("BTCUSD")
	if err != nil {
		t.Error("BitfinexGetStatsTest init error: ", err)
	}

	_, err = b.GetStats("wigwham")
	if err == nil {
		t.Error("Test Failed - GetStats() error")
	}
}

func TestGetFundingBook(t *testing.T) {
	t.Parallel()
	_, err := b.GetFundingBook("USD")
	if err != nil {
		t.Error("Testing Failed - GetFundingBook() error")
	}
	_, err = b.GetFundingBook("wigwham")
	if err == nil {
		t.Error("Testing Failed - GetFundingBook() error")
	}
}

func TestGetLendbook(t *testing.T) {
	t.Parallel()

	_, err := b.GetLendbook("BTCUSD", url.Values{})
	if err != nil {
		t.Error("Testing Failed - GetLendbook() error: ", err)
	}
}

func TestGetOrderbook(t *testing.T) {
	t.Parallel()

	_, err := b.GetOrderbook("BTCUSD", url.Values{})
	if err != nil {
		t.Error("BitfinexGetOrderbook init error: ", err)
	}
}

func TestGetOrderbookV2(t *testing.T) {
	t.Parallel()

	_, err := b.GetOrderbookV2("tBTCUSD", "P0", url.Values{})
	if err != nil {
		t.Errorf("GetOrderbookV2 error: %s", err)
	}

	_, err = b.GetOrderbookV2("fUSD", "P0", url.Values{})
	if err != nil {
		t.Errorf("GetOrderbookV2 error: %s", err)
	}
}

func TestGetTrades(t *testing.T) {
	t.Parallel()

	_, err := b.GetTrades("BTCUSD", url.Values{})
	if err != nil {
		t.Error("BitfinexGetTrades init error: ", err)
	}
}

func TestGetTradesv2(t *testing.T) {
	t.Parallel()

	_, err := b.GetTradesV2("tBTCUSD", 0, 0, true)
	if err != nil {
		t.Error("BitfinexGetTrades init error: ", err)
	}
}

func TestGetLends(t *testing.T) {
	t.Parallel()

	_, err := b.GetLends("BTC", url.Values{})
	if err != nil {
		t.Error("BitfinexGetLends init error: ", err)
	}
}

func TestGetSymbols(t *testing.T) {
	t.Parallel()

	symbols, err := b.GetSymbols()
	if err != nil {
		t.Fatal("BitfinexGetSymbols init error: ", err)
	}
	if reflect.TypeOf(symbols[0]).String() != "string" {
		t.Error("Bitfinex GetSymbols is not a string")
	}

	expectedCurrencies := []string{
		"rrtbtc",
		"zecusd",
		"zecbtc",
		"xmrusd",
		"xmrbtc",
		"dshusd",
		"dshbtc",
		"bccbtc",
		"bcubtc",
		"bccusd",
		"bcuusd",
		"btcusd",
		"ltcusd",
		"ltcbtc",
		"ethusd",
		"ethbtc",
		"etcbtc",
		"etcusd",
		"bfxusd",
		"bfxbtc",
		"rrtusd",
	}
	if len(expectedCurrencies) <= len(symbols) {

		for _, explicitSymbol := range expectedCurrencies {
			if common.StringDataCompare(expectedCurrencies, explicitSymbol) {
				break
			} else {
				t.Error("BitfinexGetSymbols currency mismatch with: ", explicitSymbol)
			}
		}
	} else {
		t.Error("BitfinexGetSymbols currency mismatch, Expected Currencies < Exchange Currencies")
	}
}

func TestGetSymbolsDetails(t *testing.T) {
	t.Parallel()

	_, err := b.GetSymbolsDetails()
	if err != nil {
		t.Error("BitfinexGetSymbolsDetails init error: ", err)
	}
}

func TestGetAccountInfo(t *testing.T) {
	if !b.ValidateAPICredentials() {
		t.SkipNow()
	}
	t.Parallel()

	_, err := b.GetAccountInfo()
	if err == nil {
		t.Error("Test Failed - GetAccountInfo error")
	}
}

func TestGetAccountFees(t *testing.T) {
	if !b.ValidateAPICredentials() {
		t.SkipNow()
	}
	t.Parallel()

	_, err := b.GetAccountFees()
	if err == nil {
		t.Error("Test Failed - GetAccountFees error")
	}
}

func TestGetAccountSummary(t *testing.T) {
	if !b.ValidateAPICredentials() {
		t.SkipNow()
	}
	t.Parallel()

	_, err := b.GetAccountSummary()
	if err == nil {
		t.Error("Test Failed - GetAccountSummary() error:")
	}
}

func TestNewDeposit(t *testing.T) {
	if !b.ValidateAPICredentials() {
		t.SkipNow()
	}
	t.Parallel()

	_, err := b.NewDeposit("blabla", "testwallet", 1)
	if err == nil {
		t.Error("Test Failed - NewDeposit() error:", err)
	}
}

func TestGetKeyPermissions(t *testing.T) {
	if !b.ValidateAPICredentials() {
		t.SkipNow()
	}
	t.Parallel()

	_, err := b.GetKeyPermissions()
	if err == nil {
		t.Error("Test Failed - GetKeyPermissions() error:")
	}
}

func TestGetMarginInfo(t *testing.T) {
	if !b.ValidateAPICredentials() {
		t.SkipNow()
	}
	t.Parallel()

	_, err := b.GetMarginInfo()
	if err == nil {
		t.Error("Test Failed - GetMarginInfo() error")
	}
}

func TestGetAccountBalance(t *testing.T) {
	if !b.ValidateAPICredentials() {
		t.SkipNow()
	}
	t.Parallel()

	_, err := b.GetAccountBalance()
	if err == nil {
		t.Error("Test Failed - GetAccountBalance() error")
	}
}

func TestWalletTransfer(t *testing.T) {
	if !b.ValidateAPICredentials() {
		t.SkipNow()
	}
	t.Parallel()

	_, err := b.WalletTransfer(0.01, "bla", "bla", "bla")
	if err == nil {
		t.Error("Test Failed - WalletTransfer() error")
	}
}

func TestWithdrawal(t *testing.T) {
	if !b.ValidateAPICredentials() {
		t.SkipNow()
	}
	t.Parallel()

	_, err := b.Withdrawal("LITECOIN", "deposit", "1000", 0.01)
	if err == nil {
		t.Error("Test Failed - Withdrawal() error")
	}
}

func TestNewOrder(t *testing.T) {
	if !b.ValidateAPICredentials() {
		t.SkipNow()
	}
	t.Parallel()

	_, err := b.NewOrder("BTCUSD", 1, 2, true, "market", false)
	if err == nil {
		t.Error("Test Failed - NewOrder() error")
	}
}

func TestNewOrderMulti(t *testing.T) {
	if !b.ValidateAPICredentials() {
		t.SkipNow()
	}
	t.Parallel()

	newOrder := []PlaceOrder{
		{
			Symbol:   "BTCUSD",
			Amount:   1,
			Price:    1,
			Exchange: "bitfinex",
			Side:     "buy",
			Type:     "market",
		},
	}

	_, err := b.NewOrderMulti(newOrder)
	if err == nil {
		t.Error("Test Failed - NewOrderMulti() error")
	}
}

func TestCancelOrder(t *testing.T) {
	if !b.ValidateAPICredentials() {
		t.SkipNow()
	}
	t.Parallel()

	_, err := b.CancelExistingOrder(1337)
	if err == nil {
		t.Error("Test Failed - CancelExistingOrder() error")
	}
}

func TestCancelMultipleOrders(t *testing.T) {
	if !b.ValidateAPICredentials() {
		t.SkipNow()
	}
	t.Parallel()

	_, err := b.CancelMultipleOrders([]int64{1337, 1336})
	if err == nil {
		t.Error("Test Failed - CancelMultipleOrders() error")
	}
}

func TestCancelAllOrders(t *testing.T) {
	if !b.ValidateAPICredentials() {
		t.SkipNow()
	}
	t.Parallel()

	_, err := b.CancelAllExistingOrders()
	if err == nil {
		t.Error("Test Failed - CancelAllExistingOrders() error")
	}
}

func TestReplaceOrder(t *testing.T) {
	if !b.ValidateAPICredentials() {
		t.SkipNow()
	}
	t.Parallel()

	_, err := b.ReplaceOrder(1337, "BTCUSD", 1, 1, true, "market", false)
	if err == nil {
		t.Error("Test Failed - ReplaceOrder() error")
	}
}

func TestGetOrderStatus(t *testing.T) {
	if !b.ValidateAPICredentials() {
		t.SkipNow()
	}
	t.Parallel()

	_, err := b.GetOrderStatus(1337)
	if err == nil {
		t.Error("Test Failed - GetOrderStatus() error")
	}
}

func TestGetActiveOrders(t *testing.T) {
	if !b.ValidateAPICredentials() {
		t.SkipNow()
	}
	t.Parallel()

	_, err := b.GetActiveOrders()
	if err == nil {
		t.Error("Test Failed - GetActiveOrders() error")
	}
}

func TestGetActivePositions(t *testing.T) {
	if !b.ValidateAPICredentials() {
		t.SkipNow()
	}
	t.Parallel()

	_, err := b.GetActivePositions()
	if err == nil {
		t.Error("Test Failed - GetActivePositions() error")
	}
}

func TestClaimPosition(t *testing.T) {
	if !b.ValidateAPICredentials() {
		t.SkipNow()
	}
	t.Parallel()

	_, err := b.ClaimPosition(1337)
	if err == nil {
		t.Error("Test Failed - ClaimPosition() error")
	}
}

func TestGetBalanceHistory(t *testing.T) {
	if !b.ValidateAPICredentials() {
		t.SkipNow()
	}
	t.Parallel()

	_, err := b.GetBalanceHistory("USD", time.Time{}, time.Time{}, 1, "deposit")
	if err == nil {
		t.Error("Test Failed - GetBalanceHistory() error")
	}
}

func TestGetMovementHistory(t *testing.T) {
	if !b.ValidateAPICredentials() {
		t.SkipNow()
	}
	t.Parallel()

	_, err := b.GetMovementHistory("USD", "bitcoin", time.Time{}, time.Time{}, 1)
	if err == nil {
		t.Error("Test Failed - GetMovementHistory() error")
	}
}

func TestGetTradeHistory(t *testing.T) {
	if !b.ValidateAPICredentials() {
		t.SkipNow()
	}
	t.Parallel()

	_, err := b.GetTradeHistory("BTCUSD", time.Time{}, time.Time{}, 1, 0)
	if err == nil {
		t.Error("Test Failed - GetTradeHistory() error")
	}
}

func TestNewOffer(t *testing.T) {
	if !b.ValidateAPICredentials() {
		t.SkipNow()
	}
	t.Parallel()

	_, err := b.NewOffer("BTC", 1, 1, 1, "loan")
	if err == nil {
		t.Error("Test Failed - NewOffer() error")
	}
}

func TestCancelOffer(t *testing.T) {
	if !b.ValidateAPICredentials() {
		t.SkipNow()
	}
	t.Parallel()

	_, err := b.CancelOffer(1337)
	if err == nil {
		t.Error("Test Failed - CancelOffer() error")
	}
}

func TestGetOfferStatus(t *testing.T) {
	if !b.ValidateAPICredentials() {
		t.SkipNow()
	}
	t.Parallel()

	_, err := b.GetOfferStatus(1337)
	if err == nil {
		t.Error("Test Failed - NewOffer() error")
	}
}

func TestGetActiveCredits(t *testing.T) {
	if !b.ValidateAPICredentials() {
		t.SkipNow()
	}
	t.Parallel()

	_, err := b.GetActiveCredits()
	if err == nil {
		t.Error("Test Failed - GetActiveCredits() error", err)
	}
}

func TestGetActiveOffers(t *testing.T) {
	if !b.ValidateAPICredentials() {
		t.SkipNow()
	}
	t.Parallel()

	_, err := b.GetActiveOffers()
	if err == nil {
		t.Error("Test Failed - GetActiveOffers() error", err)
	}
}

func TestGetActiveMarginFunding(t *testing.T) {
	if !b.ValidateAPICredentials() {
		t.SkipNow()
	}
	t.Parallel()

	_, err := b.GetActiveMarginFunding()
	if err == nil {
		t.Error("Test Failed - GetActiveMarginFunding() error", err)
	}
}

func TestGetUnusedMarginFunds(t *testing.T) {
	if !b.ValidateAPICredentials() {
		t.SkipNow()
	}
	t.Parallel()

	_, err := b.GetUnusedMarginFunds()
	if err == nil {
		t.Error("Test Failed - GetUnusedMarginFunds() error", err)
	}
}

func TestGetMarginTotalTakenFunds(t *testing.T) {
	if !b.ValidateAPICredentials() {
		t.SkipNow()
	}
	t.Parallel()

	_, err := b.GetMarginTotalTakenFunds()
	if err == nil {
		t.Error("Test Failed - GetMarginTotalTakenFunds() error", err)
	}
}

func TestCloseMarginFunding(t *testing.T) {
	if !b.ValidateAPICredentials() {
		t.SkipNow()
	}
	t.Parallel()

	_, err := b.CloseMarginFunding(1337)
	if err == nil {
		t.Error("Test Failed - CloseMarginFunding() error")
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
	b.SetDefaults()
	b.Verbose = true
	TestSetup(t)
	var feeBuilder = setFeeBuilder()

	if testAPIKey != "" || testAPISecret != "" {
		// CryptocurrencyTradeFee Basic
		if resp, err := b.GetFee(feeBuilder); resp != float64(0.002) || err != nil {
			t.Error(err)
			t.Errorf("Test Failed - GetFee() error. Expected: %f, Recieved: %f", float64(0.002), resp)
		}

		// CryptocurrencyTradeFee High quantity
		feeBuilder = setFeeBuilder()
		feeBuilder.Amount = 1000
		feeBuilder.PurchasePrice = 1000
		if resp, err := b.GetFee(feeBuilder); resp != float64(2000) || err != nil {
			t.Errorf("Test Failed - GetFee() error. Expected: %f, Recieved: %f", float64(2000), resp)
			t.Error(err)
		}

		// CryptocurrencyTradeFee IsMaker
		feeBuilder = setFeeBuilder()
		feeBuilder.IsMaker = true
		if resp, err := b.GetFee(feeBuilder); resp != float64(0.001) || err != nil {
			t.Errorf("Test Failed - GetFee() error. Expected: %f, Recieved: %f", float64(0.001), resp)
			t.Error(err)
		}

		// CryptocurrencyTradeFee Negative purchase price
		feeBuilder = setFeeBuilder()
		feeBuilder.PurchasePrice = -1000
		if resp, err := b.GetFee(feeBuilder); resp != float64(0) || err != nil {
			t.Errorf("Test Failed - GetFee() error. Expected: %f, Recieved: %f", float64(0), resp)
			t.Error(err)
		}

		// CryptocurrencyWithdrawalFee Basic
		feeBuilder = setFeeBuilder()
		feeBuilder.FeeType = exchange.CryptocurrencyWithdrawalFee
		if resp, err := b.GetFee(feeBuilder); resp != float64(0.0004) || err != nil {
			t.Errorf("Test Failed - GetFee() error. Expected: %f, Recieved: %f", float64(0.0004), resp)
			t.Error(err)
		}
	}

	// CyptocurrencyDepositFee Basic
	feeBuilder = setFeeBuilder()
	feeBuilder.FeeType = exchange.CyptocurrencyDepositFee
	if resp, err := b.GetFee(feeBuilder); resp != float64(0) || err != nil {
		t.Errorf("Test Failed - GetFee() error. Expected: %f, Recieved: %f", float64(0), resp)
		t.Error(err)
	}

	// InternationalBankDepositFee Basic
	feeBuilder = setFeeBuilder()
	feeBuilder.FeeType = exchange.InternationalBankDepositFee
	feeBuilder.CurrencyItem = symbol.HKD
	if resp, err := b.GetFee(feeBuilder); resp != float64(0.001) || err != nil {
		t.Errorf("Test Failed - GetFee() error. Expected: %f, Recieved: %f", float64(0.001), resp)
		t.Error(err)
	}

	// InternationalBankWithdrawalFee Basic
	feeBuilder = setFeeBuilder()
	feeBuilder.FeeType = exchange.InternationalBankWithdrawalFee
	feeBuilder.CurrencyItem = symbol.HKD
	if resp, err := b.GetFee(feeBuilder); resp != float64(0.001) || err != nil {
		t.Errorf("Test Failed - GetFee() error. Expected: %f, Recieved: %f", float64(0.001), resp)
		t.Error(err)
	}
}

func TestFormatWithdrawPermissions(t *testing.T) {
	// Arrange
	b.SetDefaults()
	expectedResult := exchange.AutoWithdrawCryptoWithAPIPermissionText + " & " + exchange.AutoWithdrawFiatWithAPIPermissionText
	// Act
	withdrawPermissions := b.FormatWithdrawPermissions()
	// Assert
	if withdrawPermissions != expectedResult {
		t.Errorf("Expected: %s, Recieved: %s", expectedResult, withdrawPermissions)
	}
}

// Any tests below this line have the ability to impact your orders on the exchange. Enable canManipulateRealOrders to run them
// ----------------------------------------------------------------------------------------------------------------------------
func isRealOrderTestEnabled() bool {
	if !b.ValidateAPICredentials() || !canManipulateRealOrders {
		return false
	}
	return true
}

func TestSubmitOrder(t *testing.T) {
	b.SetDefaults()
	TestSetup(t)

	if !isRealOrderTestEnabled() {
		t.Skip()
	}

	var p = pair.CurrencyPair{
		Delimiter:      "",
		FirstCurrency:  symbol.LTC,
		SecondCurrency: symbol.BTC,
	}

	response, err := b.SubmitOrder(p, exchange.Buy, exchange.Market, 1, 1, "clientId")
	if err != nil || !response.IsOrderPlaced {
		t.Errorf("Order failed to be placed: %v", err)
	}
}

func TestCancelExchangeOrder(t *testing.T) {
	// Arrange
	b.SetDefaults()
	TestSetup(t)

	if !isRealOrderTestEnabled() {
		t.Skip()
	}

	b.Verbose = true
	currencyPair := pair.NewCurrencyPair(symbol.LTC, symbol.BTC)

	var orderCancellation = exchange.OrderCancellation{
		OrderID:       "1",
		WalletAddress: "1F5zVDgNjorJ51oGebSvNCrSAHpwGkUdDB",
		AccountID:     "1",
		CurrencyPair:  currencyPair,
	}

	// Act
	err := b.CancelOrder(orderCancellation)

	// Assert
	if err != nil {
		t.Errorf("Could not cancel order: %s", err)
	}
}
