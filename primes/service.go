package primes

import (
	"context"
	"log/slog"
	"strconv"
	"time"

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
	Get(ctx context.Context, min, max int64) (int64, error)
}

type Service struct {
	pb.UnimplementedPrimesServer

	repo Repository

	m      *metrics.Metrics
	logger *slog.Logger
}

func (s Service) Get(ctx context.Context, req *pb.PrimeRequest) (*pb.PrimeResponse, error) {
	if req.Min < defaultMin {
		req.Min = defaultMin
	}

	if req.Max > defaultMax {
		req.Max = defaultMax
	}

	start := time.Now()
	minString := strconv.Itoa(int(req.Min))
	maxString := strconv.Itoa(int(req.Max))

	defer func() {
		s.m.ObserveRequestLatency(ctx, minString, maxString, time.Since(start))
	}()

	s.m.IncRequestsReceivedTotal(minString, maxString)

	prime, err := s.repo.Get(ctx, req.Min, req.Max)
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

	return &pb.PrimeResponse{Prime: prime}, nil
}

func NewService(repo Repository, logger *slog.Logger, m *metrics.Metrics) Service {
	return Service{
		repo:   repo,
		logger: logger,
		m:      m,
	}
}
