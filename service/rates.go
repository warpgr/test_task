package service

import (
	"context"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	pb "github.com/warpgr/test_task/proto"
	"github.com/warpgr/test_task/repository/db"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ratesRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "rates_requests_total",
		Help: "The total number of rates requests",
	}, []string{"pair", "method"})

	ratesRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "rates_request_duration_seconds",
		Help:    "The duration of rates requests in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"pair", "method"})
)

func NewRatesService(ratesDB db.OrderBookSnapshotsDBReader, logger *zap.Logger) RateService {
	return &ratesSerivce{
		ratesDB: ratesDB,
		logger:  logger,
	}
}

type ratesSerivce struct {
	pb.UnimplementedRateServiceServer

	ratesDB db.OrderBookSnapshotsDBReader
	logger  *zap.Logger
}

func (svc *ratesSerivce) GetRates(ctx context.Context, req *pb.GetRatesRequest) (*pb.GetRatesResponse, error) {
	ctx, span := otel.Tracer("rates-service").Start(ctx, "GetRates", trace.WithAttributes(
		attribute.String("pair", req.Pair),
		attribute.String("method", req.CalcMethod.String()),
	))
	defer span.End()

	startTime := time.Now()
	defer func() {
		ratesRequestsTotal.WithLabelValues(req.Pair, req.CalcMethod.String()).Inc()
		ratesRequestDuration.WithLabelValues(req.Pair, req.CalcMethod.String()).Observe(time.Since(startTime).Seconds())
	}()

	svc.logger.Info("GetRates request received", zap.String("pair", req.Pair), zap.String("method", req.CalcMethod.String()))
	if req.Pair == "" {
		return nil, status.Error(codes.InvalidArgument, "pair is required")
	}

	// 1. Read from DB
	orderBook, err := svc.ratesDB.GetOrderBookSnapshot(ctx, req.Pair)
	if err != nil {
		if err == db.ErrNotFound {
			svc.logger.Warn("no data found for pair", zap.String("pair", req.Pair))
			return nil, status.Errorf(codes.NotFound, "no data found for pair %s", req.Pair)
		}
		svc.logger.Error("failed to read snapshot from DB", zap.String("pair", req.Pair), zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to read snapshot from DB: %v", err)
	}

	// 2. Calculate rates
	var ask, bid float64

	switch req.CalcMethod {
	case pb.CalcMethod_CALC_METHOD_TOP_N:
		n := int(req.N)
		if n < 0 || n >= len(orderBook.Asks) || n >= len(orderBook.Bids) {
			return nil, status.Errorf(codes.InvalidArgument, "index N=%d out of range (asks=%d, bids=%d)", n, len(orderBook.Asks), len(orderBook.Bids))
		}
		ask = orderBook.Asks[n].Price
		bid = orderBook.Bids[n].Price

	case pb.CalcMethod_CALC_METHOD_AVG_N_M:
		n, m := int(req.N), int(req.M)
		if n < 0 || m < n || m >= len(orderBook.Asks) || m >= len(orderBook.Bids) {
			return nil, status.Errorf(codes.InvalidArgument, "range [%d, %d] invalid or out of range", n, m)
		}

		var sumAsk, sumBid float64
		count := float64(m - n + 1)
		for i := n; i <= m; i++ {
			sumAsk += orderBook.Asks[i].Price
			sumBid += orderBook.Bids[i].Price
		}
		ask = sumAsk / count
		bid = sumBid / count

	default:
		return nil, status.Error(codes.InvalidArgument, "invalid or unspecified calculation method")
	}

	return &pb.GetRatesResponse{
		Ask:         ask,
		Bid:         bid,
		TimestampMs: orderBook.Timestamp,
	}, nil
}

func (svc *ratesSerivce) HealthCheck(ctx context.Context, req *empty.Empty) (*pb.HealthCheckResponse, error) {
	return &pb.HealthCheckResponse{
		IsHealthy:     true,
		StatusMessage: "OK",
	}, nil
}
