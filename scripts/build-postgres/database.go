package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

const (
	checkTableExists = `SELECT EXISTS (
SELECT 1
FROM information_schema.tables
WHERE table_name = 'primes'
) AS table_existence;
`

	createTableQuery = `
	CREATE TABLE primes (
    prime BIGINT
	);
`
)

func initDatabase(ctx context.Context, logger *slog.Logger, db *sql.DB) error {
	r, err := db.QueryContext(ctx, checkTableExists)
	if err != nil {
		return err
	}

	defer r.Close()

	for r.Next() {
		var exists bool
		if err = r.Scan(&exists); err != nil {
			return err
		}

		if exists {
			logger.InfoContext(ctx, "primes table already exists in the database")

			return nil
		}
	}

	logger.InfoContext(ctx, "initializing primes table in the database")
	if _, err = db.ExecContext(ctx, createTableQuery); err != nil {
		return err
	}

	return nil
}

var errIncompatibleDriver = errors.New("the underlying driver value is not the expected driver.Conn interface")

func insertValues(ctx context.Context, logger *slog.Logger, db *sql.DB, primes [][]any) error {
	if len(primes) == 0 {
		return nil
	}

	logger.InfoContext(ctx, "creating a raw connection to the database")
	conn, err := db.Conn(ctx)
	if err != nil {
		return err
	}

	defer conn.Close()

	err = conn.Raw(func(driverConn any) error {
		logger.InfoContext(ctx, "getting standard-library connection")
		stdlibConn, ok := driverConn.(*stdlib.Conn)
		if !ok {
			return fmt.Errorf("%w: %T", errIncompatibleDriver, driverConn)
		}

		pgxConn := stdlibConn.Conn()

		logger.InfoContext(ctx, "starting copy-from bulk insert operation")
		_, err = pgxConn.CopyFrom(ctx, pgx.Identifier{"primes"}, []string{"prime"}, pgx.CopyFromRows(primes))

		return err
	})

	if err != nil {
		return err
	}

	logger.InfoContext(ctx, "insert operation completed successfully")

	return nil
}
