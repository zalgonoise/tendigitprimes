package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const uriFormat = "postgres://%s:%s@%s/%s"

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	code, err := run(logger)
	if err != nil {
		logger.ErrorContext(context.Background(), "runtime error", slog.String("error", err.Error()))
	}

	os.Exit(code)
}

func run(logger *slog.Logger) (int, error) {
	ctx := context.Background()

	input := flag.String("i", "./raw", "path to the input data to consume. Default is './input'")
	user := flag.String("u", "postgres", "postgres user. Default is 'postgres'")
	password := flag.String("p", "postgres", "postgres password. Default is 'postgres'")
	host := flag.String("h", "localhost:5432", "postgres host, with a port specification. Default is 'localhost:5432'")
	database := flag.String("d", "postgres", "postgres database name. Default is 'postgres'")

	flag.Parse()

	logger.InfoContext(ctx, "opening DB")
	db, err := sql.Open("pgx", fmt.Sprintf(uriFormat, *user, *password, *host, *database))
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
