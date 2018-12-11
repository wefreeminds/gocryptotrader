package events

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/thrasher-/gocryptotrader/common"
	"github.com/thrasher-/gocryptotrader/communications"
	"github.com/thrasher-/gocryptotrader/communications/base"
	"github.com/thrasher-/gocryptotrader/config"
	"github.com/thrasher-/gocryptotrader/currency/pair"
	"github.com/thrasher-/gocryptotrader/exchanges/assets"
	"github.com/thrasher-/gocryptotrader/exchanges/orderbook"
	"github.com/thrasher-/gocryptotrader/exchanges/ticker"
)

// Event const vars
const (
	ItemPrice     = "PRICE"
	ItemOrderbook = "ORDERBOOK"

	ConditionGreaterThan        = ">"
	ConditionGreaterThanOrEqual = ">="
	ConditionLessThan           = "<"
	ConditionLessThanOrEqual    = "<="
	ConditionIsEqual            = "=="

	ActionSMSNotify    = "SMS"
	ActionConsolePrint = "CONSOLE_PRINT"
	ActionTest         = "ACTION_TEST"

	defaultSleepDelay = time.Duration(time.Millisecond * 500)
	defaultVerbose    = true
)

// vars related to events package
var (
	errInvalidItem      = errors.New("invalid item")
	errInvalidCondition = errors.New("invalid conditional option")
	errInvalidAction    = errors.New("invalid action")
	errExchangeDisabled = errors.New("desired exchange is disabled")

	SleepDelay = defaultSleepDelay
	Verbose    = defaultVerbose

	// NOTE comms is an interim implementation
	comms *communications.Communications
)

// ConditionParams holds the event condition variables
type ConditionParams struct {
	Condition string
	Price     float64

	CheckBids        bool
	CheckBidsAndAsks bool
	OrderbookAmount  float64
}

// Event struct holds the event variables
type Event struct {
	ID        int
	Exchange  string
	Item      string
	Condition ConditionParams
	Pair      pair.CurrencyPair
	Asset     assets.AssetType
	Action    string
	Executed  bool
}

// Events variable is a pointer array to the event structures that will be
// appended
var Events []*Event

// SetComms is an interim function that will support a median integration. This
// sets the current comms package.
func SetComms(commsP *communications.Communications) {
	comms = commsP
}

// Add adds an event to the Events chain and returns an index/eventID
// and an error
func Add(Exchange, Item string, Condition ConditionParams, CurrencyPair pair.CurrencyPair, Asset assets.AssetType, Action string) (int, error) {
	err := IsValidEvent(Exchange, Item, Condition, Action)
	if err != nil {
		return 0, err
	}

	Event := &Event{}

	if len(Events) == 0 {
		Event.ID = 0
	} else {
		Event.ID = len(Events) + 1
	}

	Event.Exchange = Exchange
	Event.Item = Item
	Event.Condition = Condition
	Event.Pair = CurrencyPair
	Event.Asset = Asset
	Event.Action = Action
	Event.Executed = false
	Events = append(Events, Event)
	return Event.ID, nil
}

// Remove deletes and event by its ID
func Remove(EventID int) bool {
	for i, x := range Events {
		if x.ID == EventID {
			Events = append(Events[:i], Events[i+1:]...)
			return true
		}
	}
	return false
}

// GetEventCounter displays the emount of total events on the chain and the
// events that have been executed.
func GetEventCounter() (int, int) {
	total := len(Events)
	executed := 0

	for _, x := range Events {
		if x.Executed {
			executed++
		}
	}
	return total, executed
}

// ExecuteAction will execute the action pending on the chain
func (e *Event) ExecuteAction() bool {
	if common.StringContains(e.Action, ",") {
		action := common.SplitStrings(e.Action, ",")
		if action[0] == ActionSMSNotify {
			message := fmt.Sprintf("Event triggered: %s", e.String())
			if action[1] == "ALL" {
				comms.PushEvent(base.Event{TradeDetails: message})
			}
		}
	} else {
		log.Printf("Event triggered: %s", e.String())
	}
	return true
}

// String turns the structure event into a string
func (e *Event) String() string {
	return fmt.Sprintf(
		"If the %s%s [%s] %s on %s meets the following %v then %s.", e.Pair.FirstCurrency.String(),
		e.Pair.SecondCurrency.String(), e.Asset, e.Item, e.Exchange, e.Condition, e.Action,
	)
}

func (e *Event) processTicker() bool {
	targetPrice := e.Condition.Price

	t, err := ticker.GetTicker(e.Exchange, e.Pair, e.Asset)
	if err != nil {
		if Verbose {
			log.Printf("Events: failed to get ticker. Err: %s", err)
		}
		return false
	}

	lastPrice := t.Last

	if lastPrice == 0 {
		if Verbose {
			log.Printf("Events: ticker last price is 0")
		}
		return false
	}

	return e.processCondition(lastPrice, targetPrice)
}

func (e *Event) processCondition(actual, threshold float64) bool {
	switch e.Condition.Condition {
	case ConditionGreaterThan:
		{
			if actual > threshold {
				return e.ExecuteAction()
			}
		}
	case ConditionGreaterThanOrEqual:
		{
			if actual >= threshold {
				return e.ExecuteAction()
			}
		}
	case ConditionLessThan:
		{
			if actual < threshold {
				return e.ExecuteAction()
			}
		}
	case ConditionLessThanOrEqual:
		{
			if actual <= threshold {
				return e.ExecuteAction()
			}
		}
	case ConditionIsEqual:
		{
			if actual == threshold {
				return e.ExecuteAction()
			}
		}
	}
	return false
}

func (e *Event) processOrderbook() bool {
	ob, err := orderbook.GetOrderbook(e.Exchange, e.Pair, e.Asset)
	if err != nil {
		if Verbose {
			log.Printf("Events: Failed to get orderbook. Err: %s", err)
		}
		return false
	}

	success := false
	if e.Condition.CheckBids || e.Condition.CheckBidsAndAsks {
		for x := range ob.Bids {
			subtotal := ob.Bids[x].Amount * ob.Bids[x].Price
			result := e.processCondition(subtotal, e.Condition.OrderbookAmount)
			if result {
				success = true
				log.Printf("Events: Bid Amount: %f Price: %v Subtotal: %v", ob.Bids[x].Amount, ob.Bids[x].Price, subtotal)
			}
		}
	}

	if !e.Condition.CheckBids || e.Condition.CheckBidsAndAsks {
		for x := range ob.Asks {
			subtotal := ob.Asks[x].Amount * ob.Asks[x].Price
			result := e.processCondition(subtotal, e.Condition.OrderbookAmount)
			if result {
				success = true
				log.Printf("Events: Ask Amount: %f Price: %v Subtotal: %v", ob.Asks[x].Amount, ob.Asks[x].Price, subtotal)
			}
		}
	}
	return success
}

// CheckCondition will check the event structure to see if there is a condition
// met
func (e *Event) CheckCondition() bool {
	if e.Item == ItemPrice {
		return e.processTicker()
	}

	return e.processOrderbook()
}

// IsValidEvent checks the actions to be taken and returns an error if incorrect
func IsValidEvent(Exchange, Item string, Condition ConditionParams, Action string) error {
	Exchange = common.StringToUpper(Exchange)
	Item = common.StringToUpper(Item)
	Action = common.StringToUpper(Action)

	if !IsValidExchange(Exchange) {
		return errExchangeDisabled
	}

	if !IsValidItem(Item) {
		return errInvalidItem
	}

	if !IsValidCondition(Condition.Condition) {
		return errInvalidCondition
	}

	if Item == ItemPrice {
		if Condition.Price == 0 {
			return errInvalidCondition
		}
	}

	if Item == ItemOrderbook {
		if Condition.OrderbookAmount == 0 {
			return errInvalidAction
		}
	}

	if common.StringContains(Action, ",") {
		action := common.SplitStrings(Action, ",")

		if action[0] != ActionSMSNotify {
			return errInvalidAction
		}

		if action[1] != "ALL" {
			comms.PushEvent(base.Event{Type: action[1]})
		}
	} else {
		if Action != ActionConsolePrint && Action != ActionTest {
			return errInvalidAction
		}
	}
	return nil
}

// EventManger is the overarching routine that will iterate through the Events
// chain
func EventManger() {
	log.Printf("EventManager started. SleepDelay: %v", SleepDelay.String())

	for {
		total, executed := GetEventCounter()
		if total > 0 && executed != total {
			for _, event := range Events {
				if !event.Executed {
					if Verbose {
						log.Printf("Events: Processing event %s.", event.String())
					}
					success := event.CheckCondition()
					if success {
						log.Printf(
							"Events: ID: %d triggered on %s successfully.\n", event.ID,
							event.Exchange,
						)
						event.Executed = true
					}
				}
			}
		}
		time.Sleep(SleepDelay)
	}
}

// IsValidExchange validates the exchange
func IsValidExchange(Exchange string) bool {
	Exchange = common.StringToLower(Exchange)
	cfg := config.GetConfig()
	for _, x := range cfg.Exchanges {
		if common.StringToLower(x.Name) == Exchange && x.Enabled {
			return true
		}
	}
	return false
}

// IsValidCondition validates passed in condition
func IsValidCondition(Condition string) bool {
	switch Condition {
	case ConditionGreaterThan, ConditionGreaterThanOrEqual, ConditionLessThan, ConditionLessThanOrEqual, ConditionIsEqual:
		return true
	}
	return false
}

// IsValidAction validates passed in action
func IsValidAction(Action string) bool {
	Action = common.StringToUpper(Action)
	switch Action {
	case ActionSMSNotify, ActionConsolePrint, ActionTest:
		return true
	}
	return false
}

// IsValidItem validates passed in Item
func IsValidItem(Item string) bool {
	Item = common.StringToUpper(Item)
	switch Item {
	case ItemPrice, ItemOrderbook:
		return true
	}
	return false
}
