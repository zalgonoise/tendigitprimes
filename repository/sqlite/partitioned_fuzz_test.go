package sqlite

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zalgonoise/tendigitprimes/config"
	"github.com/zalgonoise/tendigitprimes/database"
	"github.com/zalgonoise/tendigitprimes/log"
)

func FuzzPartitionSet_Random(f *testing.F) {
	f.Add(int64(1_000_000_000), int64(5_000_000_000))

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
	require.NoError(f, err)

	repo, err := NewPartitionSet(db)
	require.NoError(f, err)

	logger.InfoContext(context.Background(), "service is ready")

	f.Fuzz(func(t *testing.T, min int64, max int64) {
		if min < 2 || max > 9_999_999_999 {
			return
		}

		n, err := repo.Random(context.Background(), min, max)
		if err != nil {
			t.Fatal(err)
		}

		if n < 2 {
			t.Fatal("number cannot be lower than two")
		}

		if n > 9_999_999_999 {
			t.Fatal("number cannot be over ten-digits-long")
		}
	})
}
