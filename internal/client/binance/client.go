package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/kazah/golang-task-03/internal/domain"
)

// Default values and the Binance depth endpoint.
const (
	DefaultBaseURL = "https://api.binance.com"
	DefaultSymbol  = "USDTTRY"
	DefaultTimeout = 5 * time.Second

	depthPath = "/api/v3/depth"
	// how many levels we ask from Binance
	depthLimit = 20
)

// Client gets the order book from the exchange.
type Client interface {
	GetDepth(ctx context.Context, symbol string) (*domain.OrderBook, error)
}

// Config is the Binance client settings.
type Config struct {
	BaseURL string
	Timeout time.Duration
}

// binanceClient sends HTTP requests to Binance with resty.
type binanceClient struct {
	http   *resty.Client
	tracer trace.Tracer
}

// depthResponse is the Binance depth answer.
// Each level is a pair of strings: ["price", "quantity"].
type depthResponse struct {
	Asks [][]string `json:"asks"`
	Bids [][]string `json:"bids"`
}

// New creates the Binance client.
func New(cfg Config) Client {
	if cfg.BaseURL == "" {
		cfg.BaseURL = DefaultBaseURL
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = DefaultTimeout
	}

	r := resty.New().
		SetBaseURL(cfg.BaseURL).
		SetTimeout(cfg.Timeout)

	return &binanceClient{
		http:   r,
		tracer: otel.Tracer("binance-client"),
	}
}

// GetDepth requests the order book for the symbol, for example USDTTRY.
func (c *binanceClient) GetDepth(ctx context.Context, symbol string) (*domain.OrderBook, error) {
	ctx, span := c.tracer.Start(ctx, "BinanceClient.GetDepth",
		trace.WithAttributes(attribute.String("exchange.symbol", symbol)),
	)
	defer span.End()

	if symbol == "" {
		symbol = DefaultSymbol
	}

	resp, err := c.http.R().
		SetContext(ctx).
		SetQueryParam("symbol", strings.ToUpper(symbol)).
		SetQueryParam("limit", strconv.Itoa(depthLimit)).
		Get(depthPath)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("%w: %w", domain.ErrExternalAPIUnavailable, err)
	}

	if resp.IsError() {
		err = fmt.Errorf("%w: status %d", domain.ErrExternalAPIUnavailable, resp.StatusCode())
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	var raw depthResponse
	if err := json.Unmarshal(resp.Body(), &raw); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("failed to parse binance response: %w", err)
	}

	asks, err := parseLevels(raw.Asks)
	if err != nil {
		return nil, fmt.Errorf("failed to parse asks: %w", err)
	}
	bids, err := parseLevels(raw.Bids)
	if err != nil {
		return nil, fmt.Errorf("failed to parse bids: %w", err)
	}

	// Binance does not send a timestamp in the depth response,
	// so we use the time when we got the answer.
	return &domain.OrderBook{
		Timestamp: time.Now().Unix(),
		Asks:      asks,
		Bids:      bids,
	}, nil
}

// parseLevels converts Binance levels from strings to numbers.
func parseLevels(items [][]string) ([]domain.PriceLevel, error) {
	levels := make([]domain.PriceLevel, 0, len(items))
	for _, item := range items {
		if len(item) < 2 {
			return nil, fmt.Errorf("bad level: %v", item)
		}
		price, err := strconv.ParseFloat(item[0], 64)
		if err != nil {
			return nil, fmt.Errorf("bad price %q: %w", item[0], err)
		}
		volume, err := strconv.ParseFloat(item[1], 64)
		if err != nil {
			return nil, fmt.Errorf("bad volume %q: %w", item[1], err)
		}
		levels = append(levels, domain.PriceLevel{Price: price, Volume: volume})
	}
	return levels, nil
}
