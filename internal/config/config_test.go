package config_test

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kazah/golang-task-03/internal/config"
)

func TestConfig_Defaults(t *testing.T) {
	cfg, err := config.Load([]string{})
	require.NoError(t, err)

	assert.Equal(t, 50051, cfg.GRPCPort)
	assert.Equal(t, 9090, cfg.MetricsPort)
	assert.Equal(t, "info", cfg.LogLevel)
	assert.Equal(t, "localhost", cfg.DB.Host)
	assert.Equal(t, 5432, cfg.DB.Port)
	assert.Equal(t, "postgres", cfg.DB.User)
	assert.Equal(t, "ratedb", cfg.DB.DBName)
	assert.Equal(t, "https://api.binance.com", cfg.Binance.BaseURL)
}

func TestConfig_EnvOverride(t *testing.T) {
	_ = os.Setenv("GRPC_PORT", "50055")
	_ = os.Setenv("DB_HOST", "db.example.com")
	_ = os.Setenv("BINANCE_TIMEOUT", "10s")
	defer func() {
		_ = os.Unsetenv("GRPC_PORT")
		_ = os.Unsetenv("DB_HOST")
		_ = os.Unsetenv("BINANCE_TIMEOUT")
	}()

	cfg, err := config.Load([]string{})
	require.NoError(t, err)

	assert.Equal(t, 50055, cfg.GRPCPort)
	assert.Equal(t, "db.example.com", cfg.DB.Host)
	assert.Equal(t, 10*time.Second, cfg.Binance.Timeout)
}

func TestConfig_FlagOverride(t *testing.T) {
	_ = os.Setenv("GRPC_PORT", "50055")
	_ = os.Setenv("DB_HOST", "db.from.env")
	defer func() {
		_ = os.Unsetenv("GRPC_PORT")
		_ = os.Unsetenv("DB_HOST")
	}()

	args := []string{
		"-grpc-port=50060",
		"-db-host=db.from.flag",
	}

	cfg, err := config.Load(args)
	require.NoError(t, err)

	assert.Equal(t, 50060, cfg.GRPCPort)
	assert.Equal(t, "db.from.flag", cfg.DB.Host)
}
