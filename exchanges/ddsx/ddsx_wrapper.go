package ddsx

import (
	"fmt"
	"sync"
	"time"

	"github.com/thrasher-corp/gocryptotrader/common"
	"github.com/thrasher-corp/gocryptotrader/config"
	"github.com/thrasher-corp/gocryptotrader/currency"
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/account"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/exchanges/kline"
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
	"github.com/thrasher-corp/gocryptotrader/exchanges/orderbook"
	"github.com/thrasher-corp/gocryptotrader/exchanges/protocol"
	"github.com/thrasher-corp/gocryptotrader/exchanges/request"
	"github.com/thrasher-corp/gocryptotrader/exchanges/stream"
	"github.com/thrasher-corp/gocryptotrader/exchanges/ticker"
	"github.com/thrasher-corp/gocryptotrader/exchanges/trade"
	"github.com/thrasher-corp/gocryptotrader/log"
	"github.com/thrasher-corp/gocryptotrader/portfolio/withdraw"
)

// GetDefaultConfig returns a default exchange config
func (dd *DDSX) GetDefaultConfig() (*config.ExchangeConfig, error) {
	dd.SetDefaults()
	exchCfg := new(config.ExchangeConfig)
	exchCfg.Name = dd.Name
	exchCfg.HTTPTimeout = exchange.DefaultHTTPTimeout
	exchCfg.BaseCurrencies = dd.BaseCurrencies

	dd.SetupDefaults(exchCfg)

	if dd.Features.Supports.RESTCapabilities.AutoPairUpdates {
		err := dd.UpdateTradablePairs(true)
		if err != nil {
			return nil, err
		}
	}
	return exchCfg, nil
}

func (dd *DDSX) SetName(name string) {
	dd.Name = name
}

// SetDefaults sets the basic defaults for DDSX
func (dd *DDSX) SetDefaults() {
	dd.Enabled = true
	dd.Verbose = true
	// dd.API.CredentialsValidator.RequiresKey = true
	// dd.API.CredentialsValidator.RequiresSecret = true

	// If using only one pair format for request and configuration, across all
	// supported asset types either SPOT and FUTURES etc. You can use the
	// example below:

	// Request format denotes what the pair as a string will be, when you send
	// a request to an exchange.
	requestFmt := &currency.PairFormat{
		/*Set pair request formatting details here for e.g.*/
		Uppercase: true,
		Delimiter: ":",
	}
	// Config format denotes what the pair as a string will be, when saved to
	// the config.json file.
	configFmt := &currency.PairFormat{
		/*Set pair request formatting details here*/
	}
	err := dd.SetGlobalPairsManager(requestFmt, configFmt /*multiple assets can be set here using the asset package ie asset.Spot*/)
	if err != nil {
		log.Errorln(log.ExchangeSys, err)
	}

	// If assets require multiple differences in formating for request and
	// configuration, another exchange method can be be used e.g. futures
	// contracts require a dash as a delimiter rather than an underscore. You
	// can use this example below:

	spot := currency.PairStore{
		RequestFormat: &currency.PairFormat{Uppercase: true},
		ConfigFormat: &currency.PairFormat{
			Uppercase: true,
			Delimiter: currency.DashDelimiter,
		},
	}

	err = dd.StoreAssetPairFormat(asset.Spot, spot)
	if err != nil {
		log.Errorln(log.ExchangeSys, err)
	}

	// Fill out the capabilities/features that the exchange supports
	dd.Features = exchange.Features{
		Supports: exchange.FeaturesSupported{
			REST:      true,
			Websocket: true,
			RESTCapabilities: protocol.Features{
				TickerFetching:    true,
				OrderbookFetching: true,
			},
			WebsocketCapabilities: protocol.Features{
				TickerFetching:    true,
				OrderbookFetching: true,
			},
			WithdrawPermissions: exchange.AutoWithdrawCrypto |
				exchange.AutoWithdrawFiat,
		},
		Enabled: exchange.FeaturesEnabled{
			AutoPairUpdates: true,
			Kline: kline.ExchangeCapabilitiesEnabled{
				Intervals: map[string]bool{
					kline.FifteenSecond.Word(): true,
					kline.OneMin.Word():        true,
					kline.FiveMin.Word():       true,
					kline.FifteenMin.Word():    true,
					kline.OneHour.Word():       true,
					kline.FourHour.Word():      true,
					kline.OneDay.Word():        true,
				},
				ResultLimit: 5000,
			},
		},
	}
	// NOTE: SET THE EXCHANGES RATE LIMIT HERE
	dd.Requester = request.New(dd.Name,
		common.NewHTTPClientWithTimeout(exchange.DefaultHTTPTimeout))

	// NOTE: SET THE URLs HERE
	dd.API.Endpoints = dd.NewEndpoints()
	dd.API.Endpoints.SetDefaultEndpoints(map[exchange.URL]string{
		exchange.RestSpot: ddsxAPIURL,
		// exchange.WebsocketSpot: ddsxWSAPIURL,
	})
	dd.Websocket = stream.New()
	dd.WebsocketResponseMaxLimit = exchange.DefaultWebsocketResponseMaxLimit
	dd.WebsocketResponseCheckTimeout = exchange.DefaultWebsocketResponseCheckTimeout
	dd.WebsocketOrderbookBufferLimit = exchange.DefaultWebsocketOrderbookBufferLimit
}

// Setup takes in the supplied exchange configuration details and sets params
func (dd *DDSX) Setup(exch *config.ExchangeConfig) error {
	if !exch.Enabled {
		dd.SetEnabled(false)
		return nil
	}

	dd.SetupDefaults(exch)

	/*
		wsRunningEndpoint, err := dd.API.Endpoints.GetURL(exchange.WebsocketSpot)
		if err != nil {
			return err
		}

		// If websocket is supported, please fill out the following

		err = dd.Websocket.Setup(
			&stream.WebsocketSetup{
				Enabled:                          exch.Features.Enabled.Websocket,
				Verbose:                          exch.Verbose,
				AuthenticatedWebsocketAPISupport: exch.API.AuthenticatedWebsocketSupport,
				WebsocketTimeout:                 exch.WebsocketTrafficTimeout,
				DefaultURL:                       ddsxWSAPIURL,
				ExchangeName:                     exch.Name,
				RunningURL:                       wsRunningEndpoint,
				Connector:                        dd.WsConnect,
				Subscriber:                       dd.Subscribe,
				UnSubscriber:                     dd.Unsubscribe,
				Features:                         &dd.Features.Supports.WebsocketCapabilities,
			})
		if err != nil {
			return err
		}

		dd.WebsocketConn = &stream.WebsocketConnection{
			ExchangeName:         dd.Name,
			URL:                  dd.Websocket.GetWebsocketURL(),
			ProxyURL:             dd.Websocket.GetProxyAddress(),
			Verbose:              dd.Verbose,
			ResponseCheckTimeout: exch.WebsocketResponseCheckTimeout,
			ResponseMaxLimit:     exch.WebsocketResponseMaxLimit,
		}

		// NOTE: PLEASE ENSURE YOU SET THE ORDERBOOK BUFFER SETTINGS CORRECTLY
		dd.Websocket.Orderbook.Setup(
			exch.OrderbookConfig.WebsocketBufferLimit,
			true,
			true,
			false,
			false,
			exch.Name)
	*/
	return nil
}

// Start starts the DDSX go routine
func (dd *DDSX) Start(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		dd.Run()
		wg.Done()
	}()
}

// Run implements the Ddsx wrapper
func (dd *DDSX) Run() {
	if dd.Verbose {
		log.Debugf(log.ExchangeSys,
			"%s Websocket: %s.",
			dd.Name,
			common.IsEnabled(dd.Websocket.IsEnabled()))
		dd.PrintEnabledPairs()
	}

	if !dd.GetEnabledFeatures().AutoPairUpdates {
		return
	}

	err := dd.UpdateTradablePairs(false)
	if err != nil {
		log.Errorf(log.ExchangeSys,
			"%s failed to update tradable pairs. Err: %s",
			dd.Name,
			err)
	}
}

// FetchTradablePairs returns a list of the exchanges tradable pairs
func (dd *DDSX) FetchTradablePairs(asset asset.Item) ([]string, error) {
	if !dd.SupportsAsset(asset) {
		return nil, fmt.Errorf("asset type of %s is not supported by %s", asset, dd.Name)
	}

	return []string{"BTC-USDT", "ETH-USDT"}, nil
}

// UpdateTradablePairs updates the exchanges available pairs and stores
// them in the exchanges config
func (dd *DDSX) UpdateTradablePairs(forceUpdate bool) error {
	pairs, err := dd.FetchTradablePairs(asset.Spot)
	if err != nil {
		return err
	}

	p, err := currency.NewPairsFromStrings(pairs)
	if err != nil {
		return err
	}

	return dd.UpdatePairs(p, asset.Spot, false, forceUpdate)
}

// UpdateTicker updates and returns the ticker for a currency pair
func (dd *DDSX) UpdateTicker(p currency.Pair, assetType asset.Item) (*ticker.Price, error) {
	tickerPrice, err := dd.GetTicker(p.String())
	if err != nil {
		return nil, err
	}
	tickerPrice.ExchangeName = dd.Name
	tickerPrice.AssetType = assetType
	err = ticker.ProcessTicker(tickerPrice)
	if err != nil {
		return tickerPrice, err
	}
	return ticker.GetTicker(dd.Name, p, assetType)
}

// FetchTicker returns the ticker for a currency pair
func (dd *DDSX) FetchTicker(p currency.Pair, assetType asset.Item) (*ticker.Price, error) {
	tickerNew, err := ticker.GetTicker(dd.Name, p, assetType)
	if err != nil {
		return dd.UpdateTicker(p, assetType)
	}
	return tickerNew, nil
}

// FetchOrderbook returns orderbook base on the currency pair
func (dd *DDSX) FetchOrderbook(currency currency.Pair, assetType asset.Item) (*orderbook.Base, error) {
	ob, err := orderbook.Get(dd.Name, currency, assetType)
	if err != nil {
		return dd.UpdateOrderbook(currency, assetType)
	}
	return ob, nil
}

// UpdateOrderbook updates and returns the orderbook for a currency pair
func (dd *DDSX) UpdateOrderbook(p currency.Pair, assetType asset.Item) (*orderbook.Base, error) {
	book := &orderbook.Base{
		Exchange:        dd.Name,
		Pair:            p,
		Asset:           assetType,
		VerifyOrderbook: dd.CanVerifyOrderbook,
	}

	ob, err := dd.GetOrderBook(p)
	if err != nil {
		return book, err
	}

	for x := range ob.Bids {
		book.Bids = append(book.Bids, orderbook.Item{
			Amount: ob.Bids[x].Quantity,
			Price:  ob.Bids[x].Price,
		})
	}

	for x := range ob.Asks {
		book.Asks = append(book.Asks, orderbook.Item{
			Amount: ob.Asks[x].Quantity,
			Price:  ob.Asks[x].Price,
		})
	}

	err = book.Process()
	if err != nil {
		return book, err
	}

	return orderbook.Get(dd.Name, p, assetType)
}

// UpdateAccountInfo retrieves balances for all enabled currencies
func (dd *DDSX) UpdateAccountInfo(assetType asset.Item) (account.Holdings, error) {
	return account.Holdings{}, common.ErrNotYetImplemented
}

// FetchAccountInfo retrieves balances for all enabled currencies
func (dd *DDSX) FetchAccountInfo(assetType asset.Item) (account.Holdings, error) {
	return account.Holdings{}, common.ErrNotYetImplemented
}

// GetFundingHistory returns funding history, deposits and
// withdrawals
func (dd *DDSX) GetFundingHistory() ([]exchange.FundHistory, error) {
	return nil, common.ErrNotYetImplemented
}

// GetWithdrawalsHistory returns previous withdrawals data
func (dd *DDSX) GetWithdrawalsHistory(c currency.Code) (resp []exchange.WithdrawalHistory, err error) {
	return nil, common.ErrNotYetImplemented
}

// GetRecentTrades returns the most recent trades for a currency and asset
func (dd *DDSX) GetRecentTrades(p currency.Pair, assetType asset.Item) ([]trade.Data, error) {
	return nil, common.ErrNotYetImplemented
}

// GetHistoricTrades returns historic trade data within the timeframe provided
func (dd *DDSX) GetHistoricTrades(p currency.Pair, assetType asset.Item, timestampStart, timestampEnd time.Time) ([]trade.Data, error) {
	return nil, common.ErrNotYetImplemented
}

// SubmitOrder submits a new order
func (dd *DDSX) SubmitOrder(s *order.Submit) (order.SubmitResponse, error) {
	var submitOrderResponse order.SubmitResponse
	if err := s.Validate(); err != nil {
		return submitOrderResponse, err
	}
	return submitOrderResponse, common.ErrNotYetImplemented
}

// ModifyOrder will allow of changing orderbook placement and limit to
// market conversion
func (dd *DDSX) ModifyOrder(action *order.Modify) (string, error) {
	// if err := action.Validate(); err != nil {
	// 	return "", err
	// }
	return "", common.ErrNotYetImplemented
}

// CancelOrder cancels an order by its corresponding ID number
func (dd *DDSX) CancelOrder(ord *order.Cancel) error {
	// if err := ord.Validate(ord.StandardCancel()); err != nil {
	//	 return err
	// }
	return common.ErrNotYetImplemented
}

// CancelBatchOrders cancels orders by their corresponding ID numbers
func (dd *DDSX) CancelBatchOrders(orders []order.Cancel) (order.CancelBatchResponse, error) {
	return order.CancelBatchResponse{}, common.ErrNotYetImplemented
}

// CancelAllOrders cancels all orders associated with a currency pair
func (dd *DDSX) CancelAllOrders(orderCancellation *order.Cancel) (order.CancelAllResponse, error) {
	// if err := orderCancellation.Validate(); err != nil {
	//	 return err
	// }
	return order.CancelAllResponse{}, common.ErrNotYetImplemented
}

// GetOrderInfo returns order information based on order ID
func (dd *DDSX) GetOrderInfo(orderID string, pair currency.Pair, assetType asset.Item) (order.Detail, error) {
	return order.Detail{}, common.ErrNotYetImplemented
}

// GetDepositAddress returns a deposit address for a specified currency
func (dd *DDSX) GetDepositAddress(cryptocurrency currency.Code, accountID string) (string, error) {
	return "", common.ErrNotYetImplemented
}

// WithdrawCryptocurrencyFunds returns a withdrawal ID when a withdrawal is
// submitted
func (dd *DDSX) WithdrawCryptocurrencyFunds(withdrawRequest *withdraw.Request) (*withdraw.ExchangeResponse, error) {
	// if err := withdrawRequest.Validate(); err != nil {
	//	return nil, err
	// }
	return nil, common.ErrNotYetImplemented
}

// WithdrawFiatFunds returns a withdrawal ID when a withdrawal is
// submitted
func (dd *DDSX) WithdrawFiatFunds(withdrawRequest *withdraw.Request) (*withdraw.ExchangeResponse, error) {
	// if err := withdrawRequest.Validate(); err != nil {
	//	return nil, err
	// }
	return nil, common.ErrNotYetImplemented
}

// WithdrawFiatFundsToInternationalBank returns a withdrawal ID when a withdrawal is
// submitted
func (dd *DDSX) WithdrawFiatFundsToInternationalBank(withdrawRequest *withdraw.Request) (*withdraw.ExchangeResponse, error) {
	// if err := withdrawRequest.Validate(); err != nil {
	//	return nil, err
	// }
	return nil, common.ErrNotYetImplemented
}

// GetActiveOrders retrieves any orders that are active/open
func (dd *DDSX) GetActiveOrders(getOrdersRequest *order.GetOrdersRequest) ([]order.Detail, error) {
	// if err := getOrdersRequest.Validate(); err != nil {
	//	return nil, err
	// }
	return nil, common.ErrNotYetImplemented
}

// GetOrderHistory retrieves account order information
// Can Limit response to specific order status
func (dd *DDSX) GetOrderHistory(getOrdersRequest *order.GetOrdersRequest) ([]order.Detail, error) {
	// if err := getOrdersRequest.Validate(); err != nil {
	//	return nil, err
	// }
	return nil, common.ErrNotYetImplemented
}

// GetFeeByType returns an estimate of fee based on the type of transaction
func (dd *DDSX) GetFeeByType(feeBuilder *exchange.FeeBuilder) (float64, error) {
	return 0, common.ErrNotYetImplemented
}

// ValidateCredentials validates current credentials used for wrapper
func (dd *DDSX) ValidateCredentials(assetType asset.Item) error {
	_, err := dd.UpdateAccountInfo(assetType)
	return dd.CheckTransientError(err)
}

// GetHistoricCandles returns candles between a time period for a set time interval
func (dd *DDSX) GetHistoricCandles(pair currency.Pair, a asset.Item, start, end time.Time, interval kline.Interval) (kline.Item, error) {
	return kline.Item{}, common.ErrNotYetImplemented
}

// GetHistoricCandlesExtended returns candles between a time period for a set time interval
func (dd *DDSX) GetHistoricCandlesExtended(pair currency.Pair, a asset.Item, start, end time.Time, interval kline.Interval) (kline.Item, error) {
	if err := dd.ValidateKline(pair, a, interval); err != nil {
		return kline.Item{}, err
	}

	ret := kline.Item{
		Exchange: dd.Name,
		Pair:     pair,
		Asset:    a,
		Interval: interval,
	}

	dates, err := kline.CalculateCandleDateRanges(start, end, interval, dd.Features.Enabled.Kline.ResultLimit)
	if err != nil {
		return kline.Item{}, err
	}

	formattedPair, err := dd.FormatExchangeCurrency(pair, a)
	if err != nil {
		return kline.Item{}, err
	}

	for x := range dates.Ranges {
		var ohlcData []OHLCVData
		ohlcData, err = dd.GetHistoricalData(formattedPair.String(),
			int64(interval.Duration().Seconds()),
			int64(dd.Features.Enabled.Kline.ResultLimit),
			dates.Ranges[x].Start.Time, dates.Ranges[x].End.Time)
		if err != nil {
			return kline.Item{}, err
		}

		for i := range ohlcData {
			ret.Candles = append(ret.Candles, kline.Candle{
				Time:   ohlcData[i].StartTime,
				Open:   ohlcData[i].Open,
				High:   ohlcData[i].High,
				Low:    ohlcData[i].Low,
				Close:  ohlcData[i].Close,
				Volume: ohlcData[i].Volume,
			})
		}
	}

	dates.SetHasDataFromCandles(ret.Candles)
	summary := dates.DataSummary(false)
	if len(summary) > 0 {
		log.Warnf(log.ExchangeSys, "%v - %v", dd.Name, summary)
	}
	ret.RemoveDuplicates()
	ret.RemoveOutsideRange(start, end)
	ret.SortCandlesByTimestamp(false)
	return ret, nil
}
