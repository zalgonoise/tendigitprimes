package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand/v2"
)

const (
	minAlloc = 64

	querySelectScopes = `
	SELECT id, from, to FROM scopes;
`

	primesPartitionedQuery = `
		SELECT prime FROM db%s.primes
			WHERE prime BETWEEN ? AND ?
`

	primesPartitionedLimitQuery = `
		SELECT prime FROM db%s.primes
			WHERE prime BETWEEN ? AND ?
			LIMIT %d
`
)

type partition struct {
	from int64
	to   int64
	id   string
}

type PartitionSet struct {
	parts []partition

	DB   *sql.DB
	Conn *sql.Conn
}

func (r *PartitionSet) Random(ctx context.Context, min, max int64) (int64, error) {
	targets := make([]partition, 0, len(r.parts))

	for i := range r.parts {
		if contains(r.parts[i], min, max) {
			targets = append(targets, r.parts[i])
		}
	}

	t := targets[rand.IntN(len(targets))]

	rows, err := r.Conn.QueryContext(ctx, fmt.Sprintf(primesPartitionedQuery, t.id), min, max)
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

	return ns[rand.IntN(len(ns))], nil
}

func (r *PartitionSet) List(ctx context.Context, min, max, limit int64) ([]int64, error) {
	if limit == 0 {
		limit = defaultLimit
	}

	targets := make([]partition, 0, len(r.parts))

	for i := range r.parts {
		if contains(r.parts[i], min, max) {
			targets = append(targets, r.parts[i])
		}
	}

	return listPrimes(ctx, r.Conn, targets, min, max, limit)
}

func listPrimes(ctx context.Context, db *sql.Conn, targets []partition, min, max, limit int64) ([]int64, error) {
	results := make([]int64, 0, limit)

	for i := range targets {
		rows, err := db.QueryContext(ctx, fmt.Sprintf(primesPartitionedLimitQuery, targets[i].id, limit), min, max)
		if err != nil {
			return nil, err
		}

		defer rows.Close()

		for rows.Next() {
			var n int64

			if err = rows.Scan(&n); err != nil {
				return nil, err
			}

			results = append(results, n)

			if len(results) == int(limit) {
				return results, nil
			}
		}

		if err = rows.Close(); err != nil {
			return nil, err
		}

		if err = rows.Err(); err != nil {
			return nil, err
		}
	}

	return results, nil
}

func NewPartitionSet(db *sql.DB, conn *sql.Conn) (*PartitionSet, error) {
	parts, err := getPartitions(conn)
	if err != nil {
		return nil, err
	}

	return &PartitionSet{parts: parts, DB: db, Conn: conn}, nil
}

func getPartitions(db *sql.Conn) ([]partition, error) {
	rows, err := db.QueryContext(context.Background(), querySelectScopes)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	parts := make([]partition, 0, minAlloc)

	for rows.Next() {
		var (
			id       string
			from, to int64
		)

		if err = rows.Scan(&id, &from, &to); err != nil {
			return nil, err
		}

		parts = append(parts, partition{
			from: from,
			to:   to,
			id:   id,
		})
	}

	if err = rows.Close(); err != nil {
		return nil, err
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return parts, nil
}

func contains(part partition, min, max int64) bool {
	switch {
	case min >= part.to, max <= part.from:
		return false
	default:
		return true
	}
}
