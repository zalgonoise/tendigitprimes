//go:build bench

package sqlite

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zalgonoise/tendigitprimes/config"
	"github.com/zalgonoise/tendigitprimes/database"
	"github.com/zalgonoise/tendigitprimes/log"
)

func BenchmarkPartitionSet(b *testing.B) {
	c := &config.Primes{
		LogLevel: "debug",
		Database: config.Database{
			URI:         "./testdata/parts",
			Partitioned: true,
		},
		Server: config.Server{
			HTTPPort: 8080,
			GRPCPort: 8081,
		},
	}

	logger := log.New(c.LogLevel)

	db, err := database.AttachSQLite(c.Database.URI, database.ReadOnlyPragmas(), logger)
	require.NoError(b, err)

	repo, err := NewPartitionSet(db)
	require.NoError(b, err)

	ctx := context.Background()
	logger.InfoContext(ctx, "service is ready")

	b.Run("Benchmark/Random", func(b *testing.B) {
		var n int64

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			n, err = repo.Random(ctx, 1_000_000_000, 5_000_000_000)
			if err != nil {
				b.Fatal(err)

				return
			}

			_ = n
		}
	})

	b.Run("Benchmark/List", func(b *testing.B) {
		var ns []int64

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			ns, err = repo.List(ctx, 1_000_000_000, 5_000_000_000, 10)
			if err != nil {
				b.Fatal(err)

				return
			}

			_ = ns
		}
	})

	require.NoError(b, repo.Close())
}
