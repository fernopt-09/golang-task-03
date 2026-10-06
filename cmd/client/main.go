package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	ratev1 "github.com/kazah/golang-task-03/gen/rate/v1"
)

// main is a small CLI to call the service by hand: GetRates or HealthCheck.
func main() {
	addr := flag.String("addr", "localhost:50051", "gRPC server address")
	method := flag.String("method", "topN", "calculation method: topN or avgNM")
	n := flag.Uint("n", 1, "position N (1-based)")
	m := flag.Uint("m", 1, "position M for avgNM (1-based)")
	healthCheck := flag.Bool("health", false, "check service health")
	flag.Parse()

	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect to %s: %v\n", *addr, err)
		os.Exit(1)
	}
	defer func() { _ = conn.Close() }()

	client := ratev1.NewRateServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if *healthCheck {
		resp, err := client.HealthCheck(ctx, &ratev1.HealthCheckRequest{})
		if err != nil {
			fmt.Fprintf(os.Stderr, "healthcheck failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("healthcheck status: %s\n", resp.Status.String())
		return
	}

	// convert the flag value to the proto enum
	var protoMethod ratev1.CalculationMethod
	switch *method {
	case "topN":
		protoMethod = ratev1.CalculationMethod_CALCULATION_METHOD_TOP_N
	case "avgNM":
		protoMethod = ratev1.CalculationMethod_CALCULATION_METHOD_AVG_NM
	default:
		fmt.Fprintf(os.Stderr, "unknown method %q (supported: topN, avgNM)\n", *method)
		os.Exit(1)
	}

	// positions are small numbers, so the conversion is safe
	nUint32 := uint32(*n) //nolint:gosec
	mUint32 := uint32(*m) //nolint:gosec

	req := &ratev1.GetRatesRequest{
		Method: protoMethod,
		N:      &nUint32,
		M:      &mUint32,
	}

	resp, err := client.GetRates(ctx, req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "GetRates error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("=== USDT Rate Response ===\n")
	fmt.Printf("Method:    %s\n", resp.Method.String())
	fmt.Printf("Ask Price: %.4f\n", resp.Ask)
	fmt.Printf("Bid Price: %.4f\n", resp.Bid)
	fmt.Printf("Timestamp: %s (unix: %d)\n", time.Unix(resp.Timestamp, 0).UTC().Format(time.RFC3339), resp.Timestamp)
	fmt.Printf("N: %d, M: %d\n", resp.N, resp.M)
}
