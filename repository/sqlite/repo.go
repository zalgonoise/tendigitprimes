package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	//_ "github.com/mattn/go-sqlite3"
	//"modernc.org/sqlite"
	_ "modernc.org/sqlite" // Database driver
)

const (
	defaultLimit = 5000

	primesQuery = `
		SELECT prime FROM primes
			WHERE prime BETWEEN ? AND ?
`
	primesLimitQuery = `
		SELECT prime FROM primes
			WHERE prime BETWEEN ? AND ?
			LIMIT %d
`
)

type Repository struct {
	DB *sql.DB
}

func (r Repository) Random(ctx context.Context, min, max int64) (int64, error) {
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

func (r Repository) List(ctx context.Context, min, max, limit int64) ([]int64, error) {
	if limit == 0 {
		limit = defaultLimit
	}

	rows, err := r.DB.QueryContext(ctx, fmt.Sprintf(primesLimitQuery, limit), min, max)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	ns := make([]int64, 0, max-min)

	for rows.Next() {
		var n int64

		if err := rows.Scan(&n); err != nil {
			return nil, err
		}

		ns = append(ns, n)
	}

	if err = rows.Close(); err != nil {
		return nil, err
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return ns, nil
}

func (r Repository) Close() error {
	return r.DB.Close()
}

func NewRepository(db *sql.DB) (Repository, error) {
	return Repository{db}, nil
}
