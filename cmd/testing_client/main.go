package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/warpgr/test_task/observability"
	pb "github.com/warpgr/test_task/proto"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Initialize Tracer for the client
	shutdown, err := observability.InitTracer(context.Background(), "testing-client", "localhost:4317")
	if err != nil {
		log.Printf("Failed to initialize tracer: %v", err)
	} else {
		defer shutdown(context.Background())
	}

	// Connect to gRPC server
	conn, err := grpc.NewClient("localhost:5001",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewRateServiceClient(conn)

	// Context for cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Signal handling
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		log.Printf("Received signal: %s. Shutting down...", sig)
		cancel()
	}()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	log.Println("Starting testing client. Press Ctrl-C to exit.")

	for {
		select {
		case <-ctx.Done():
			log.Println("Client stopped.")
			return
		case <-ticker.C:
			resp, err := client.GetRates(ctx, &pb.GetRatesRequest{
				Pair:       "USDT",
				CalcMethod: pb.CalcMethod_CALC_METHOD_TOP_N,
				N:          0,
			})
			if err != nil {
				log.Printf("Error calling GetRates: %v", err)
				continue
			}
			log.Printf("Rates for USDT: Ask=%f, Bid=%f, TS=%d", resp.Ask, resp.Bid, resp.TimestampMs)
		}
	}
}
