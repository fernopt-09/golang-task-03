package domain

import "errors"

// Errors of the app. Other layers wrap them to add details.
var (
	ErrEmptyOrderBook         = errors.New("orderbook is empty")
	ErrInsufficientDepth      = errors.New("orderbook depth is insufficient for requested position")
	ErrInvalidParameters      = errors.New("invalid calculation parameters")
	ErrUnknownMethod          = errors.New("unknown calculation method")
	ErrExternalAPIUnavailable = errors.New("external exchange API is unavailable")
	ErrRateNotFound           = errors.New("rate record not found")
)
