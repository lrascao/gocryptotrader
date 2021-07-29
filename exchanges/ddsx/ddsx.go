package ddsx

import (
	"errors"
	"net/url"
	"strconv"
	"time"

	"github.com/thrasher-corp/gocryptotrader/currency"
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/ticker"
)

// Ddsx is the overarching type across this package
type DDSX struct {
	exchange.Base
}

const (
	ddsxAPIURL     = ""
	ddsxAPIVersion = ""

	// Public endpoints

	// Authenticated endpoints
)

var (
	errStartTimeCannotBeAfterEndTime = errors.New("start timestamp cannot be after end timestamp")

	validResolutionData = []int64{15, 60, 300, 900, 3600, 14400, 86400}
)

// GetTickers returns the ticker data for the last 24 hrs
func (dd *DDSX) GetTicker(pair string) (*ticker.Price, error) {
	cp, err := currency.NewPairFromString(pair)
	if err != nil {
		return nil, err
	}
	tickerPrice := &ticker.Price{
		High:  2,
		Low:   1,
		Bid:   0,
		Ask:   0,
		Open:  0,
		Close: 0,
		Pair:  cp,
	}
	return tickerPrice, nil
}

// GetOrderBook returns full orderbook information
//
// OrderBookDataRequestParams contains the following members
// symbol: string of currency pair
// limit: returned limit amount
func (dd *DDSX) GetOrderBook(p currency.Pair) (OrderBook, error) {
	var orderbook OrderBook

	orderbook.Bids = append(orderbook.Bids, OrderbookItem{
		Price:    1.0,
		Quantity: 1,
	})
	orderbook.Asks = append(orderbook.Asks, OrderbookItem{
		Price:    2.0,
		Quantity: 2,
	})

	orderbook.LastUpdateID = 2
	return orderbook, nil
}

// GetHistoricalData gets historical OHLCV data for a given market pair
func (dd *DDSX) GetHistoricalData(marketName string, timeInterval, limit int64, startTime, endTime time.Time) ([]OHLCVData, error) {
	if marketName == "" {
		return nil, errors.New("a market pair must be specified")
	}

	err := checkResolution(timeInterval)
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	params.Set("resolution", strconv.FormatInt(timeInterval, 10))
	if limit != 0 {
		params.Set("limit", strconv.FormatInt(limit, 10))
	}
	if !startTime.IsZero() && !endTime.IsZero() {
		if startTime.After(endTime) {
			return nil, errStartTimeCannotBeAfterEndTime
		}
		params.Set("start_time", strconv.FormatInt(startTime.Unix(), 10))
		params.Set("end_time", strconv.FormatInt(endTime.Unix(), 10))
	}
	// resp := struct {
	// 	Data []OHLCVData `json:"result"`
	// }{}
	// endpoint := common.EncodeURLValues(fmt.Sprintf(getHistoricalData, marketName), params)
	// return resp.Data, f.SendHTTPRequest(exchange.RestSpot, endpoint, &resp)

	var dummy []OHLCVData
	dummy = append(dummy, OHLCVData{
		Close:     1000,
		High:      2000,
		Low:       500,
		Open:      1000,
		StartTime: startTime,
		Time:      startTime.UnixNano(),
		Volume:    1,
	})
	return dummy, nil
}

// Helper functions
func checkResolution(res int64) error {
	for x := range validResolutionData {
		if validResolutionData[x] == res {
			return nil
		}
	}
	return errors.New("resolution data is a mandatory field and the data provided is invalid")
}
