package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/zalgonoise/tendigitprimes/repository"
	"github.com/zalgonoise/x/cli"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/zalgonoise/tendigitprimes/config"
	"github.com/zalgonoise/tendigitprimes/grpcserver"
	"github.com/zalgonoise/tendigitprimes/httpserver"
	"github.com/zalgonoise/tendigitprimes/log"
	"github.com/zalgonoise/tendigitprimes/metrics"
	pb "github.com/zalgonoise/tendigitprimes/pb/primes/v1"
	"github.com/zalgonoise/tendigitprimes/primes"
	"github.com/zalgonoise/tendigitprimes/repository/sqlite"
)

// shutdownTimeout sets a duration for servers to terminate gracefully
const shutdownTimeout = 1 * time.Minute

func main() {
	runner := cli.NewRunner("primes",
		cli.WithOneOf("serve"),
		cli.WithExecutors(map[string]cli.Executor{
			"serve": cli.Executable(ExecService),
		}),
	)

	cli.Run(runner)
}

func ExecService(ctx context.Context, logger *slog.Logger, args []string) (int, error) {
	c, err := config.New()

	repo, err := sqlite.NewRepository(c.Database.URI)
	if err != nil {
		return 1, err
	}

	logger = log.From(c.LogLevel, logger.Handler())
	m := metrics.NewMetrics()
	m.RegisterCollector(collectors.NewDBStatsCollector(repo.DB, "primes"))
	m.RegisterCollector(repository.NewPingCollector(repo.DB, "primes"))
	m.InitRequestsMetrics("2", "9999999999")

	service := primes.NewService(repo, logger, m)

	server, err := httpserver.NewServer(fmt.Sprintf(":%d", c.Server.HTTPPort))
	if err != nil {
		return 1, err
	}

	if err := registerMetrics(server, m); err != nil {
		return 1, err
	}

	grpcServer, err := runGRPCServer(ctx, logger, &c.Server, service, server, m)
	if err != nil {
		return 1, err
	}

	go runHTTPServer(ctx, logger, &c.Server, server)

	return shutdown(server, grpcServer)
}

func registerMetrics(
	httpServer *httpserver.Server,
	m *metrics.Metrics,
) error {
	prometheusRegistry, err := m.Registry()
	if err != nil {
		return err
	}

	err = httpServer.RegisterHTTP(http.MethodGet, "/metrics",
		promhttp.HandlerFor(prometheusRegistry,
			promhttp.HandlerOpts{
				Registry:          prometheusRegistry,
				EnableOpenMetrics: true,
			}))
	if err != nil {
		return err
	}

	return nil
}

func runGRPCServer(
	ctx context.Context,
	logger *slog.Logger,
	cfg *config.Server,
	primes pb.PrimesServer,
	httpServer *httpserver.Server,
	m *metrics.Metrics,
) (*grpcserver.Server, error) {
	loggingOpts := []logging.Option{
		logging.WithLogOnEvents(logging.StartCall, logging.FinishCall),
	}

	grpcSrv := grpcserver.NewServer(m, []grpc.UnaryServerInterceptor{
		logging.UnaryServerInterceptor(log.InterceptorLogger(logger), loggingOpts...),
	}, []grpc.StreamServerInterceptor{
		logging.StreamServerInterceptor(log.InterceptorLogger(logger), loggingOpts...),
	})

	grpcSrv.RegisterPrimesServer(primes)
	logger.InfoContext(ctx, "listening on gRPC", slog.Int("port", cfg.GRPCPort))

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
	if err != nil {
		return nil, err
	}

	go func() {
		err = grpcSrv.Serve(lis)
		if err != nil {
			logger.ErrorContext(ctx, "failed to start gRPC server",
				slog.Int("port", cfg.GRPCPort),
				slog.String("error", err.Error()),
			)
			os.Exit(1)
		}
	}()

	grpcClient, err := grpc.Dial(
		fmt.Sprintf("localhost:%d", cfg.GRPCPort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	primesClient := pb.NewPrimesClient(grpcClient)

	err = httpServer.RegisterGRPC(context.Background(), primesClient)
	if err != nil {
		return nil, err
	}

	return grpcSrv, nil
}

func runHTTPServer(
	ctx context.Context, logger *slog.Logger,
	cfg *config.Server,
	httpServer *httpserver.Server,
) {
	logger.InfoContext(ctx, "listening on http", slog.Int("port", cfg.HTTPPort))

	err := httpServer.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.ErrorContext(ctx, "http server listen error",
			slog.Int("port", cfg.HTTPPort),
			slog.String("error", err.Error()),
		)
		os.Exit(1)
	}
}

func shutdown(
	httpServer *httpserver.Server,
	gRPCServer *grpcserver.Server,
) (int, error) {
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

	<-signalChannel

	shutdownContext, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := httpServer.Shutdown(shutdownContext); err != nil {
		return 1, err
	}

	gRPCServer.Shutdown()

	return 0, nil
}
