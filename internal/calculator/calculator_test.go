package calculator_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kazah/golang-task-03/internal/calculator"
	"github.com/kazah/golang-task-03/internal/domain"
)

func sampleOrderBook() *domain.OrderBook {
	return &domain.OrderBook{
		Timestamp: 1712000000,
		Asks: []domain.PriceLevel{
			{Price: 97.0, Volume: 1.0},
			{Price: 95.0, Volume: 2.0},
			{Price: 99.0, Volume: 0.5},
			{Price: 96.0, Volume: 1.5},
			{Price: 98.0, Volume: 3.0},
		},
		Bids: []domain.PriceLevel{
			{Price: 92.0, Volume: 1.0},
			{Price: 94.0, Volume: 2.0},
			{Price: 90.0, Volume: 0.5},
			{Price: 93.0, Volume: 1.5},
			{Price: 91.0, Volume: 3.0},
		},
	}
}

func TestCalculator_TopN(t *testing.T) {
	calc := calculator.New()
	ob := sampleOrderBook()

	t.Run("Top 1 (best ask and bid)", func(t *testing.T) {
		ask, bid, err := calc.CalculateTopN(ob, 1)
		require.NoError(t, err)
		assert.Equal(t, 95.0, ask)
		assert.Equal(t, 94.0, bid)
	})

	t.Run("Top 3", func(t *testing.T) {
		ask, bid, err := calc.CalculateTopN(ob, 3)
		require.NoError(t, err)
		assert.Equal(t, 97.0, ask)
		assert.Equal(t, 92.0, bid)
	})

	t.Run("Top 5 (last position)", func(t *testing.T) {
		ask, bid, err := calc.CalculateTopN(ob, 5)
		require.NoError(t, err)
		assert.Equal(t, 99.0, ask)
		assert.Equal(t, 90.0, bid)
	})

	t.Run("Error when N is zero", func(t *testing.T) {
		_, _, err := calc.CalculateTopN(ob, 0)
		require.ErrorIs(t, err, domain.ErrInvalidParameters)
	})

	t.Run("Error when N exceeds depth", func(t *testing.T) {
		_, _, err := calc.CalculateTopN(ob, 6)
		require.ErrorIs(t, err, domain.ErrInsufficientDepth)
	})

	t.Run("Error on empty orderbook", func(t *testing.T) {
		_, _, err := calc.CalculateTopN(nil, 1)
		require.ErrorIs(t, err, domain.ErrEmptyOrderBook)

		_, _, err = calc.CalculateTopN(&domain.OrderBook{}, 1)
		require.ErrorIs(t, err, domain.ErrEmptyOrderBook)
	})
}

func TestCalculator_AvgNM(t *testing.T) {
	calc := calculator.New()
	ob := sampleOrderBook()

	t.Run("Range [1, 1] equals Top 1", func(t *testing.T) {
		ask, bid, err := calc.CalculateAvgNM(ob, 1, 1)
		require.NoError(t, err)
		assert.Equal(t, 95.0, ask)
		assert.Equal(t, 94.0, bid)
	})

	t.Run("Range [1, 3]", func(t *testing.T) {
		ask, bid, err := calc.CalculateAvgNM(ob, 1, 3)
		require.NoError(t, err)
		assert.Equal(t, 96.0, ask)
		assert.Equal(t, 93.0, bid)
	})

	t.Run("Range [2, 4]", func(t *testing.T) {
		ask, bid, err := calc.CalculateAvgNM(ob, 2, 4)
		require.NoError(t, err)
		assert.Equal(t, 97.0, ask)
		assert.Equal(t, 92.0, bid)
	})

	t.Run("Error when M < N", func(t *testing.T) {
		_, _, err := calc.CalculateAvgNM(ob, 3, 2)
		require.ErrorIs(t, err, domain.ErrInvalidParameters)
	})

	t.Run("Error when N < 1", func(t *testing.T) {
		_, _, err := calc.CalculateAvgNM(ob, 0, 2)
		require.ErrorIs(t, err, domain.ErrInvalidParameters)
	})

	t.Run("Error when M exceeds depth", func(t *testing.T) {
		_, _, err := calc.CalculateAvgNM(ob, 1, 6)
		require.ErrorIs(t, err, domain.ErrInsufficientDepth)
	})

	t.Run("Error on empty orderbook", func(t *testing.T) {
		_, _, err := calc.CalculateAvgNM(nil, 1, 2)
		require.ErrorIs(t, err, domain.ErrEmptyOrderBook)
	})
}

func TestCalculator_Calculate(t *testing.T) {
	calc := calculator.New()
	ob := sampleOrderBook()

	t.Run("TopN method with defaults", func(t *testing.T) {
		ask, bid, err := calc.Calculate(ob, domain.CalculationParams{
			Method: domain.MethodTopN,
			N:      0,
		})
		require.NoError(t, err)
		assert.Equal(t, 95.0, ask)
		assert.Equal(t, 94.0, bid)
	})

	t.Run("AvgNM method with defaults", func(t *testing.T) {
		ask, bid, err := calc.Calculate(ob, domain.CalculationParams{
			Method: domain.MethodAvgNM,
			N:      0,
			M:      0,
		})
		require.NoError(t, err)
		assert.Equal(t, 95.0, ask)
		assert.Equal(t, 94.0, bid)
	})

	t.Run("Unknown method", func(t *testing.T) {
		_, _, err := calc.Calculate(ob, domain.CalculationParams{
			Method: "unsupported",
		})
		require.ErrorIs(t, err, domain.ErrUnknownMethod)
	})
}
