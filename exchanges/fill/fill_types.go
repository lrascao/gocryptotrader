package fill

import (
	"time"

	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
)

type Fills struct {
	exchangeName     string
	dataHandler      chan interface{}
	fillsFeedEnabled bool
}

// Data defines trade data
type Data struct {
	Timestamp     time.Time
	Exchange      string
	AssetType     asset.Item
	CurrencyPair  currency.Pair
	OrderID       string
	ClientOrderID string
	TradeID       string
	Price         float64
	Amount        float64
}
