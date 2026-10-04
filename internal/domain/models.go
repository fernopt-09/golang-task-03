package domain

import (
	"time"
)

// CalculationMethod is the way how the final rate is taken from the order book.
type CalculationMethod string

// Supported calculation methods.
const (
	// MethodTopN takes the price from position N.
	MethodTopN CalculationMethod = "topN"
	// MethodAvgNM takes the average price of positions from N to M.
	MethodAvgNM CalculationMethod = "avgNM"
)

// PriceLevel is one line of the order book.
type PriceLevel struct {
	Price  float64
	Volume float64
}

// OrderBook contains asks and bids from the exchange.
type OrderBook struct {
	Timestamp int64
	Asks      []PriceLevel
	Bids      []PriceLevel
}

// CalculationParams are the parameters of GetRates request.
type CalculationParams struct {
	Method CalculationMethod
	N      uint32
	M      uint32
}

// Rate is the calculated rate that is saved to the database.
type Rate struct {
	ID        int64
	Ask       float64
	Bid       float64
	Timestamp time.Time
	Method    CalculationMethod
	Params    map[string]any
	CreatedAt time.Time
}
