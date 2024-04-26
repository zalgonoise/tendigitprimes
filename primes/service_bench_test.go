//go:build bench

package primes

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zalgonoise/tendigitprimes/config"
	"github.com/zalgonoise/tendigitprimes/database"
	"github.com/zalgonoise/tendigitprimes/log"
	"github.com/zalgonoise/tendigitprimes/metrics"
	pb "github.com/zalgonoise/tendigitprimes/pb/primes/v1"
	"github.com/zalgonoise/tendigitprimes/repository/sqlite"
)

func BenchmarkService(b *testing.B) {
	c := &config.Primes{
		LogLevel: "debug",
		Database: config.Database{
			URI:         "./testdata",
			Partitioned: true,
		},
		Server: config.Server{
			HTTPPort: 8080,
			GRPCPort: 8081,
		},
	}

	logger := log.New(c.LogLevel)
	m := metrics.Noop{}

	db, err := database.AttachSQLite(c.Database.URI, database.ReadOnlyPragmas(), logger)
	require.NoError(b, err)

	repo, err := sqlite.NewPartitionSet(db)
	require.NoError(b, err)

	service := NewService(repo, logger, m)

	ctx := context.Background()
	logger.InfoContext(ctx, "service is ready")

	b.Run("Benchmark/Random", func(b *testing.B) {
		var res *pb.RandomResponse

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			res, err = service.Random(ctx, &pb.RandomRequest{
				Min: 1_000_000_000,
				Max: 5_000_000_000,
			})
			if err != nil {
				b.Fatal(err)

				return
			}

			_ = res
		}
	})

	b.Run("Benchmark/List", func(b *testing.B) {
		var res *pb.ListResponse

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			res, err = service.List(ctx, &pb.ListRequest{
				Min:        1_000_000_000,
				Max:        5_000_000_000,
				MaxResults: 10,
			})
			if err != nil {
				b.Fatal(err)

				return
			}

			_ = res
		}
	})

	require.NoError(b, repo.Close())
}
