package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/rand/v2"
	"time"
)

const (
	minAlloc       = 64
	defaultTimeout = time.Minute

	querySelectScopes = `SELECT id, min, max, total FROM scopes;`

	primesPartitionedQuery = `SELECT prime FROM db%s.primes AS p
	LIMIT 1 OFFSET %d;`
)

type partition struct {
	from  int64
	to    int64
	total int64
	id    string
}

type PartitionSet struct {
	parts []partition

	DB   *sql.DB
	Conn *sql.Conn
}

func (r *PartitionSet) Random(ctx context.Context, min, max int64) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	targets := scanPartitions(r.parts, min, max)

	t := targets[rand.IntN(len(targets))]

	var n int64
	var err error

	for {
		n, err = randomPrime(ctx, r.Conn, t)
		if err != nil {
			return 0, err
		}

		if n >= min && n <= max {
			return n, nil
		}
	}
}

func (r *PartitionSet) List(ctx context.Context, min, max, limit int64) ([]int64, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	if limit == 0 {
		limit = defaultLimit
	}

	targets := scanPartitions(r.parts, min, max)

	return listRandomPrimes(ctx, r.Conn, targets, min, max, int(limit))
}

func (r *PartitionSet) Close() error {
	return errors.Join(r.Conn.Close(), r.DB.Close())
}

func scanPartitions(parts []partition, min, max int64) []partition {
	targets := make([]partition, 0, len(parts))

	for i := range parts {
		isPresent, isOver := contains(parts[i], min, max)
		if isPresent {
			targets = append(targets, parts[i])
		}

		if isOver {
			break
		}
	}

	return targets
}

func listRandomPrimes(ctx context.Context, conn *sql.Conn, targets []partition, min, max int64, limit int) ([]int64, error) {
	results := make([]int64, 0, limit)

	var idx int

	for len(results) < limit {
		n, err := randomPrime(ctx, conn, targets[idx])
		if err != nil {
			return nil, err
		}

		if n >= min && n <= max {
			results = append(results, n)
		}

		idx = (idx + rand.IntN(len(targets))) % len(targets)
	}

	return results, nil
}

func randomPrime(ctx context.Context, conn *sql.Conn, target partition) (int64, error) {
	offset := rand.Int64N(target.total - 1)

	query := fmt.Sprintf(primesPartitionedQuery, target.id, offset)

	row := conn.QueryRowContext(ctx, query)

	var n int64

	if err := row.Scan(&n); err != nil {
		return 0, err
	}

	if err := row.Err(); err != nil {
		return 0, err
	}

	return n, nil
}

func NewPartitionSet(db *sql.DB, conn *sql.Conn) (*PartitionSet, error) {
	parts, err := getPartitions(conn)
	if err != nil {
		return nil, err
	}

	return &PartitionSet{parts: parts, DB: db, Conn: conn}, nil
}

func getPartitions(conn *sql.Conn) ([]partition, error) {
	ctx := context.Background()

	rows, err := conn.QueryContext(ctx, querySelectScopes)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	parts := make([]partition, 0, minAlloc)

	for rows.Next() {
		var (
			id       string
			from, to int64
			total    int64
		)

		if err = rows.Scan(&id, &from, &to, &total); err != nil {
			return nil, err
		}

		parts = append(parts, partition{
			from:  from,
			to:    to,
			total: total,
			id:    id,
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

func contains(part partition, min, max int64) (isPresent bool, isOver bool) {
	switch {
	case part.from > max:
		return false, true
	case part.to < min:
		return false, false
	default:
		return true, false
	}
}
