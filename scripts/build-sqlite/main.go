package main

import (
	"context"
	"flag"
	"log/slog"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"github.com/zalgonoise/tendigitprimes/database"
	//_ "modernc.org/sqlite"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		AddSource:   true,
		Level:       nil,
		ReplaceAttr: nil,
	}))
	ctx := context.Background()

	code, err := run(logger)
	if err != nil {
		logger.ErrorContext(ctx, "runtime error", slog.String("error", err.Error()))
	}

	os.Exit(code)
}

func run(logger *slog.Logger) (int, error) {
	ctx := context.Background()
	input := flag.String("i", "./raw", "path to the input data to consume. Default is './input'")
	output := flag.String("o", "./sqlite/primes.db", "path to place the sqlite file in. Default is './sqlite/primes.db'")
	partition := flag.Bool("p", false, "partition database in multiple files")

	flag.Parse()

	if !*partition {
		logger.InfoContext(ctx, "validating output URI")
		db, err := database.OpenSQLite(*output, database.ReadWritePragmas(), logger)
		if err != nil {
			return 1, err
		}

		if err = database.MigrateSQLite(ctx, db, *input, logger); err != nil {
			return 1, err
		}

		if err = db.Close(); err != nil {
			return 1, err
		}

		return 0, nil
	}

	if err := database.Partition(ctx, 100_000_000, *input, *output, logger); err != nil {
		return 1, err
	}

	return 0, nil
}
