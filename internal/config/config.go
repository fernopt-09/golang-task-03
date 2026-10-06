package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/kazah/golang-task-03/internal/client/binance"
	"github.com/kazah/golang-task-03/internal/repository/postgres"
)

// Config is the full app configuration.
type Config struct {
	GRPCPort    int
	MetricsPort int
	LogLevel    string
	AutoMigrate bool

	DB      postgres.Config
	Binance binance.Config
	Symbol  string
}

// Load reads config from environment variables and flags.
// A flag has higher priority than an environment variable.
func Load(args []string) (*Config, error) {
	fs := flag.NewFlagSet("rate-service", flag.ContinueOnError)

	defaultGRPCPort := getEnvInt("GRPC_PORT", 50051)
	defaultMetricsPort := getEnvInt("METRICS_PORT", 9090)
	defaultLogLevel := getEnvStr("LOG_LEVEL", "info")
	defaultAutoMigrate := getEnvBool("AUTO_MIGRATE", true)

	defaultDBHost := getEnvStr("DB_HOST", "localhost")
	defaultDBPort := getEnvInt("DB_PORT", 5432)
	defaultDBUser := getEnvStr("DB_USER", "postgres")
	defaultDBPassword := getEnvStr("DB_PASSWORD", "postgres")
	defaultDBName := getEnvStr("DB_NAME", "ratedb")
	defaultDBSSLMode := getEnvStr("DB_SSLMODE", "disable")

	defaultExchangeURL := getEnvStr("BINANCE_API_URL", binance.DefaultBaseURL)
	defaultExchangeSymbol := getEnvStr("BINANCE_SYMBOL", binance.DefaultSymbol)
	defaultExchangeTimeout := getEnvDuration("BINANCE_TIMEOUT", binance.DefaultTimeout)

	grpcPort := fs.Int("grpc-port", defaultGRPCPort, "gRPC server port")
	metricsPort := fs.Int("metrics-port", defaultMetricsPort, "Prometheus metrics HTTP port")
	logLevel := fs.String("log-level", defaultLogLevel, "log level: debug, info, warn, error")
	autoMigrate := fs.Bool("auto-migrate", defaultAutoMigrate, "run DB migrations on startup")

	dbHost := fs.String("db-host", defaultDBHost, "PostgreSQL host")
	dbPort := fs.Int("db-port", defaultDBPort, "PostgreSQL port")
	dbUser := fs.String("db-user", defaultDBUser, "PostgreSQL user")
	dbPassword := fs.String("db-password", defaultDBPassword, "PostgreSQL password")
	dbName := fs.String("db-name", defaultDBName, "PostgreSQL db name")
	dbSSLMode := fs.String("db-sslmode", defaultDBSSLMode, "PostgreSQL ssl mode")

	exchangeURL := fs.String("exchange-url", defaultExchangeURL, "Binance exchange base URL")
	exchangeSymbol := fs.String("exchange-symbol", defaultExchangeSymbol, "trading symbol")
	exchangeTimeout := fs.Duration("exchange-timeout", defaultExchangeTimeout, "HTTP client timeout")

	if err := fs.Parse(args); err != nil {
		return nil, fmt.Errorf("failed to parse flags: %w", err)
	}

	return &Config{
		GRPCPort:    *grpcPort,
		MetricsPort: *metricsPort,
		LogLevel:    *logLevel,
		AutoMigrate: *autoMigrate,
		DB: postgres.Config{
			Host:     *dbHost,
			Port:     *dbPort,
			User:     *dbUser,
			Password: *dbPassword,
			DBName:   *dbName,
			SSLMode:  *dbSSLMode,
		},
		Binance: binance.Config{
			BaseURL: *exchangeURL,
			Timeout: *exchangeTimeout,
		},
		Symbol: *exchangeSymbol,
	}, nil
}

// getEnvStr returns the env variable or the default value if it is empty.
func getEnvStr(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return defaultVal
}

// getEnvInt is like getEnvStr, but for numbers. A broken value gives the default one.
func getEnvInt(key string, defaultVal int) int {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}

// getEnvBool is like getEnvStr, but for bool values.
func getEnvBool(key string, defaultVal bool) bool {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		if b, err := strconv.ParseBool(val); err == nil {
			return b
		}
	}
	return defaultVal
}

// getEnvDuration is like getEnvStr, but for durations like "5s".
func getEnvDuration(key string, defaultVal time.Duration) time.Duration {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			return d
		}
	}
	return defaultVal
}
