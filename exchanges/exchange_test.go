package exchange

import (
	"net/http"
	"testing"
	"time"

	"github.com/thrasher-/gocryptotrader/common"
	"github.com/thrasher-/gocryptotrader/config"
	"github.com/thrasher-/gocryptotrader/currency/pair"
	"github.com/thrasher-/gocryptotrader/exchanges/assets"
	"github.com/thrasher-/gocryptotrader/exchanges/request"
)

func TestSupportsRESTTickerBatchUpdates(t *testing.T) {
	b := Base{
		Name: "RAWR",
		Features: Features{
			Supports: FeaturesSupported{
				RESTTickerBatching: true,
			},
		},
	}

	if !b.SupportsRESTTickerBatchUpdates() {
		t.Fatal("Test failed. TestSupportsRESTTickerBatchUpdates returned false")
	}
}

func TestHTTPClient(t *testing.T) {
	r := Base{Name: "asdf"}
	r.SetHTTPClientTimeout(time.Duration(time.Second * 5))

	if r.GetHTTPClient().Timeout != time.Second*5 {
		t.Fatalf("Test failed. TestHTTPClient unexpected value")
	}

	r.Requester = nil
	newClient := new(http.Client)
	newClient.Timeout = time.Duration(time.Second * 10)

	r.SetHTTPClient(newClient)
	if r.GetHTTPClient().Timeout != time.Second*10 {
		t.Fatalf("Test failed. TestHTTPClient unexpected value")
	}

	r.Requester = nil
	if r.GetHTTPClient() == nil {
		t.Fatalf("Test failed. TestHTTPClient unexpected value")
	}

	b := Base{Name: "RAWR"}
	b.Requester = request.New(b.Name,
		request.NewRateLimit(time.Second, 1),
		request.NewRateLimit(time.Second, 1),
		new(http.Client))

	b.SetHTTPClientTimeout(time.Second * 5)
	if b.GetHTTPClient().Timeout != time.Second*5 {
		t.Fatalf("Test failed. TestHTTPClient unexpected value")
	}

	newClient = new(http.Client)
	newClient.Timeout = time.Duration(time.Second * 10)

	b.SetHTTPClient(newClient)
	if b.GetHTTPClient().Timeout != time.Second*10 {
		t.Fatalf("Test failed. TestHTTPClient unexpected value")
	}
}

func TestSetClientProxyAddress(t *testing.T) {
	requester := request.New("testicles",
		&request.RateLimit{},
		&request.RateLimit{},
		&http.Client{})

	newBase := Base{Name: "Testicles", Requester: requester}

	newBase.WebsocketInit()

	err := newBase.SetClientProxyAddress(":invalid")
	if err == nil {
		t.Error("Test failed. SetClientProxyAddress parsed invalid URL")
	}

	if newBase.Websocket.GetProxyAddress() != "" {
		t.Error("Test failed. SetClientProxyAddress error", err)
	}

	err = newBase.SetClientProxyAddress("www.valid.com")
	if err != nil {
		t.Error("Test failed. SetClientProxyAddress error", err)
	}

	if newBase.Websocket.GetProxyAddress() != "www.valid.com" {
		t.Error("Test failed. SetClientProxyAddress error", err)
	}
}

func TestSetAutoPairDefaults(t *testing.T) {
	cfg := config.GetConfig()
	err := cfg.LoadConfig(config.ConfigTestFile)
	if err != nil {
		t.Fatalf("Test failed. TestSetAutoPairDefaults failed to load config file. Error: %s", err)
	}

	b := Base{
		Name: "TESTNAME",
		Features: Features{
			Supports: FeaturesSupported{
				AutoPairUpdates: true,
			},
		},
	}

	err = b.SetFeatureDefaults()
	if err == nil {
		t.Fatal("Test failed. TestSetAutoPairDefaults returned nil error for a non-existent exchange")
	}

	b.Name = "Bitstamp"
	err = b.SetFeatureDefaults()
	if err != nil {
		t.Fatalf("Test failed. TestSetAutoPairDefaults. Error %s", err)
	}

	exch, err := cfg.GetExchangeConfig(b.Name)
	if err != nil {
		t.Fatalf("Test failed. TestSetAutoPairDefaults load config failed. Error %s", err)
	}

	if !exch.Features.Supports.AutoPairUpdates {
		t.Fatalf("Test failed. TestSetAutoPairDefaults Incorrect value")
	}

	if exch.CurrencyPairs.LastUpdated != 0 {
		t.Fatalf("Test failed. TestSetAutoPairDefaults Incorrect value")
	}

	exch.Features.Supports.AutoPairUpdates = false
	err = cfg.UpdateExchangeConfig(exch)
	if err != nil {
		t.Fatalf("Test failed. TestSetAutoPairDefaults update config failed. Error %s", err)
	}

	exch, err = cfg.GetExchangeConfig(b.Name)
	if err != nil {
		t.Fatalf("Test failed. TestSetAutoPairDefaults load config failed. Error %s", err)
	}

	if exch.Features.Supports.AutoPairUpdates != false {
		t.Fatal("Test failed. TestSetAutoPairDefaults Incorrect value")
	}

	err = b.SetFeatureDefaults()
	if err != nil {
		t.Fatalf("Test failed. TestSetAutoPairDefaults. Error %s", err)
	}

	exch, err = cfg.GetExchangeConfig(b.Name)
	if err != nil {
		t.Fatalf("Test failed. TestSetAutoPairDefaults load config failed. Error %s", err)
	}

	if exch.Features.Supports.AutoPairUpdates == false {
		t.Fatal("Test failed. TestSetAutoPairDefaults Incorrect value")
	}

	b.Name = "Bitflyer"
	err = b.SetFeatureDefaults()
	if err != nil {
		t.Fatalf("Test failed. TestSetAutoPairDefaults. Error %s", err)
	}

	if !exch.Features.Supports.AutoPairUpdates {
		t.Fatal("Test failed. TestSetAutoPairDefaults Incorrect value")
	}

	if b.PairsLastUpdated != 0 {
		t.Fatal("Test failed. TestSetAutoPairDefaults Incorrect value")
	}
}

func TestSupportsAutoPairUpdates(t *testing.T) {
	b := Base{
		Name: "TESTNAME",
	}

	if b.SupportsAutoPairUpdates() {
		t.Fatal("Test failed. TestSupportsAutoPairUpdates Incorrect value")
	}
}

func TestGetLastPairsUpdateTime(t *testing.T) {
	testTime := time.Now().Unix()
	b := Base{
		Name:             "TESTNAME",
		PairsLastUpdated: testTime,
	}

	if b.GetLastPairsUpdateTime() != testTime {
		t.Fatal("Test failed. TestGetLastPairsUpdateTim Incorrect value")
	}
}

func TestSetAssetTypes(t *testing.T) {
	cfg := config.GetConfig()
	err := cfg.LoadConfig(config.ConfigTestFile)
	if err != nil {
		t.Fatalf("Test failed. TestSetAssetTypes failed to load config file. Error: %s", err)
	}

	b := Base{
		Name: "TESTNAME",
	}

	err = b.SetAssetTypes()
	if err == nil {
		t.Fatal("Test failed. TestSetAssetTypes returned nil error for a non-existent exchange")
	}

	b.Name = "ANX"
	b.CurrencyPairs.AssetTypes = assets.AssetTypes{assets.AssetTypeSpot}
	err = b.SetAssetTypes()
	if err != nil {
		t.Fatalf("Test failed. TestSetAssetTypes. Error %s", err)
	}

	exch, err := cfg.GetExchangeConfig(b.Name)
	if err != nil {
		t.Fatalf("Test failed. TestSetAssetTypes load config failed. Error %s", err)
	}

	exch.CurrencyPairs.AssetTypes = ""
	err = cfg.UpdateExchangeConfig(exch)
	if err != nil {
		t.Fatalf("Test failed. TestSetAssetTypes update config failed. Error %s", err)
	}

	exch, err = cfg.GetExchangeConfig(b.Name)
	if err != nil {
		t.Fatalf("Test failed. TestSetAssetTypes load config failed. Error %s", err)
	}

	if exch.CurrencyPairs.AssetTypes != "" {
		t.Fatal("Test failed. TestSetAssetTypes assetTypes != ''")
	}

	err = b.SetAssetTypes()
	if err != nil {
		t.Fatalf("Test failed. TestSetAssetTypes. Error %s", err)
	}

	if !common.StringDataCompare(b.CurrencyPairs.AssetTypes.ToStringArray(), assets.AssetTypeSpot.String()) {
		t.Fatal("Test failed. TestSetAssetTypes assetTypes is not set")
	}
}

func TestGetAssetTypes(t *testing.T) {
	testExchange := Base{
		CurrencyPairs: CurrencyPairs{
			AssetTypes: assets.AssetTypes{
				assets.AssetTypeSpot,
				assets.AssetTypeBinary,
				assets.AssetTypeFutures,
			},
		},
	}

	aT := testExchange.GetAssetTypes()
	if len(aT) != 3 {
		t.Error("Test failed. TestGetAssetTypes failed")
	}
}

func TestGetExchangeAssetTypes(t *testing.T) {
	cfg := config.GetConfig()
	err := cfg.LoadConfig(config.ConfigTestFile)
	if err != nil {
		t.Fatalf("Failed to load config file. Error: %s", err)
	}

	result, err := GetExchangeAssetTypes("Bitfinex")
	if err != nil {
		t.Fatal("Test failed. Unable to obtain Bitfinex asset types")
	}

	if !common.StringDataCompareUpper(result.ToStringArray(), assets.AssetTypeSpot.String()) {
		t.Fatal("Test failed. Bitfinex does not contain default asset type 'SPOT'")
	}

	_, err = GetExchangeAssetTypes("non-existent-exchange")
	if err == nil {
		t.Fatal("Test failed. Got asset types for non-existent exchange")
	}
}

func TestCompareCurrencyPairFormats(t *testing.T) {
	cfgOne := config.CurrencyPairFormatConfig{
		Delimiter: "-",
		Uppercase: true,
		Index:     "",
		Separator: ",",
	}

	cfgTwo := cfgOne
	if !CompareCurrencyPairFormats(cfgOne, &cfgTwo) {
		t.Fatal("Test failed. CompareCurrencyPairFormats should be true")
	}

	cfgTwo.Delimiter = "~"
	if CompareCurrencyPairFormats(cfgOne, &cfgTwo) {
		t.Fatal("Test failed. CompareCurrencyPairFormats should not be true")
	}
}

func TestSetCurrencyPairFormat(t *testing.T) {
	cfg := config.GetConfig()
	err := cfg.LoadConfig(config.ConfigTestFile)
	if err != nil {
		t.Fatalf("Test failed. TestSetCurrencyPairFormat failed to load config file. Error: %s", err)
	}

	b := Base{
		Name: "TESTNAME",
	}

	err = b.SetCurrencyPairFormat()
	if err == nil {
		t.Fatal("Test failed. TestSetCurrencyPairFormat returned nil error for a non-existent exchange")
	}

	b.Name = "ANX"
	err = b.SetCurrencyPairFormat()
	if err != nil {
		t.Fatalf("Test failed. TestSetCurrencyPairFormat. Error %s", err)
	}

	exch, err := cfg.GetExchangeConfig(b.Name)
	if err != nil {
		t.Fatalf("Test failed. TestSetCurrencyPairFormat load config failed. Error %s", err)
	}

	exch.ConfigCurrencyPairFormat = nil
	exch.RequestCurrencyPairFormat = nil
	err = cfg.UpdateExchangeConfig(exch)
	if err != nil {
		t.Fatalf("Test failed. TestSetCurrencyPairFormat update config failed. Error %s", err)
	}

	exch, err = cfg.GetExchangeConfig(b.Name)
	if err != nil {
		t.Fatalf("Test failed. TestSetCurrencyPairFormat load config failed. Error %s", err)
	}

	if exch.ConfigCurrencyPairFormat != nil && exch.RequestCurrencyPairFormat != nil {
		t.Fatal("Test failed. TestSetCurrencyPairFormat exch values are not nil")
	}

	err = b.SetCurrencyPairFormat()
	if err != nil {
		t.Fatalf("Test failed. TestSetCurrencyPairFormat. Error %s", err)
	}

	if b.CurrencyPairs.ConfigFormat.Delimiter != "" &&
		b.CurrencyPairs.ConfigFormat.Index != "BTC" &&
		b.CurrencyPairs.ConfigFormat.Uppercase {
		t.Fatal("Test failed. TestSetCurrencyPairFormat ConfigCurrencyPairFormat values are incorrect")
	}

	if b.CurrencyPairs.ConfigFormat.Delimiter != "" &&
		b.CurrencyPairs.ConfigFormat.Index != "BTC" &&
		b.CurrencyPairs.ConfigFormat.Uppercase {
		t.Fatal("Test failed. TestSetCurrencyPairFormat RequestCurrencyPairFormat values are incorrect")
	}

	// if currency pairs are the same as the config, should load from config
	err = b.SetCurrencyPairFormat()
	if err != nil {
		t.Fatalf("Test failed. TestSetCurrencyPairFormat. Error %s", err)
	}
}

func TestGetAuthenticatedAPISupport(t *testing.T) {
	var base Base
	if base.GetAuthenticatedAPISupport() {
		t.Fatal("Test failed. TestGetAuthenticatedAPISupport returned true when it should of been false.")
	}
}

func TestGetName(t *testing.T) {
	GetName := Base{
		Name: "TESTNAME",
	}

	name := GetName.GetName()
	if name != "TESTNAME" {
		t.Error("Test Failed - Exchange GetName() returned incorrect name")
	}
}

func TestGetEnabledPairs(t *testing.T) {
	b := Base{
		Name: "TESTNAME",
	}

	b.CurrencyPairs.Spot.Enabled = []string{"BTC-USD"}
	format := config.CurrencyPairFormatConfig{
		Delimiter: "-",
		Index:     "",
	}

	assetType := assets.AssetTypeSpot
	b.CurrencyPairs.UseGlobalPairFormat = true
	b.CurrencyPairs.RequestFormat = format
	b.CurrencyPairs.ConfigFormat = format

	c := b.GetEnabledPairs(assetType)
	if c[0].Pair().String() != "BTC-USD" {
		t.Error("Test Failed - Exchange GetAvailablePairs() incorrect string")
	}

	format.Delimiter = "~"
	b.CurrencyPairs.RequestFormat = format
	c = b.GetEnabledPairs(assetType)
	if c[0].Pair().String() != "BTC-USD" {
		t.Error("Test Failed - Exchange GetAvailablePairs() incorrect string")
	}

	format.Delimiter = ""
	b.CurrencyPairs.ConfigFormat = format
	c = b.GetEnabledPairs(assetType)
	if c[0].Pair().String() != "BTC-USD" {
		t.Error("Test Failed - Exchange GetAvailablePairs() incorrect string")
	}

	b.CurrencyPairs.Spot.Enabled = []string{"BTCDOGE"}
	format.Index = "BTC"
	b.CurrencyPairs.ConfigFormat = format
	c = b.GetEnabledPairs(assetType)
	if c[0].FirstCurrency.String() != "BTC" && c[0].SecondCurrency.String() != "DOGE" {
		t.Error("Test Failed - Exchange GetAvailablePairs() incorrect string")
	}

	b.CurrencyPairs.Spot.Enabled = []string{"BTC_USD"}
	b.CurrencyPairs.RequestFormat.Delimiter = ""
	b.CurrencyPairs.ConfigFormat.Delimiter = "_"
	c = b.GetEnabledPairs(assetType)
	if c[0].FirstCurrency.String() != "BTC" && c[0].SecondCurrency.String() != "USD" {
		t.Error("Test Failed - Exchange GetAvailablePairs() incorrect string")
	}

	b.CurrencyPairs.Spot.Enabled = []string{"BTCDOGE"}
	b.CurrencyPairs.RequestFormat.Delimiter = ""
	b.CurrencyPairs.ConfigFormat.Delimiter = ""
	b.CurrencyPairs.ConfigFormat.Index = "BTC"
	c = b.GetEnabledPairs(assetType)
	if c[0].FirstCurrency.String() != "BTC" && c[0].SecondCurrency.String() != "DOGE" {
		t.Error("Test Failed - Exchange GetAvailablePairs() incorrect string")
	}

	b.CurrencyPairs.Spot.Enabled = []string{"BTCUSD"}
	b.CurrencyPairs.ConfigFormat.Index = ""
	c = b.GetEnabledPairs(assetType)
	if c[0].FirstCurrency.String() != "BTC" && c[0].SecondCurrency.String() != "USD" {
		t.Error("Test Failed - Exchange GetAvailablePairs() incorrect string")
	}
}

func TestGetAvailablePairs(t *testing.T) {
	b := Base{
		Name: "TESTNAME",
	}

	b.CurrencyPairs.Spot.Available = []string{"BTC-USD"}
	format := config.CurrencyPairFormatConfig{
		Delimiter: "-",
		Index:     "",
	}

	assetType := assets.AssetTypeSpot
	b.CurrencyPairs.UseGlobalPairFormat = true
	b.CurrencyPairs.RequestFormat = format
	b.CurrencyPairs.ConfigFormat = format

	c := b.GetAvailablePairs(assetType)
	if c[0].Pair().String() != "BTC-USD" {
		t.Error("Test Failed - Exchange GetAvailablePairs() incorrect string")
	}

	format.Delimiter = "~"
	b.CurrencyPairs.RequestFormat = format
	c = b.GetAvailablePairs(assetType)
	if c[0].Pair().String() != "BTC-USD" {
		t.Error("Test Failed - Exchange GetAvailablePairs() incorrect string")
	}

	format.Delimiter = ""
	b.CurrencyPairs.ConfigFormat = format
	c = b.GetAvailablePairs(assetType)
	if c[0].Pair().String() != "BTC-USD" {
		t.Error("Test Failed - Exchange GetAvailablePairs() incorrect string")
	}

	b.CurrencyPairs.Spot.Available = []string{"BTCDOGE"}
	format.Index = "BTC"
	b.CurrencyPairs.ConfigFormat = format
	c = b.GetAvailablePairs(assetType)
	if c[0].FirstCurrency.String() != "BTC" && c[0].SecondCurrency.String() != "DOGE" {
		t.Error("Test Failed - Exchange GetAvailablePairs() incorrect string")
	}

	b.CurrencyPairs.Spot.Available = []string{"BTC_USD"}
	b.CurrencyPairs.RequestFormat.Delimiter = ""
	b.CurrencyPairs.ConfigFormat.Delimiter = "_"
	c = b.GetAvailablePairs(assetType)
	if c[0].FirstCurrency.String() != "BTC" && c[0].SecondCurrency.String() != "USD" {
		t.Error("Test Failed - Exchange GetAvailablePairs() incorrect string")
	}

	b.CurrencyPairs.Spot.Available = []string{"BTCDOGE"}
	b.CurrencyPairs.RequestFormat.Delimiter = ""
	b.CurrencyPairs.ConfigFormat.Delimiter = ""
	b.CurrencyPairs.ConfigFormat.Index = "BTC"
	c = b.GetAvailablePairs(assetType)
	if c[0].FirstCurrency.String() != "BTC" && c[0].SecondCurrency.String() != "DOGE" {
		t.Error("Test Failed - Exchange GetAvailablePairs() incorrect string")
	}

	b.CurrencyPairs.Spot.Available = []string{"BTCUSD"}
	b.CurrencyPairs.ConfigFormat.Index = ""
	c = b.GetAvailablePairs(assetType)
	if c[0].FirstCurrency.String() != "BTC" && c[0].SecondCurrency.String() != "USD" {
		t.Error("Test Failed - Exchange GetAvailablePairs() incorrect string")
	}
}

func TestSupportsPair(t *testing.T) {
	b := Base{
		Name: "TESTNAME",
	}

	b.CurrencyPairs.Spot.Available = []string{"BTC-USD", "ETH-USD"}
	b.CurrencyPairs.Spot.Enabled = []string{"BTC-USD"}

	format := config.CurrencyPairFormatConfig{
		Delimiter: "-",
		Index:     "",
	}

	b.CurrencyPairs.UseGlobalPairFormat = true
	b.CurrencyPairs.RequestFormat = format
	b.CurrencyPairs.ConfigFormat = format
	assetType := assets.AssetTypeSpot

	if !b.SupportsPair(pair.NewCurrencyPair("BTC", "USD"), true, assetType) {
		t.Error("Test Failed - Exchange SupportsPair() incorrect value")
	}

	if !b.SupportsPair(pair.NewCurrencyPair("ETH", "USD"), false, assetType) {
		t.Error("Test Failed - Exchange SupportsPair() incorrect value")
	}

	if b.SupportsPair(pair.NewCurrencyPair("ASD", "ASDF"), true, assetType) {
		t.Error("Test Failed - Exchange SupportsPair() incorrect value")
	}
}
func TestGetExchangeFormatCurrencySeperator(t *testing.T) {
	cfg := config.GetConfig()
	err := cfg.LoadConfig(config.ConfigTestFile)
	if err != nil {
		t.Fatalf("Failed to load config file. Error: %s", err)
	}

	expected := true
	actual := GetExchangeFormatCurrencySeperator("WEX")

	if expected != actual {
		t.Errorf("Test failed - TestGetExchangeFormatCurrencySeperator expected %v != actual %v",
			expected, actual)
	}

	expected = false
	actual = GetExchangeFormatCurrencySeperator("LocalBitcoins")

	if expected != actual {
		t.Errorf("Test failed - TestGetExchangeFormatCurrencySeperator expected %v != actual %v",
			expected, actual)
	}

	expected = false
	actual = GetExchangeFormatCurrencySeperator("blah")

	if expected != actual {
		t.Errorf("Test failed - TestGetExchangeFormatCurrencySeperator expected %v != actual %v",
			expected, actual)
	}
}

func TestFormatExchangeCurrencies(t *testing.T) {
	cfg := config.GetConfig()
	err := cfg.LoadConfig(config.ConfigTestFile)
	if err != nil {
		t.Fatalf("Failed to load config file. Error: %s", err)
	}

	var pairs []pair.CurrencyPair
	pairs = append(pairs, pair.NewCurrencyPairDelimiter("BTC_USD", "_"))
	pairs = append(pairs, pair.NewCurrencyPairDelimiter("LTC_BTC", "_"))
	assetType := assets.AssetTypeSpot

	actual, err := FormatExchangeCurrencies("Liqui", pairs, assetType)
	if err != nil {
		t.Errorf("Test failed - Exchange TestFormatExchangeCurrencies error %s", err)
	}
	expected := pair.CurrencyItem("btc_usd-ltc_btc")

	if actual.String() != expected.String() {
		t.Errorf("Test failed - Exchange TestFormatExchangeCurrencies %s != %s",
			actual, expected)
	}

	_, err = FormatExchangeCurrencies("non-existent", pairs, assetType)
	if err == nil {
		t.Errorf("Test failed - Exchange TestFormatExchangeCurrencies returned nil error on non-existent exchange")
	}
}

func TestFormatExchangeCurrency(t *testing.T) {
	cfg := config.GetConfig()
	err := cfg.LoadConfig(config.ConfigTestFile)
	if err != nil {
		t.Fatalf("Failed to load config file. Error: %s", err)
	}

	pair := pair.NewCurrencyPair("BTC", "USD")
	expected := "BTC-USD"
	actual := FormatExchangeCurrency("CoinbasePro", pair, assets.AssetTypeSpot)

	if actual.String() != expected {
		t.Errorf("Test failed - Exchange TestFormatExchangeCurrency %s != %s",
			actual, expected)
	}
}

func TestFormatCurrency(t *testing.T) {
	cfg := config.GetConfig()
	err := cfg.LoadConfig(config.ConfigTestFile)
	if err != nil {
		t.Fatalf("Failed to load config file. Error: %s", err)
	}

	currency := pair.NewCurrencyPair("btc", "usd")
	expected := "BTC-USD"
	actual := FormatCurrency(currency).String()
	if actual != expected {
		t.Errorf("Test failed - Exchange TestFormatCurrency %s != %s",
			actual, expected)
	}
}

func TestSetEnabled(t *testing.T) {
	SetEnabled := Base{
		Name:    "TESTNAME",
		Enabled: false,
	}

	SetEnabled.SetEnabled(true)
	if !SetEnabled.Enabled {
		t.Error("Test Failed - Exchange SetEnabled(true) did not set boolean")
	}
}

func TestIsEnabled(t *testing.T) {
	IsEnabled := Base{
		Name:    "TESTNAME",
		Enabled: false,
	}

	if IsEnabled.IsEnabled() {
		t.Error("Test Failed - Exchange IsEnabled() did not return correct boolean")
	}
}

func TestSetAPIKeys(t *testing.T) {
	SetAPIKeys := Base{
		Name:    "TESTNAME",
		Enabled: false,
	}

	SetAPIKeys.SetAPIKeys("RocketMan", "Digereedoo", "007")
	if SetAPIKeys.API.Credentials.Key != "RocketMan" && SetAPIKeys.API.Credentials.Secret != "Digereedoo" && SetAPIKeys.API.Credentials.ClientID != "007" {
		t.Error("Test Failed - SetAPIKeys() unable to set API credentials")
	}

	SetAPIKeys.API.CredentialsValidator.RequiresBase64DecodeSecret = true
	SetAPIKeys.SetAPIKeys("RocketMan", "Digereedoo", "007")
}

func TestSetPairs(t *testing.T) {
	cfg := config.GetConfig()
	err := cfg.LoadConfig(config.ConfigTestFile)
	if err != nil {
		t.Fatal("Test failed. TestSetPairs failed to load config")
	}

	UAC := Base{Name: "ASDF"}
	UAC.CurrencyPairs.Spot.Available = []string{"ETHLTC", "LTCBTC"}
	UAC.CurrencyPairs.Spot.Enabled = []string{"ETHLTC"}
	newPair := pair.NewCurrencyPairDelimiter("ETH_USDT", "_")
	assetType := assets.AssetTypeSpot

	err = UAC.SetPairs([]pair.CurrencyPair{newPair}, true)
	if err == nil {
		t.Fatal("Test failed. TestSetPairs returned nil error on non-existent exchange")
	}

	anxCfg, err := cfg.GetExchangeConfig("ANX")
	if err != nil {
		t.Fatal("Test failed. TestSetPairs failed to load config")
	}

	UAC.Name = "ANX"
	UAC.CurrencyPairs.UseGlobalPairFormat = true
	UAC.CurrencyPairs.ConfigFormat.Delimiter = anxCfg.CurrencyPairs.ConfigFormat.Delimiter
	err = UAC.SetPairs([]pair.CurrencyPair{newPair}, true)
	if err != nil {
		t.Fatalf("Test failed. TestSetPairs failed to set currencies: %s", err)
	}

	if !pair.Contains(UAC.GetEnabledPairs(assetType), newPair, true) {
		t.Fatal("Test failed. TestSetPairs failed to set currencies")
	}

	UAC.SetPairs([]pair.CurrencyPair{newPair}, false)
	if !pair.Contains(UAC.GetAvailablePairs(assetType), newPair, true) {
		t.Fatal("Test failed. TestSetPairs failed to set currencies")
	}

	err = UAC.SetPairs(nil, false)
	if err == nil {
		t.Fatal("Test failed. TestSetPairs should return an error when attempting to set an empty pairs array")
	}
}

func TestUpdatePairs(t *testing.T) {
	cfg := config.GetConfig()
	err := cfg.LoadConfig(config.ConfigTestFile)
	if err != nil {
		t.Fatal("Test failed. TestUpdateEnabledCurrencies failed to load config")
	}

	UAC := Base{Name: "ANX"}
	exchangeProducts := []string{"ltc", "btc", "usd", "aud", ""}

	// Test updating exchange products for an exchange which doesn't exist
	UAC.Name = "Blah"
	err = UAC.UpdatePairs(exchangeProducts, true, false)
	if err == nil {
		t.Errorf("Test Failed - Exchange TestUpdatePairs succeeded on an exchange which doesn't exist")
	}

	// Test updating exchange products
	UAC.Name = "ANX"
	err = UAC.UpdatePairs(exchangeProducts, true, false)
	if err != nil {
		t.Errorf("Test Failed - Exchange TestUpdatePairs error: %s", err)
	}

	// Test updating the same new products, diff should be 0
	UAC.Name = "ANX"
	err = UAC.UpdatePairs(exchangeProducts, true, false)
	if err != nil {
		t.Errorf("Test Failed - Exchange TestUpdatePairs error: %s", err)
	}

	// Test force updating to only one product
	exchangeProducts = []string{"btc"}
	err = UAC.UpdatePairs(exchangeProducts, true, true)
	if err != nil {
		t.Errorf("Test Failed - Forced Exchange TestUpdatePairs error: %s", err)
	}

	exchangeProducts = []string{"ltc", "btc", "usd", "aud"}
	// Test updating exchange products for an exchange which doesn't exist
	UAC.Name = "Blah"
	err = UAC.UpdatePairs(exchangeProducts, false, false)
	if err == nil {
		t.Errorf("Test Failed - Exchange UpdatePairs() succeeded on an exchange which doesn't exist")
	}

	// Test updating exchange products
	UAC.Name = "ANX"
	err = UAC.UpdatePairs(exchangeProducts, false, false)
	if err != nil {
		t.Errorf("Test Failed - Exchange UpdatePairs() error: %s", err)
	}

	// Test updating the same new products, diff should be 0
	UAC.Name = "ANX"
	err = UAC.UpdatePairs(exchangeProducts, false, false)
	if err != nil {
		t.Errorf("Test Failed - Exchange UpdatePairs() error: %s", err)
	}

	// Test force updating to only one product
	exchangeProducts = []string{"btc"}
	err = UAC.UpdatePairs(exchangeProducts, false, true)
	if err != nil {
		t.Errorf("Test Failed - Forced Exchange UpdatePairs() error: %s", err)
	}

	// Test update currency pairs with btc excluded
	exchangeProducts = []string{"ltc", "eth"}
	err = UAC.UpdatePairs(exchangeProducts, false, false)
	if err != nil {
		t.Errorf("Test Failed - Forced Exchange UpdatePairs() error: %s", err)
	}

	// Test that empty exchange products should return an error
	exchangeProducts = nil
	err = UAC.UpdatePairs(exchangeProducts, false, false)
	if err == nil {
		t.Errorf("Test failed - empty available pairs should return an error")
	}
}

func TestAPIURL(t *testing.T) {
	testURL := "https://api.something.com"
	testURLSecondary := "https://api.somethingelse.com"
	testURLDefault := "https://api.defaultsomething.com"
	testURLSecondaryDefault := "https://api.defaultsomethingelse.com"

	tester := Base{Name: "test"}

	test := config.ExchangeConfig{}

	err := tester.SetAPIURL(test)
	if err == nil {
		t.Error("test failed - setting zero value config")
	}

	test.API.Endpoints.URL = testURL
	test.API.Endpoints.URLSecondary = testURLSecondary

	tester.API.Endpoints.URLDefault = testURLDefault
	tester.API.Endpoints.URLSecondaryDefault = testURLSecondaryDefault

	err = tester.SetAPIURL(test)
	if err != nil {
		t.Error("test failed", err)
	}

	if tester.GetAPIURL() != testURL {
		t.Error("test failed - incorrect return URL")
	}

	if tester.GetSecondaryAPIURL() != testURLSecondary {
		t.Error("test failed - incorrect return URL")
	}

	if tester.GetAPIURLDefault() != testURLDefault {
		t.Error("test failed - incorrect return URL")
	}

	if tester.GetAPIURLSecondaryDefault() != testURLSecondaryDefault {
		t.Error("test failed - incorrect return URL")
	}
}

func TestSupportsWithdrawPermissions(t *testing.T) {
	UAC := Base{Name: "ANX"}
	UAC.APIWithdrawPermissions = AutoWithdrawCrypto | AutoWithdrawCryptoWithAPIPermission
	withdrawPermissions := UAC.SupportsWithdrawPermissions(AutoWithdrawCrypto)

	if !withdrawPermissions {
		t.Errorf("Expected: %v, Recieved: %v", true, withdrawPermissions)
	}

	withdrawPermissions = UAC.SupportsWithdrawPermissions(AutoWithdrawCrypto | AutoWithdrawCryptoWithAPIPermission)
	if !withdrawPermissions {
		t.Errorf("Expected: %v, Recieved: %v", true, withdrawPermissions)
	}

	withdrawPermissions = UAC.SupportsWithdrawPermissions(AutoWithdrawCrypto | WithdrawCryptoWith2FA)
	if withdrawPermissions {
		t.Errorf("Expected: %v, Recieved: %v", false, withdrawPermissions)
	}

	withdrawPermissions = UAC.SupportsWithdrawPermissions(AutoWithdrawCrypto | AutoWithdrawCryptoWithAPIPermission | WithdrawCryptoWith2FA)
	if withdrawPermissions {
		t.Errorf("Expected: %v, Recieved: %v", false, withdrawPermissions)
	}

	withdrawPermissions = UAC.SupportsWithdrawPermissions(WithdrawCryptoWith2FA)
	if withdrawPermissions {
		t.Errorf("Expected: %v, Recieved: %v", false, withdrawPermissions)
	}
}

func TestFormatWithdrawPermissions(t *testing.T) {
	UAC := Base{Name: "ANX"}
	UAC.APIWithdrawPermissions = AutoWithdrawCrypto |
		AutoWithdrawCryptoWithAPIPermission |
		AutoWithdrawCryptoWithSetup |
		WithdrawCryptoWith2FA |
		WithdrawCryptoWithSMS |
		WithdrawCryptoWithEmail |
		WithdrawCryptoWithWebsiteApproval |
		WithdrawCryptoWithAPIPermission |
		AutoWithdrawFiat |
		AutoWithdrawFiatWithAPIPermission |
		AutoWithdrawFiatWithSetup |
		WithdrawFiatWith2FA |
		WithdrawFiatWithSMS |
		WithdrawFiatWithEmail |
		WithdrawFiatWithWebsiteApproval |
		WithdrawFiatWithAPIPermission |
		WithdrawCryptoViaWebsiteOnly |
		WithdrawFiatViaWebsiteOnly |
		1<<18
	withdrawPermissions := UAC.FormatWithdrawPermissions()
	if withdrawPermissions != "AUTO WITHDRAW CRYPTO & AUTO WITHDRAW CRYPTO WITH API PERMISSION & AUTO WITHDRAW CRYPTO WITH SETUP & WITHDRAW CRYPTO WITH 2FA & WITHDRAW CRYPTO WITH SMS & WITHDRAW CRYPTO WITH EMAIL & WITHDRAW CRYPTO WITH WEBSITE APPROVAL & WITHDRAW CRYPTO WITH API PERMISSION & AUTO WITHDRAW FIAT & AUTO WITHDRAW FIAT WITH API PERMISSION & AUTO WITHDRAW FIAT WITH SETUP & WITHDRAW FIAT WITH 2FA & WITHDRAW FIAT WITH SMS & WITHDRAW FIAT WITH EMAIL & WITHDRAW FIAT WITH WEBSITE APPROVAL & WITHDRAW FIAT WITH API PERMISSION & WITHDRAW CRYPTO VIA WEBSITE ONLY & WITHDRAW FIAT VIA WEBSITE ONLY & UNKNOWN[1<<18]" {
		t.Errorf("Expected: %s, Recieved: %s", AutoWithdrawCryptoText+" & "+AutoWithdrawCryptoWithAPIPermissionText, withdrawPermissions)
	}

	UAC.APIWithdrawPermissions = NoAPIWithdrawalMethods
	withdrawPermissions = UAC.FormatWithdrawPermissions()

	if withdrawPermissions != NoAPIWithdrawalMethodsText {
		t.Errorf("Expected: %s, Recieved: %s", NoAPIWithdrawalMethodsText, withdrawPermissions)
	}
}

func TestOrderTypes(t *testing.T) {
	var ot OrderType = "Mo'Money"

	if ot.ToString() != "Mo'Money" {
		t.Errorf("test failed - unexpected string %s", ot.ToString())
	}

	var os OrderSide = "BUY"

	if os.ToString() != "BUY" {
		t.Errorf("test failed - unexpected string %s", os.ToString())
	}
}
