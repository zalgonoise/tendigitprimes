package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"

	_ "modernc.org/sqlite"
)

const (
	driver   = "sqlite"
	maxAlloc = 5_000_000

	uriFormat = "file:%s?_readonly=true&_txlock=immediate&cache=shared"
	inMemory  = ":memory:"

	applyPragma   = `PRAGMA %s;`
	applyPragmaKV = `PRAGMA %s = %s;`

	createTableQuery = `
	CREATE TABLE primes (
    prime INTEGER PRIMARY KEY NOT NULL 
	) STRICT;

	CREATE INDEX primes_0_to_1T ON primes (prime) WHERE prime < 1000000000;
	CREATE INDEX primes_1T_to_2T ON primes (prime) WHERE prime BETWEEN 1000000000 AND 1999999999;
	CREATE INDEX primes_2T_to_3T ON primes (prime) WHERE prime BETWEEN 2000000000 AND 2999999999;
	CREATE INDEX primes_3T_to_4T ON primes (prime) WHERE prime BETWEEN 3000000000 AND 3999999999;
	CREATE INDEX primes_4T_to_5T ON primes (prime) WHERE prime BETWEEN 4000000000 AND 4999999999;
	CREATE INDEX primes_5T_to_6T ON primes (prime) WHERE prime BETWEEN 5000000000 AND 5999999999;
	CREATE INDEX primes_6T_to_7T ON primes (prime) WHERE prime BETWEEN 6000000000 AND 6999999999;
	CREATE INDEX primes_7T_to_8T ON primes (prime) WHERE prime BETWEEN 7000000000 AND 7999999999;
	CREATE INDEX primes_8T_to_9T ON primes (prime) WHERE prime BETWEEN 8000000000 AND 8999999999;
	CREATE INDEX primes_9T_to_10T ON primes (prime) WHERE prime BETWEEN 9000000000 AND 9999999999;
`

	insertValueQuery = `
INSERT INTO primes (prime) 
	VALUES (?);`

	checkTableExists = `
SELECT EXISTS(SELECT 1 FROM sqlite_master 
	WHERE type='table' 
	AND name='%s');
`
)

func OpenSQLite(uri string, pragmas map[string]string, logger *slog.Logger) (*sql.DB, error) {
	switch uri {
	case inMemory:
	case "":
		uri = inMemory
	default:
		if err := validateURI(uri); err != nil {
			return nil, err
		}
	}

	if pragmas == nil {
		pragmas = ReadWritePragmas()
	}

	db, err := sql.Open(driver, fmt.Sprintf(uriFormat, uri))
	if err != nil {
		return nil, err
	}

	logger.Info("opened target DB", slog.String("uri", uri))

	if err := applyPragmas(context.Background(), db, pragmas); err != nil {
		return nil, err
	}

	logger.Info("prepared pragmas")

	return db, nil
}

func validateURI(uri string) error {
	stat, err := os.Stat(uri)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			f, err := os.Create(uri)
			if err != nil {
				return err
			}

			return f.Close()
		}

		return err
	}

	if stat.IsDir() {
		return fmt.Errorf("%s is a directory", uri)
	}

	return nil
}

func applyPragmas(ctx context.Context, db *sql.DB, pragmas map[string]string) (err error) {
	for k, v := range pragmas {
		switch v {
		case "":
			_, err = db.ExecContext(ctx, fmt.Sprintf(applyPragma, k))
		default:
			_, err = db.ExecContext(ctx, fmt.Sprintf(applyPragmaKV, k, v))
		}

		if err != nil {
			return err
		}
	}

	return nil
}
