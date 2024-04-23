package primes

import (
	"context"
	"log/slog"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/zalgonoise/tendigitprimes/metrics"
	pb "github.com/zalgonoise/tendigitprimes/pb/primes/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	defaultMin int64 = 2
	defaultMax int64 = 9999999999
)

type Repository interface {
	Random(ctx context.Context, min, max int64) (int64, error)
	List(ctx context.Context, min, max, limit int64) ([]int64, error)
	Close() error
}

type Metrics interface {
	RegisterCollector(collector prometheus.Collector)

	InitRequestsMetrics(minimum string, maximum string)
	IncRequestsReceivedTotal(minimum string, maximum string)
	IncRequestsReceivedErrored(minimum string, maximum string)
	ObserveRequestLatency(ctx context.Context, minimum string, maximum string, duration time.Duration)
}

type Service struct {
	pb.UnimplementedPrimesServer

	repo Repository

	m      Metrics
	logger *slog.Logger
}

func (s Service) Random(ctx context.Context, req *pb.RandomRequest) (*pb.RandomResponse, error) {
	if err := req.Validate(); err != nil {
		s.logger.WarnContext(ctx, "invalid request",
			slog.Any("request", req),
			slog.String("error", err.Error()),
		)

		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	start := time.Now()
	minString := strconv.Itoa(int(req.Min))
	maxString := strconv.Itoa(int(req.Max))

	defer func() {
		s.m.ObserveRequestLatency(ctx, minString, maxString, time.Since(start))
	}()

	s.m.IncRequestsReceivedTotal(minString, maxString)

	prime, err := s.repo.Random(ctx, req.Min, req.Max)
	if err != nil {
		s.m.IncRequestsReceivedErrored(minString, maxString)
		s.logger.ErrorContext(ctx, "failed to get prime number",
			slog.Int64("min", req.Min),
			slog.Int64("max", req.Max),
			slog.String("error", err.Error()),
		)

		return nil, status.Error(codes.Internal, err.Error())
	}

	slog.DebugContext(ctx, "fetched prime number", slog.Int64("prime_number", prime))

	return &pb.RandomResponse{Prime: prime}, nil
}

func (s Service) List(ctx context.Context, req *pb.ListRequest) (*pb.ListResponse, error) {
	if err := req.Validate(); err != nil {
		s.logger.WarnContext(ctx, "invalid request",
			slog.Any("request", req),
			slog.String("error", err.Error()),
		)

		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	start := time.Now()
	minString := strconv.Itoa(int(req.Min))
	maxString := strconv.Itoa(int(req.Max))

	defer func() {
		s.m.ObserveRequestLatency(ctx, minString, maxString, time.Since(start))
	}()

	s.m.IncRequestsReceivedTotal(minString, maxString)

	primes, err := s.repo.List(ctx, req.Min, req.Max, req.MaxResults)
	if err != nil {
		s.m.IncRequestsReceivedErrored(minString, maxString)
		s.logger.ErrorContext(ctx, "failed to get prime numbers list",
			slog.Int64("min", req.Min),
			slog.Int64("max", req.Max),
			slog.String("error", err.Error()),
		)

		return nil, status.Error(codes.Internal, err.Error())
	}

	slog.DebugContext(ctx, "fetched prime numbers list", slog.Int("num_prime_numbers", len(primes)))

	return &pb.ListResponse{Primes: primes}, nil
}

func NewService(repo Repository, logger *slog.Logger, m *metrics.Metrics) Service {
	return Service{
		repo:   repo,
		logger: logger,
		m:      m,
	}
}
