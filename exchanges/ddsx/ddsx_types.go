package ddsx

import "time"

// OHLCVData stores historical OHLCV data
type OHLCVData struct {
	Close     float64   `json:"close"`
	High      float64   `json:"high"`
	Low       float64   `json:"low"`
	Open      float64   `json:"open"`
	StartTime time.Time `json:"startTime"`
	Time      int64     `json:"time"`
	Volume    float64   `json:"volume"`
}

// OrderbookItem stores an individual orderbook item
type OrderbookItem struct {
	Price    float64
	Quantity float64
}

// OrderBook actual structured data that can be used for orderbook
type OrderBook struct {
	Symbol       string
	LastUpdateID int64
	Code         int
	Msg          string
	Bids         []OrderbookItem
	Asks         []OrderbookItem
}
