package app

import (
	"context"
	"net"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/warpgr/test_task/configs"
	"github.com/warpgr/test_task/controller"
	"github.com/warpgr/test_task/observability"
	pb "github.com/warpgr/test_task/proto"
	"github.com/warpgr/test_task/repository/db"
	"github.com/warpgr/test_task/repository/gateways"
	"github.com/warpgr/test_task/service"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func NewRatesServiceApp(cfg configs.RatesServiceConfig) RatesServiceApp {
	return &ratesServiceApp{cfg: cfg}
}

type ratesServiceApp struct {
	cfg        configs.RatesServiceConfig
	httpServer *http.Server

	daemons []service.Daemon

	conn *sqlx.DB

	grpcServer *grpc.Server

	cancel         context.CancelFunc
	shutdownTracer func(context.Context) error

	job sync.WaitGroup

	logger *zap.Logger
}

func (a *ratesServiceApp) Init(ctx context.Context) error {
	// Initialize logger
	logger, err := observability.InitZap(a.cfg.Service.LogLevel)
	if err != nil {
		return err
	}
	a.logger = logger
	a.logger.Info("Logger initialized successfully")

	// Initialize OpenTelemetry
	a.logger.Info("Initializing OpenTelemetry...", zap.String("service", a.cfg.Service.ServiceName), zap.String("endpoint", a.cfg.Service.OTLPEndpoint))
	shutdownTracer, err := observability.InitTracer(ctx, a.cfg.Service.ServiceName, a.cfg.Service.OTLPEndpoint)
	if err != nil {
		return err
	}
	a.shutdownTracer = shutdownTracer
	a.logger.Info("OpenTelemetry initialized successfully")

	a.logger.Info("Connecting to Database and applying migrations...", zap.String("host", a.cfg.DB.Host))
	conn, err := db.ConnectAndMigrate(a.cfg.DB, a.cfg.Service.MigrationsDir)
	if err != nil {
		return err
	}
	a.conn = conn
	a.logger.Info("Database connection and migrations completed successfully")

	snapshotsDB := db.NewOrderBookSnapshotsDB(a.conn)

	a.logger.Info("Initializing Grinex API client", zap.String("url", a.cfg.Service.GrinexAPIURL))
	client := gateways.NewGrinexAPIClient(a.cfg.Service.GrinexAPIURL, a.cfg.GrinexPairs)
	a.daemons = append(a.daemons,
		service.NewDaemon(
			snapshotsDB,
			client,
			a.cfg.GrinexPairs,
			a.logger.Named("daemon"),
		))

	ratesService := service.NewRatesService(snapshotsDB, a.logger.Named("service"))

	a.logger.Info("Registering gRPC handlers")
	a.grpcServer = grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	)
	pb.RegisterRateServiceServer(a.grpcServer, ratesService)

	// Initializing rest handlers.
	a.logger.Info("Initializing HTTP handlers for health check and metrics")
	routers := []controller.APIRouter{
		controller.NewHealthCheckController(),
	}

	g := gin.Default()
	for _, router := range routers {
		router.Register(g)
	}

	// Register Prometheus metrics endpoint
	g.GET("/metrics", gin.WrapH(promhttp.Handler()))

	a.httpServer = &http.Server{
		Addr:    a.cfg.Service.HealtechecEndpoint,
		Handler: g,
	}

	a.logger.Info("Initialization complete")
	return nil
}

func (a *ratesServiceApp) Run(parentCtx context.Context) <-chan error {
	ctx, cancel := context.WithCancel(parentCtx)
	a.cancel = cancel

	errChan := make(chan error, 1)
	go func() {
		a.job.Add(len(a.daemons) + 2)
		go func() {
			for _, daemon := range a.daemons {
				go func(daemon service.Daemon) {
					defer a.job.Done()
					if err := <-daemon.Run(ctx); err != nil {
						a.logger.Error("daemon error", zap.Error(err))
						errChan <- err
						return
					}
				}(daemon)
			}
		}()

		go func() {
			defer a.job.Done()
			lis, err := net.Listen("tcp", a.cfg.Service.Addr)
			if err != nil {
				a.logger.Error("failed to listen for grpc", zap.Error(err))
				errChan <- err
				return
			}
			a.logger.Info("starting grpc server", zap.String("addr", a.cfg.Service.Addr))
			if err := a.grpcServer.Serve(lis); err != nil {
				a.logger.Error("grpc server error", zap.Error(err))
				errChan <- err
			}
		}()

		go func() {
			defer a.job.Done()
			a.logger.Info("starting http server", zap.String("addr", a.cfg.Service.HealtechecEndpoint))
			if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				a.logger.Error("http server error", zap.Error(err))
				errChan <- err
			}
		}()

	}()

	return errChan
}

func (a *ratesServiceApp) Shutdown(ctx context.Context) error {
	a.logger.Info("shutting down application")
	a.cancel()

	err := a.httpServer.Shutdown(ctx)
	a.grpcServer.GracefulStop()

	if a.shutdownTracer != nil {
		if err := a.shutdownTracer(ctx); err != nil {
			a.logger.Error("failed to shutdown tracer", zap.Error(err))
		}
	}

	a.job.Wait()

	if a.conn != nil {
		_ = a.conn.Close()
	}

	a.logger.Info("application shutdown complete")
	return err
}
