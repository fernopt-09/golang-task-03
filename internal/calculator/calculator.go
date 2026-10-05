package calculator

import (
	"cmp"
	"fmt"
	"slices"

	"github.com/kazah/golang-task-03/internal/domain"
)

// Calculator calculates the final ask and bid from the order book.
type Calculator interface {
	Calculate(ob *domain.OrderBook, params domain.CalculationParams) (ask float64, bid float64, err error)
	CalculateTopN(ob *domain.OrderBook, n uint32) (ask float64, bid float64, err error)
	CalculateAvgNM(ob *domain.OrderBook, n, m uint32) (ask float64, bid float64, err error)
}

// rateCalculator is the default Calculator, it has no state.
type rateCalculator struct{}

// New creates the calculator.
func New() Calculator {
	return &rateCalculator{}
}

// Calculate chooses the method from params. Zero N and M are replaced by defaults.
func (c *rateCalculator) Calculate(ob *domain.OrderBook, params domain.CalculationParams) (float64, float64, error) {
	switch params.Method {
	case domain.MethodTopN:
		n := params.N
		if n == 0 {
			n = 1
		}
		return c.CalculateTopN(ob, n)
	case domain.MethodAvgNM:
		n, m := params.N, params.M
		if n == 0 {
			n = 1
		}
		if m == 0 {
			m = n
		}
		return c.CalculateAvgNM(ob, n, m)
	default:
		return 0, 0, fmt.Errorf("%w: %s", domain.ErrUnknownMethod, params.Method)
	}
}

// CalculateTopN returns ask and bid prices from position n (starts from 1).
func (c *rateCalculator) CalculateTopN(ob *domain.OrderBook, n uint32) (float64, float64, error) {
	if ob == nil || len(ob.Asks) == 0 || len(ob.Bids) == 0 {
		return 0, 0, domain.ErrEmptyOrderBook
	}

	if n < 1 {
		return 0, 0, fmt.Errorf("%w: position N must be >= 1", domain.ErrInvalidParameters)
	}

	idx := int(n - 1)
	if idx >= len(ob.Asks) || idx >= len(ob.Bids) {
		return 0, 0, fmt.Errorf("%w: position %d exceeds depth (asks: %d, bids: %d)",
			domain.ErrInsufficientDepth, n, len(ob.Asks), len(ob.Bids))
	}

	asks := sortAsks(ob.Asks)
	bids := sortBids(ob.Bids)

	return asks[idx].Price, bids[idx].Price, nil
}

// CalculateAvgNM returns average ask and bid prices of positions from n to m.
func (c *rateCalculator) CalculateAvgNM(ob *domain.OrderBook, n, m uint32) (float64, float64, error) {
	if ob == nil || len(ob.Asks) == 0 || len(ob.Bids) == 0 {
		return 0, 0, domain.ErrEmptyOrderBook
	}

	if n < 1 {
		return 0, 0, fmt.Errorf("%w: start position N must be >= 1", domain.ErrInvalidParameters)
	}

	if m < n {
		return 0, 0, fmt.Errorf("%w: end position M (%d) cannot be less than start position N (%d)",
			domain.ErrInvalidParameters, m, n)
	}

	startIdx := int(n - 1)
	endIdx := int(m - 1)

	if endIdx >= len(ob.Asks) || endIdx >= len(ob.Bids) {
		return 0, 0, fmt.Errorf("%w: end position %d exceeds depth (asks: %d, bids: %d)",
			domain.ErrInsufficientDepth, m, len(ob.Asks), len(ob.Bids))
	}

	asks := sortAsks(ob.Asks)
	bids := sortBids(ob.Bids)

	count := float64(endIdx - startIdx + 1)
	var askSum, bidSum float64

	for i := startIdx; i <= endIdx; i++ {
		askSum += asks[i].Price
		bidSum += bids[i].Price
	}

	return askSum / count, bidSum / count, nil
}

// asks are sorted from the lowest price to the highest one
func sortAsks(asks []domain.PriceLevel) []domain.PriceLevel {
	cp := slices.Clone(asks)
	slices.SortFunc(cp, func(a, b domain.PriceLevel) int {
		return cmp.Compare(a.Price, b.Price)
	})
	return cp
}

// bids are sorted from the highest price to the lowest one
func sortBids(bids []domain.PriceLevel) []domain.PriceLevel {
	cp := slices.Clone(bids)
	slices.SortFunc(cp, func(a, b domain.PriceLevel) int {
		return cmp.Compare(b.Price, a.Price)
	})
	return cp
}
