// Package market provides market data functionality for the Avanza API.
package market

import "encoding/json"

// OrderDepthLevel contains bid/ask prices and volumes at a single price level.
type OrderDepthLevel struct {
	BuyPrice   float64 `json:"buyPrice"`
	BuyVolume  float64 `json:"buyVolume"`
	SellPrice  float64 `json:"sellPrice"`
	SellVolume float64 `json:"sellVolume"`
}

// OrderDepthData contains the complete order book snapshot.
type OrderDepthData struct {
	OrderbookID           string            `json:"orderbookId"`
	Levels                []OrderDepthLevel `json:"levels"`
	MarketMakerLevelInAsk int               `json:"marketMakerLevelInAsk"`
	MarketMakerLevelInBid int               `json:"marketMakerLevelInBid"`
}

// OrderDepthEvent is a single event from the order depth subscription stream.
type OrderDepthEvent struct {
	Event string         `json:"event"`
	Data  OrderDepthData `json:"data"`
	ID    string         `json:"id"`
	Retry int            `json:"retry"`
}

// SearchRequest configures an instrument search.
type SearchRequest struct {
	// Query is the search string (required).
	Query string

	// Types filters results by instrument type (e.g. "STOCK", "FUND", "CERTIFICATE",
	// "EXCHANGE_TRADED_FUND", "WARRANT", "OPTION", "INDEX"). Empty means all types.
	Types []string

	// From is the pagination offset. Default 0.
	From int

	// Size is the number of results per page. Default 30.
	Size int
}

// searchAPIRequest is the internal request body sent to the search endpoint.
type searchAPIRequest struct {
	Query           string              `json:"query"`
	SearchFilter    searchFilter        `json:"searchFilter"`
	ScreenSize      string              `json:"screenSize"`
	OriginPath      string              `json:"originPath"`
	OriginPlatform  string              `json:"originPlatform"`
	SearchSessionID string              `json:"searchSessionId"`
	Pagination      searchAPIPagination `json:"pagination"`
}

type searchFilter struct {
	Types []string `json:"types"`
}

type searchAPIPagination struct {
	From int `json:"from"`
	Size int `json:"size"`
}

// SearchResponse contains the search results.
type SearchResponse struct {
	TotalNumberOfHits int              `json:"totalNumberOfHits"`
	Hits              []SearchHit      `json:"hits"`
	SearchQuery       string           `json:"searchQuery"`
	Facets            SearchFacets     `json:"facets"`
	Pagination        SearchPagination `json:"pagination"`
}

// SearchHit is a single search result.
type SearchHit struct {
	Type            string         `json:"type"`
	Title           string         `json:"title"`
	Description     string         `json:"description"`
	FlagCode        string         `json:"flagCode"`
	OrderbookID     string         `json:"orderBookId"`
	URLSlugName     string         `json:"urlSlugName"`
	Tradable        bool           `json:"tradeable"`
	Sellable        bool           `json:"sellable"`
	Buyable         bool           `json:"buyable"`
	Price           SearchHitPrice `json:"price"`
	StockSectors    []StockSector  `json:"stockSectors"`
	FundTags        []FundTag      `json:"fundTags"`
	MarketPlaceName string         `json:"marketPlaceName"`
	SubType         *string        `json:"subType"`
}

// SearchHitPrice contains price and change data for a search hit.
type SearchHitPrice struct {
	Last                          string  `json:"last"`
	Currency                      string  `json:"currency"`
	TodayChangePercent            string  `json:"todayChangePercent"`
	TodayChangeValue              string  `json:"todayChangeValue"`
	TodayChangeDirection          int     `json:"todayChangeDirection"`
	ThreeMonthsAgoChangePercent   *string `json:"threeMonthsAgoChangePercent"`
	ThreeMonthsAgoChangeDirection int     `json:"threeMonthsAgoChangeDirection"`
	Spread                        *string `json:"spread"`
}

// StockSector categorizes a stock by industry sector.
type StockSector struct {
	ID          int    `json:"id"`
	Level       int    `json:"level"`
	Name        string `json:"name"`
	EnglishName string `json:"englishName"`
}

// FundTag describes a fund classification.
type FundTag struct {
	Title       string `json:"title"`
	Category    string `json:"category"`
	TagCategory string `json:"tagCategory"`
}

// SearchFacets contains result counts grouped by instrument type.
type SearchFacets struct {
	Types []TypeFacet `json:"types"`
}

// TypeFacet is the hit count for a single instrument type.
type TypeFacet struct {
	Type  string `json:"type"`
	Count int    `json:"count"`
}

// SearchPagination describes the pagination state of search results.
type SearchPagination struct {
	Size int `json:"size"`
	From int `json:"from"`
}

// --- Market guide types ---

// Listing describes the market listing details of an instrument.
type Listing struct {
	ShortName             string `json:"shortName"`
	TickerSymbol          string `json:"tickerSymbol"`
	CountryCode           string `json:"countryCode"`
	Currency              string `json:"currency"`
	MarketPlaceCode       string `json:"marketPlaceCode"`
	MarketPlaceName       string `json:"marketPlaceName"`
	MarketListName        string `json:"marketListName,omitempty"`
	TickSizeListID        string `json:"tickSizeListId"`
	MarketTradesAvailable bool   `json:"marketTradesAvailable"`
}

// Quote contains real-time price and trading data.
type Quote struct {
	Buy                        float64 `json:"buy"`
	Sell                       float64 `json:"sell"`
	Last                       float64 `json:"last"`
	Highest                    float64 `json:"highest"`
	Lowest                     float64 `json:"lowest"`
	Change                     float64 `json:"change"`
	ChangePercent              float64 `json:"changePercent"`
	Spread                     float64 `json:"spread"`
	TimeOfLast                 int64   `json:"timeOfLast"`
	TotalValueTraded           float64 `json:"totalValueTraded"`
	TotalVolumeTraded          float64 `json:"totalVolumeTraded"`
	Updated                    int64   `json:"updated"`
	VolumeWeightedAveragePrice float64 `json:"volumeWeightedAveragePrice"`
	IsRealTime                 bool    `json:"isRealTime"`
}

// HistoricalClosingPrices contains closing prices across various time periods.
// Fields are pointers because available periods depend on the instrument's age.
type HistoricalClosingPrices struct {
	OneDay      *float64 `json:"oneDay,omitempty"`
	OneWeek     *float64 `json:"oneWeek,omitempty"`
	OneMonth    *float64 `json:"oneMonth,omitempty"`
	ThreeMonths *float64 `json:"threeMonths,omitempty"`
	SixMonths   *float64 `json:"sixMonths,omitempty"`
	StartOfYear *float64 `json:"startOfYear,omitempty"`
	OneYear     *float64 `json:"oneYear,omitempty"`
	ThreeYears  *float64 `json:"threeYears,omitempty"`
	FiveYears   *float64 `json:"fiveYears,omitempty"`
	TenYears    *float64 `json:"tenYears,omitempty"`
	Start       *float64 `json:"start,omitempty"`
	StartDate   string   `json:"startDate"`
}

// Underlying describes the underlying instrument for derivatives.
type Underlying struct {
	OrderbookID          string  `json:"orderbookId"`
	Name                 string  `json:"name"`
	InstrumentType       string  `json:"instrumentType"`
	InstrumentSubType    string  `json:"instrumentSubType"`
	Quote                Quote   `json:"quote"`
	Listing              Listing `json:"listing"`
	PreviousClosingPrice float64 `json:"previousClosingPrice"`
	Reference            bool    `json:"reference"`
}

// --- Stock ---

// Stock contains detailed market data for a stock.
type Stock struct {
	OrderbookID             string                  `json:"orderbookId"`
	Name                    string                  `json:"name"`
	ISIN                    string                  `json:"isin"`
	InstrumentID            string                  `json:"instrumentId"`
	Sectors                 []StockSectorInfo       `json:"sectors"`
	Tradable                string                  `json:"tradable"`
	Listing                 Listing                 `json:"listing"`
	MarketPlace             MarketPlace             `json:"marketPlace"`
	HistoricalClosingPrices HistoricalClosingPrices `json:"historicalClosingPrices"`
	KeyIndicators           StockKeyIndicators      `json:"keyIndicators"`
	Quote                   Quote                   `json:"quote"`
	Type                    string                  `json:"type"`
}

// StockSectorInfo categorizes a stock by sector in market guide responses.
type StockSectorInfo struct {
	SectorID   string `json:"sectorId"`
	SectorName string `json:"sectorName"`
}

// MarketPlace describes the trading venue and hours.
type MarketPlace struct {
	MarketOpen        bool   `json:"marketOpen"`
	TimeLeftMs        int64  `json:"timeLeftMs"`
	OpeningTime       string `json:"openingTime"`
	TodayClosingTime  string `json:"todayClosingTime"`
	NormalClosingTime string `json:"normalClosingTime"`
}

// StockKeyIndicators contains financial metrics for a stock.
type StockKeyIndicators struct {
	NumberOfOwners          int           `json:"numberOfOwners"`
	ShortSellingRatio       float64       `json:"shortSellingRatio"`
	ReportDate              string        `json:"reportDate"`
	DirectYield             float64       `json:"directYield"`
	OrdinaryDirectYield     float64       `json:"ordinaryDirectYield"`
	TotalDirectYield        float64       `json:"totalDirectYield"`
	Volatility              float64       `json:"volatility"`
	Beta                    float64       `json:"beta"`
	PriceEarningsRatio      float64       `json:"priceEarningsRatio"`
	PriceBookRatio          float64       `json:"priceBookRatio"`
	EVEBITRatio             float64       `json:"evEbitRatio"`
	InterestCoverageRatio   float64       `json:"interestCoverageRatio"`
	ReturnOnEquity          float64       `json:"returnOnEquity"`
	ReturnOnTotalAssets     float64       `json:"returnOnTotalAssets"`
	ReturnOnCapitalEmployed float64       `json:"returnOnCapitalEmployed"`
	EquityRatio             float64       `json:"equityRatio"`
	CapitalTurnover         float64       `json:"capitalTurnover"`
	MarketCapital           MonetaryValue `json:"marketCapital"`
	EquityPerShare          MonetaryValue `json:"equityPerShare"`
	TurnoverPerShare        MonetaryValue `json:"turnoverPerShare"`
	EarningsPerShare        MonetaryValue `json:"earningsPerShare"`
	OperatingCashFlow       MonetaryValue `json:"operatingCashFlow"`
	Dividend                *Dividend     `json:"dividend,omitempty"`
	DividendsPerYear        int           `json:"dividendsPerYear"`
	NextReport              *Report       `json:"nextReport,omitempty"`
	PreviousReport          *Report       `json:"previousReport,omitempty"`
}

// MonetaryValue is a value with its currency.
type MonetaryValue struct {
	Value    float64 `json:"value"`
	Currency string  `json:"currency"`
}

// Dividend describes an upcoming or past dividend payment.
type Dividend struct {
	ExDate       string  `json:"exDate"`
	PaymentDate  string  `json:"paymentDate"`
	Amount       float64 `json:"amount"`
	CurrencyCode string  `json:"currencyCode"`
	ExDateStatus string  `json:"exDateStatus"`
}

// Report describes a financial report date.
type Report struct {
	Date       string `json:"date"`
	ReportType string `json:"reportType"`
}

// --- Certificate ---

// Certificate contains detailed market data for a certificate.
type Certificate struct {
	OrderbookID             string                   `json:"orderbookId"`
	Name                    string                   `json:"name"`
	ISIN                    string                   `json:"isin"`
	Tradable                string                   `json:"tradable"`
	Listing                 Listing                  `json:"listing"`
	HistoricalClosingPrices HistoricalClosingPrices  `json:"historicalClosingPrices"`
	KeyIndicators           CertificateKeyIndicators `json:"keyIndicators"`
	Quote                   Quote                    `json:"quote"`
	Type                    string                   `json:"type"`
	Underlying              Underlying               `json:"underlying"`
	AssetCategory           string                   `json:"assetCategory"`
	Category                string                   `json:"category"`
	SubCategory             string                   `json:"subCategory"`
}

// CertificateKeyIndicators contains key metrics for a certificate.
type CertificateKeyIndicators struct {
	Leverage       float64 `json:"leverage"`
	IsAza          bool    `json:"isAza"`
	ProductLink    string  `json:"productLink"`
	NumberOfOwners int     `json:"numberOfOwners"`
}

// --- Warrant ---

// Warrant contains detailed market data for a warrant.
type Warrant struct {
	OrderbookID             string                  `json:"orderbookId"`
	Name                    string                  `json:"name"`
	ISIN                    string                  `json:"isin"`
	Tradable                string                  `json:"tradable"`
	Listing                 Listing                 `json:"listing"`
	HistoricalClosingPrices HistoricalClosingPrices `json:"historicalClosingPrices"`
	KeyIndicators           WarrantKeyIndicators    `json:"keyIndicators"`
	Quote                   Quote                   `json:"quote"`
	Type                    string                  `json:"type"`
	Underlying              Underlying              `json:"underlying"`
	AssetCategory           string                  `json:"assetCategory"`
	Category                string                  `json:"category"`
	SubCategory             string                  `json:"subCategory"`
}

// WarrantKeyIndicators contains key metrics for a warrant.
type WarrantKeyIndicators struct {
	Parity         float64 `json:"parity"`
	BarrierLevel   float64 `json:"barrierLevel"`
	FinancingLevel float64 `json:"financingLevel"`
	Direction      string  `json:"direction"`
	Leverage       float64 `json:"leverage"`
	IsAza          bool    `json:"isAza"`
	NumberOfOwners int     `json:"numberOfOwners"`
	SubType        string  `json:"subType"`
}

// --- Orderbook (trading-critical) ---

// Orderbook contains trading parameters for an instrument (tick sizes, feature support, validity dates).
type Orderbook struct {
	ID                    string         `json:"id"`
	Name                  string         `json:"name"`
	ISIN                  string         `json:"isin"`
	InstrumentID          string         `json:"instrumentId"`
	MarketPlace           string         `json:"marketPlace"`
	CountryCode           string         `json:"countryCode"`
	TickSizeList          TickSizeList   `json:"tickSizeList"`
	CollateralValue       float64        `json:"collateralValue"`
	Currency              string         `json:"currency"`
	OrderbookStatus       string         `json:"orderbookStatus"`
	MinValidUntil         string         `json:"minValidUntil"`
	MaxValidUntil         string         `json:"maxValidUntil"`
	InstrumentType        string         `json:"instrumentType"`
	VolumeFactor          int            `json:"volumeFactor"`
	FeatureSupport        FeatureSupport `json:"featureSupport"`
	PriceType             string         `json:"priceType"`
	TradingUnit           int            `json:"tradingUnit"`
	TickerSymbol          string         `json:"tickerSymbol"`
	UnderlyingOrderbook   string         `json:"underlyingOrderbook,omitempty"`
	UnderlyingCountryCode string         `json:"underlyingCountryCode,omitempty"`
}

// TickSizeList contains the tick size table for an instrument.
type TickSizeList struct {
	TickSizeEntries []TickSizeEntry `json:"tickSizeEntries"`
}

// TickSizeEntry defines the tick size for a price range.
type TickSizeEntry struct {
	Min  float64 `json:"min"`
	Max  float64 `json:"max"`
	Tick float64 `json:"tick"`
}

// FeatureSupport describes which trading features are available for an instrument.
type FeatureSupport struct {
	StopLoss                 bool `json:"stopLoss"`
	FillAndOrKill            bool `json:"fillAndOrKill"`
	OpenVolume               bool `json:"openVolume"`
	MarketTrades             bool `json:"marketTrades"`
	MarketTradesSummary      bool `json:"marketTradesSummary"`
	NordicAtMid              bool `json:"nordicAtMid"`
	StopLossMarketMakerQuote bool `json:"stopLossMarketMakerQuote"`
	RoutingStrategies        bool `json:"routingStrategies"`
	LimitOnClose             bool `json:"limitOnClose"`
}

// --- Market data (trading-critical) ---

// MarketData contains real-time quote, order depth, and trades for an instrument.
type MarketData struct {
	Quote      MarketDataQuote      `json:"quote"`
	OrderDepth MarketDataOrderDepth `json:"orderDepth"`
	Trades     []json.RawMessage    `json:"trades"`
}

// MarketDataQuote contains real-time price data from the trading-critical endpoint.
type MarketDataQuote struct {
	Buy                        float64 `json:"buy"`
	Sell                       float64 `json:"sell"`
	Last                       float64 `json:"last"`
	Highest                    float64 `json:"highest"`
	Lowest                     float64 `json:"lowest"`
	Change                     float64 `json:"change"`
	ChangePercent              float64 `json:"changePercent"`
	TimeOfLast                 string  `json:"timeOfLast"`
	TotalValueTraded           float64 `json:"totalValueTraded"`
	TotalVolumeTraded          int     `json:"totalVolumeTraded"`
	Updated                    string  `json:"updated"`
	VolumeWeightedAveragePrice float64 `json:"volumeWeightedAveragePrice"`
}

// MarketDataOrderDepth contains the order book from the trading-critical endpoint.
type MarketDataOrderDepth struct {
	ReceivedTime        int64                       `json:"receivedTime"`
	Levels              []MarketDataOrderDepthLevel `json:"levels"`
	MarketMakerExpected bool                        `json:"marketMakerExpected"`
}

// MarketDataOrderDepthLevel contains bid and ask at a single price level.
type MarketDataOrderDepthLevel struct {
	BuySide  MarketDataOrderSide `json:"buySide"`
	SellSide MarketDataOrderSide `json:"sellSide"`
}

// MarketDataOrderSide contains the price and volume for one side of the order book.
type MarketDataOrderSide struct {
	Price       float64 `json:"price"`
	Volume      int     `json:"volume"`
	PriceString string  `json:"priceString"`
}

// --- Price chart ---

// TimePeriod is the time window for a price chart request.
type TimePeriod string

// Time periods accepted by the price chart endpoint.
const (
	TimePeriodToday       TimePeriod = "today"
	TimePeriodOneWeek     TimePeriod = "one_week"
	TimePeriodOneMonth    TimePeriod = "one_month"
	TimePeriodThreeMonths TimePeriod = "three_months"
	TimePeriodSixMonths   TimePeriod = "six_months"
	TimePeriodThisYear    TimePeriod = "this_year"
	TimePeriodOneYear     TimePeriod = "one_year"
	TimePeriodInfinity    TimePeriod = "infinity"
)

// MarketMakerPriceChart contains OHLC bars and market-maker quotes for a single instrument.
type MarketMakerPriceChart struct {
	OHLC        []OHLCBar          `json:"ohlc"`
	MarketMaker []MarketMakerQuote `json:"marketMaker"`
	From        string             `json:"from"`
	To          string             `json:"to"`
	Metadata    PriceChartMetadata `json:"metadata"`
}

// OHLCBar is a single open/high/low/close bar with traded volume.
// Timestamp is Unix epoch milliseconds.
type OHLCBar struct {
	Timestamp         int64   `json:"timestamp"`
	Open              float64 `json:"open"`
	Close             float64 `json:"close"`
	Low               float64 `json:"low"`
	High              float64 `json:"high"`
	TotalVolumeTraded int64   `json:"totalVolumeTraded"`
}

// MarketMakerQuote is a single market-maker bid/ask sample.
// Buy and Sell are nil when the market maker did not publish a quote at that timestamp.
// Timestamp is Unix epoch milliseconds.
type MarketMakerQuote struct {
	Buy       *float64 `json:"buy"`
	Sell      *float64 `json:"sell"`
	Timestamp int64    `json:"timestamp"`
}

// PriceChartMetadata describes the resolution of the returned series.
type PriceChartMetadata struct {
	Resolution PriceChartResolution `json:"resolution"`
}

// PriceChartResolution reports the resolution used for this response and the
// other resolutions the server is willing to return for this time period.
type PriceChartResolution struct {
	ChartResolution      string   `json:"chartResolution"`
	AvailableResolutions []string `json:"availableResolutions"`
}
