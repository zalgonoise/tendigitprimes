package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"os"

	_ "modernc.org/sqlite"
)

const uriFormat = "file:%s?cache=shared"

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
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

	flag.Parse()

	logger.InfoContext(ctx, "validating output URI")
	if err := validateURI(ctx, logger, *output); err != nil {
		return 1, err
	}

	logger.InfoContext(ctx, "opening DB")
	db, err := sql.Open("sqlite", fmt.Sprintf(uriFormat, *output))
	if err != nil {
		return 1, err
	}

	defer func() {
		if err = db.Close(); err != nil {
			logger.WarnContext(ctx, "failed to close database", slog.String("error", err.Error()))
		}
	}()

	logger.InfoContext(ctx, "initializing database")
	if err = initDatabase(ctx, logger, db); err != nil {
		return 1, err
	}

	logger.InfoContext(ctx, "reading primes from input file(s)")
	primes, err := readPrimes(ctx, logger, *input)
	if err != nil {
		return 1, err
	}

	logger.InfoContext(ctx, "inserting primes into DB")
	if err = insertValues(ctx, logger, db, primes); err != nil {
		return 1, err
	}

	return 0, nil
}
