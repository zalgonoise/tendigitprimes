package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"os"

	_ "modernc.org/sqlite" // Database driver
)

const (
	uriFormat = "file:%s?cache=shared"
	inMemory  = ":memory:"
)

const (
	primesQuery = `
		SELECT prime FROM primes
			WHERE prime >= ? AND prime <= ?; 
`
)

type Repository struct {
	DB *sql.DB
}

func (r Repository) Get(ctx context.Context, min, max int64) (int64, error) {
	rows, err := r.DB.QueryContext(ctx, primesQuery, min, max)
	if err != nil {
		return 0, err
	}

	defer rows.Close()
	ns := make([]int64, 0, max-min)

	for rows.Next() {
		var n int64

		if err := rows.Scan(&n); err != nil {
			return 0, err
		}

		ns = append(ns, n)
	}

	return ns[rand.Intn(len(ns))], nil
}

func NewRepository(uri string) (Repository, error) {
	switch uri {
	case inMemory:
	case "":
		uri = inMemory
	default:
		if err := validateURI(uri); err != nil {
			return Repository{}, err
		}
	}

	db, err := sql.Open("sqlite", fmt.Sprintf(uriFormat, uri))
	if err != nil {
		return Repository{}, err
	}

	return Repository{db}, nil
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
