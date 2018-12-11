package engine

import (
	"errors"
	"fmt"
	"log"

	"github.com/thrasher-/gocryptotrader/common"
	"github.com/thrasher-/gocryptotrader/currency"
	"github.com/thrasher-/gocryptotrader/currency/pair"
	"github.com/thrasher-/gocryptotrader/currency/translation"
	exchange "github.com/thrasher-/gocryptotrader/exchanges"
	"github.com/thrasher-/gocryptotrader/exchanges/assets"
	"github.com/thrasher-/gocryptotrader/exchanges/orderbook"
	"github.com/thrasher-/gocryptotrader/exchanges/stats"
	"github.com/thrasher-/gocryptotrader/exchanges/ticker"
	"github.com/thrasher-/gocryptotrader/portfolio"
)

// GetAllAvailablePairs returns a list of all available pairs on either enabled
// or disabled exchanges
func GetAllAvailablePairs(enabledExchangesOnly bool, assetType assets.AssetType) []pair.CurrencyPair {
	var pairList []pair.CurrencyPair
	for x := range Bot.Config.Exchanges {
		if enabledExchangesOnly && !Bot.Config.Exchanges[x].Enabled {
			continue
		}

		exchName := Bot.Config.Exchanges[x].Name
		pairs, err := Bot.Config.GetAvailablePairs(exchName, assetType)
		if err != nil {
			continue
		}

		for y := range pairs {
			if pair.Contains(pairList, pairs[y], false) {
				continue
			}
			pairList = append(pairList, pairs[y])
		}
	}
	return pairList
}

// GetSpecificAvailablePairs returns a list of supported pairs based on specific
// parameters
func GetSpecificAvailablePairs(enabledExchangesOnly, fiatPairs, includeUSDT, cryptoPairs bool, assetType assets.AssetType) []pair.CurrencyPair {
	var pairList []pair.CurrencyPair
	supportedPairs := GetAllAvailablePairs(enabledExchangesOnly, assetType)

	for x := range supportedPairs {
		if fiatPairs {
			if currency.IsCryptoFiatPair(supportedPairs[x]) &&
				!pair.ContainsCurrency(supportedPairs[x], "USDT") ||
				(includeUSDT && pair.ContainsCurrency(supportedPairs[x], "USDT") && currency.IsCryptoPair(supportedPairs[x])) {
				if pair.Contains(pairList, supportedPairs[x], false) {
					continue
				}
				pairList = append(pairList, supportedPairs[x])
			}
		}
		if cryptoPairs {
			if currency.IsCryptoPair(supportedPairs[x]) {
				if pair.Contains(pairList, supportedPairs[x], false) {
					continue
				}
				pairList = append(pairList, supportedPairs[x])
			}
		}
	}
	return pairList
}

// IsRelatablePairs checks to see if the two pairs are relatable
func IsRelatablePairs(p1, p2 pair.CurrencyPair, includeUSDT bool) bool {
	if p1.Equal(p2, false) {
		return true
	}

	var relatablePairs []pair.CurrencyPair
	relatablePairs = GetRelatableCurrencies(p1, true, includeUSDT)

	if currency.IsCryptoFiatPair(p1) {
		for x := range relatablePairs {
			relatablePairs = append(relatablePairs, GetRelatableFiatCurrencies(relatablePairs[x])...)
		}
	}
	return pair.Contains(relatablePairs, p2, false)
}

// MapCurrenciesByExchange returns a list of currency pairs mapped to an
// exchange
func MapCurrenciesByExchange(p []pair.CurrencyPair, enabledExchangesOnly bool, assetType assets.AssetType) map[string][]pair.CurrencyPair {
	currencyExchange := make(map[string][]pair.CurrencyPair)
	for x := range p {
		for y := range Bot.Config.Exchanges {
			if enabledExchangesOnly && !Bot.Config.Exchanges[y].Enabled {
				continue
			}
			exchName := Bot.Config.Exchanges[y].Name
			success, err := Bot.Config.SupportsPair(exchName, p[x], assetType)
			if err != nil || !success {
				continue
			}

			result, ok := currencyExchange[exchName]
			if !ok {
				var pairs []pair.CurrencyPair
				pairs = append(pairs, p[x])
				currencyExchange[exchName] = pairs
			} else {
				if pair.Contains(result, p[x], false) {
					continue
				}
				result = append(result, p[x])
				currencyExchange[exchName] = result
			}
		}
	}
	return currencyExchange
}

// GetExchangeNamesByCurrency returns a list of exchanges supporting
// a currency pair based on whether the exchange is enabled or not
func GetExchangeNamesByCurrency(p pair.CurrencyPair, enabled bool, assetType assets.AssetType) []string {
	var exchanges []string
	for x := range Bot.Config.Exchanges {
		if enabled != Bot.Config.Exchanges[x].Enabled {
			continue
		}

		exchName := Bot.Config.Exchanges[x].Name
		success, err := Bot.Config.SupportsPair(exchName, p, assetType)
		if err != nil {
			continue
		}

		if success {
			exchanges = append(exchanges, exchName)
		}
	}
	return exchanges
}

// GetRelatableCryptocurrencies returns a list of currency pairs if it can find
// any relatable currencies (e.g ETHBTC -> ETHLTC -> ETHUSDT -> ETHREP)
// incOrig includes the supplied pair if desired
func GetRelatableCryptocurrencies(p pair.CurrencyPair) []pair.CurrencyPair {
	var pairs []pair.CurrencyPair
	cryptocurrencies := currency.CryptoCurrencies

	for x := range cryptocurrencies {
		newPair := pair.NewCurrencyPair(p.FirstCurrency.String(), cryptocurrencies[x])
		if pair.Contains(pairs, newPair, false) {
			continue
		}
		pairs = append(pairs, newPair)
	}
	return pairs
}

// GetRelatableFiatCurrencies returns a list of currency pairs if it can find
// any relatable currencies (e.g ETHUSD -> ETHAUD -> ETHGBP -> ETHJPY)
// incOrig includes the supplied pair if desired
func GetRelatableFiatCurrencies(p pair.CurrencyPair) []pair.CurrencyPair {
	var pairs []pair.CurrencyPair
	fiatCurrencies := currency.FiatCurrencies

	for x := range fiatCurrencies {
		newPair := pair.NewCurrencyPair(p.FirstCurrency.String(), fiatCurrencies[x])
		if pair.Contains(pairs, newPair, false) {
			continue
		}
		pairs = append(pairs, newPair)
	}
	return pairs
}

// GetRelatableCurrencies returns a list of currency pairs if it can find
// any relatable currencies (e.g BTCUSD -> BTC USDT -> XBT USDT -> XBT USD)
// incOrig includes the supplied pair if desired
func GetRelatableCurrencies(p pair.CurrencyPair, incOrig, incUSDT bool) []pair.CurrencyPair {
	var pairs []pair.CurrencyPair

	addPair := func(p pair.CurrencyPair) {
		if pair.Contains(pairs, p, true) {
			return
		}
		pairs = append(pairs, p)
	}

	buildPairs := func(p pair.CurrencyPair, incOrig bool) {
		if incOrig {
			addPair(p)
		}

		first, err := translation.GetTranslation(p.FirstCurrency)
		if err == nil {
			addPair(pair.NewCurrencyPair(first.String(),
				p.SecondCurrency.String()))

			second, err := translation.GetTranslation(p.SecondCurrency)
			if err == nil {
				addPair(pair.NewCurrencyPair(first.String(),
					second.String()))
			}
		}

		second, err := translation.GetTranslation(p.SecondCurrency)
		if err == nil {
			addPair(pair.NewCurrencyPair(p.FirstCurrency.String(),
				second.String()))
		}
	}

	buildPairs(p, incOrig)
	buildPairs(p.Swap(), incOrig)

	if !incUSDT {
		pairs = pair.RemovePairsByFilter(pairs, "USDT")
	}

	return pairs
}

// GetSpecificOrderbook returns a specific orderbook given the currency,
// exchangeName and assetType
func GetSpecificOrderbook(currency, exchangeName string, assetType assets.AssetType) (orderbook.Base, error) {
	var specificOrderbook orderbook.Base
	var err error
	for x := range Bot.Exchanges {
		if Bot.Exchanges[x] != nil {
			if Bot.Exchanges[x].GetName() == exchangeName {
				specificOrderbook, err = Bot.Exchanges[x].FetchOrderbook(
					pair.NewCurrencyPairFromString(currency),
					assetType,
				)
				break
			}
		}
	}
	return specificOrderbook, err
}

// GetSpecificTicker returns a specific ticker given the currency,
// exchangeName and assetType
func GetSpecificTicker(currency, exchangeName string, assetType assets.AssetType) (ticker.Price, error) {
	var specificTicker ticker.Price
	var err error
	for x := range Bot.Exchanges {
		if Bot.Exchanges[x] != nil {
			if Bot.Exchanges[x].GetName() == exchangeName {
				specificTicker, err = Bot.Exchanges[x].FetchTicker(
					pair.NewCurrencyPairFromString(currency),
					assetType,
				)
				break
			}
		}
	}
	return specificTicker, err
}

// GetCollatedExchangeAccountInfoByCoin collates individual exchange account
// information and turns into into a map string of
// exchange.AccountCurrencyInfo
func GetCollatedExchangeAccountInfoByCoin(accounts []exchange.AccountInfo) map[string]exchange.AccountCurrencyInfo {
	result := make(map[string]exchange.AccountCurrencyInfo)
	for i := 0; i < len(accounts); i++ {
		for j := 0; j < len(accounts[i].Currencies); j++ {
			currencyName := accounts[i].Currencies[j].CurrencyName
			avail := accounts[i].Currencies[j].TotalValue
			onHold := accounts[i].Currencies[j].Hold

			info, ok := result[currencyName]
			if !ok {
				accountInfo := exchange.AccountCurrencyInfo{CurrencyName: currencyName, Hold: onHold, TotalValue: avail}
				result[currencyName] = accountInfo
			} else {
				info.Hold += onHold
				info.TotalValue += avail
				result[currencyName] = info
			}
		}
	}
	return result
}

// GetAccountCurrencyInfoByExchangeName returns info for an exchange
func GetAccountCurrencyInfoByExchangeName(accounts []exchange.AccountInfo, exchangeName string) (exchange.AccountInfo, error) {
	for i := 0; i < len(accounts); i++ {
		if accounts[i].ExchangeName == exchangeName {
			return accounts[i], nil
		}
	}
	return exchange.AccountInfo{}, errors.New(exchange.ErrExchangeNotFound)
}

// GetExchangeHighestPriceByCurrencyPair returns the exchange with the highest
// price for a given currency pair and asset type
func GetExchangeHighestPriceByCurrencyPair(p pair.CurrencyPair, assetType assets.AssetType) (string, error) {
	result := stats.SortExchangesByPrice(p, assetType, true)
	if len(result) == 0 {
		return "", fmt.Errorf("no stats for supplied currency pair and asset type")
	}

	return result[0].Exchange, nil
}

// GetExchangeLowestPriceByCurrencyPair returns the exchange with the lowest
// price for a given currency pair and asset type
func GetExchangeLowestPriceByCurrencyPair(p pair.CurrencyPair, assetType assets.AssetType) (string, error) {
	result := stats.SortExchangesByPrice(p, assetType, false)
	if len(result) == 0 {
		return "", fmt.Errorf("no stats for supplied currency pair and asset type")
	}

	return result[0].Exchange, nil
}

// SeedExchangeAccountInfo seeds account info
func SeedExchangeAccountInfo(data []exchange.AccountInfo) {
	if len(data) == 0 {
		return
	}

	port := portfolio.GetPortfolio()

	for i := 0; i < len(data); i++ {
		exchangeName := data[i].ExchangeName
		for j := 0; j < len(data[i].Currencies); j++ {
			currencyName := data[i].Currencies[j].CurrencyName
			onHold := data[i].Currencies[j].Hold
			avail := data[i].Currencies[j].TotalValue
			total := onHold + avail

			if !port.ExchangeAddressExists(exchangeName, currencyName) {
				if total <= 0 {
					continue
				}
				log.Printf("Portfolio: Adding new exchange address: %s, %s, %f, %s\n",
					exchangeName, currencyName, total, portfolio.PortfolioAddressExchange)
				port.Addresses = append(
					port.Addresses,
					portfolio.Address{Address: exchangeName, CoinType: currencyName,
						Balance: total, Description: portfolio.PortfolioAddressExchange},
				)
			} else {
				if total <= 0 {
					log.Printf("Portfolio: Removing %s %s entry.\n", exchangeName,
						currencyName)
					port.RemoveExchangeAddress(exchangeName, currencyName)
				} else {
					balance, ok := port.GetAddressBalance(exchangeName, currencyName, portfolio.PortfolioAddressExchange)
					if !ok {
						continue
					}
					if balance != total {
						log.Printf("Portfolio: Updating %s %s entry with balance %f.\n",
							exchangeName, currencyName, total)
						port.UpdateExchangeAddressBalance(exchangeName, currencyName, total)
					}
				}
			}
		}
	}
}

// GetCryptocurrenciesByExchange returns a list of cryptocurrencies the exchange supports
func GetCryptocurrenciesByExchange(exchangeName string, enabledExchangesOnly, enabledPairs bool, assetType assets.AssetType) ([]string, error) {
	var cryptocurrencies []string
	for x := range Bot.Config.Exchanges {
		if Bot.Config.Exchanges[x].Name == exchangeName {
			if enabledExchangesOnly && !Bot.Config.Exchanges[x].Enabled {
				continue
			}

			exchName := Bot.Config.Exchanges[x].Name
			var pairs []pair.CurrencyPair
			var err error

			if enabledPairs {
				pairs, err = Bot.Config.GetEnabledPairs(exchName, assetType)
				if err != nil {
					return nil, err
				}
			} else {
				pairs, err = Bot.Config.GetAvailablePairs(exchName, assetType)
				if err != nil {
					return nil, err
				}
			}

			for y := range pairs {
				if currency.IsCryptocurrency(pairs[y].FirstCurrency.String()) &&
					!common.StringDataContainsUpper(cryptocurrencies, pairs[y].FirstCurrency.String()) {
					cryptocurrencies = append(cryptocurrencies, pairs[y].FirstCurrency.String())
				}

				if currency.IsCryptocurrency(pairs[y].SecondCurrency.String()) &&
					!common.StringDataContainsUpper(cryptocurrencies, pairs[y].SecondCurrency.String()) {
					cryptocurrencies = append(cryptocurrencies, pairs[y].SecondCurrency.String())
				}
			}
		}
	}
	return cryptocurrencies, nil
}

// GetExchangeCryptocurrencyDepositAddress returns the cryptocurrency deposit address for a particular
// exchange
func GetExchangeCryptocurrencyDepositAddress(exchName string, item pair.CurrencyItem) (string, error) {
	exch := GetExchangeByName(exchName)
	if exch == nil {
		return "", ErrExchangeNotFound
	}

	return exch.GetDepositAddress(item)
}

// GetExchangeCryptocurrencyDepositAddresses obtains an exchanges deposit cryptocurrency list
func GetExchangeCryptocurrencyDepositAddresses() map[string]map[string]string {
	result := make(map[string]map[string]string)

	for x := range Bot.Exchanges {
		if !Bot.Exchanges[x].IsEnabled() {
			continue
		}
		exchName := Bot.Exchanges[x].GetName()

		if !Bot.Exchanges[x].GetAuthenticatedAPISupport() {
			if Bot.Settings.Verbose {
				log.Printf("GetExchangeCryptocurrencyDepositAddresses: Skippping %s due to disabled authenticated API support.", exchName)
			}
			continue
		}

		cryptoCurrencies, err := GetCryptocurrenciesByExchange(exchName, true, true, assets.AssetTypeSpot)
		if err != nil {
			log.Printf("%s failed to get cryptocurrency deposit addresses. Err: %s", exchName, err)
			continue
		}

		cryptoAddr := make(map[string]string)
		for y := range cryptoCurrencies {
			cryptocurrency := cryptoCurrencies[y]
			depositAddr, err := Bot.Exchanges[x].GetDepositAddress(pair.CurrencyItem(cryptocurrency))
			if err != nil {
				log.Printf("%s failed to get cryptocurrency deposit addresses. Err: %s", exchName, err)
				continue
			}
			cryptoAddr[cryptocurrency] = depositAddr
		}
		result[exchName] = cryptoAddr
	}

	return result
}

// GetDepositAddressByExchange returns a deposit address for the specified exchange and cryptocurrency
// if it exists
func GetDepositAddressByExchange(exchName string, currencyItem pair.CurrencyItem) string {
	for x, y := range Bot.CryptocurrencyDepositAddresses {
		if exchName == x {
			addr, ok := y[currencyItem.String()]
			if ok {
				return addr
			}
		}
	}
	return ""
}

// GetOrderByExchange returns the order info for the desired exchange
func GetOrderByExchange(exchName string, orderID int64) (exchange.OrderDetail, error) {
	exch := GetExchangeByName(exchName)
	if exch == nil {
		return exchange.OrderDetail{}, ErrExchangeNotFound
	}

	return exch.GetOrderInfo(orderID)
}

// GetDepositAddressesByExchange returns a list of cryptocurrency addresses for the specified
// exchange if they exist
func GetDepositAddressesByExchange(exchName string) map[string]string {
	for x, y := range Bot.CryptocurrencyDepositAddresses {
		if exchName == x {
			return y
		}
	}
	return nil
}

// CancelAllOrdersByExchange sends a cancel all orders request to the desired exchange
func CancelAllOrdersByExchange(exchName string) error {
	exch := GetExchangeByName(exchName)
	if exch == nil {
		return ErrExchangeNotFound
	}

	return exch.CancelAllOrders()
}

// CancelOrderByExchange sends a cancel order request based on the desired exchange
// and order ID
func CancelOrderByExchange(exchName string, order exchange.OrderCancellation) error {
	exch := GetExchangeByName(exchName)
	if exch == nil {
		return ErrExchangeNotFound
	}

	return exch.CancelOrder(order)
}

// WithdrawCryptocurrencyFundsByExchange withdraws the desired cryptocurrency and amount to a desired cryptocurrency address
func WithdrawCryptocurrencyFundsByExchange(exchName string, cryptocurrency pair.CurrencyItem, address string, amount float64) (string, error) {
	exch := GetExchangeByName(exchName)
	if exch == nil {
		return "", ErrExchangeNotFound
	}

	return exch.WithdrawCryptocurrencyFunds(address, cryptocurrency, amount)
}
