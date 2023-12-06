package main

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
)

const (
	checkTableExists = `SELECT EXISTS(SELECT 1 FROM sqlite_master 
	WHERE type='table' 
	AND name='primes');`

	createTableQuery = `
	CREATE TABLE primes (
    id INTEGER PRIMARY KEY,
    prime INTEGER
	);

	CREATE INDEX one_digit_primes ON primes (prime) 
	WHERE prime > 0 AND prime < 10;

	CREATE INDEX two_digit_primes ON primes (prime) 
	WHERE prime > 9 AND prime < 100;

	CREATE INDEX three_digit_primes ON primes (prime) 
	WHERE prime > 99 AND prime < 1000;

	CREATE INDEX four_digit_primes ON primes (prime) 
	WHERE prime > 999 AND prime < 10000;

	CREATE INDEX five_digit_primes ON primes (prime) 
	WHERE prime > 9999 AND prime < 100000;

	CREATE INDEX six_digit_primes ON primes (prime) 
	WHERE prime > 99999 AND prime < 1000000;

	CREATE INDEX seven_digit_primes ON primes (prime) 
	WHERE prime > 999999 AND prime < 10000000;

	CREATE INDEX eight_digit_primes ON primes (prime) 
	WHERE prime > 9999999 AND prime < 100000000;

	CREATE INDEX nine_digit_primes ON primes (prime) 
	WHERE prime > 99999999 AND prime < 1000000000;

	CREATE INDEX ten_digit_primes ON primes (prime) 
	WHERE prime > 999999999;
`

	insertValueQuery = `
INSERT INTO primes (prime) 
	VALUES (?);`
)

func initDatabase(ctx context.Context, logger *slog.Logger, db *sql.DB) error {
	r, err := db.QueryContext(ctx, checkTableExists)
	if err != nil {
		return err
	}

	defer r.Close()

	for r.Next() {
		var count int
		if err = r.Scan(&count); err != nil {
			return err
		}

		if count == 1 {
			logger.InfoContext(ctx, "primes table already exists in the database")

			return nil
		}
	}

	logger.InfoContext(ctx, "initializing primes table in the database")
	_, err = db.ExecContext(ctx, createTableQuery)
	if err != nil {
		return err
	}

	return nil
}

func insertValues(ctx context.Context, logger *slog.Logger, db *sql.DB, primes []int) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	logger.InfoContext(ctx, "preparing transaction", slog.Int("num_primes", len(primes)))

	for idx := range primes {
		if _, err = tx.ExecContext(ctx, insertValueQuery, primes[idx]); err != nil {
			return errors.Join(err, tx.Rollback())
		}
	}

	logger.InfoContext(ctx, "committing transaction")

	if err = tx.Commit(); err != nil {
		return errors.Join(err, tx.Rollback())
	}

	return nil
}
