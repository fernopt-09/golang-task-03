package binance_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kazah/golang-task-03/internal/client/binance"
	"github.com/kazah/golang-task-03/internal/domain"
)

func TestGetDepth_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v3/depth", r.URL.Path)
		assert.Equal(t, "USDTTRY", r.URL.Query().Get("symbol"))

		_, _ = w.Write([]byte(`{
			"lastUpdateId": 1,
			"asks": [["49.12", "100.5"], ["49.13", "200"]],
			"bids": [["49.11", "150"], ["49.10", "300"]]
		}`))
	}))
	defer server.Close()

	client := binance.New(binance.Config{BaseURL: server.URL, Timeout: 2 * time.Second})

	// lower case symbol must be converted to upper case
	ob, err := client.GetDepth(context.Background(), "usdttry")
	require.NoError(t, err)

	require.Len(t, ob.Asks, 2)
	assert.Equal(t, 49.12, ob.Asks[0].Price)
	assert.Equal(t, 100.5, ob.Asks[0].Volume)
	require.Len(t, ob.Bids, 2)
	assert.Equal(t, 49.11, ob.Bids[0].Price)
	assert.NotZero(t, ob.Timestamp)
}

func TestGetDepth_DefaultSymbol(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, binance.DefaultSymbol, r.URL.Query().Get("symbol"))
		_, _ = w.Write([]byte(`{"asks": [], "bids": []}`))
	}))
	defer server.Close()

	client := binance.New(binance.Config{BaseURL: server.URL})

	_, err := client.GetDepth(context.Background(), "")
	require.NoError(t, err)
}

func TestGetDepth_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"code": -1121, "msg": "Invalid symbol."}`))
	}))
	defer server.Close()

	client := binance.New(binance.Config{BaseURL: server.URL})

	_, err := client.GetDepth(context.Background(), "BAD")
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrExternalAPIUnavailable)
}

func TestGetDepth_BadPrice(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"asks": [["abc", "1"]], "bids": []}`))
	}))
	defer server.Close()

	client := binance.New(binance.Config{BaseURL: server.URL})

	_, err := client.GetDepth(context.Background(), "USDTTRY")
	require.Error(t, err)
}
