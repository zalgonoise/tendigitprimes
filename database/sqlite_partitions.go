package database

import (
	"context"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log/slog"

	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

const (
	minAlloc              = 64
	sqliteAttachHardLimit = 125

	pathBlock = "/blk_"

	queryPartitionIDs = `SELECT id FROM scopes;`

	queryAttachDB = `ATTACH DATABASE '%s%s%s.db' AS db%s;`
)

type block struct {
	from int
	to   int
	id   string
}

// AttachSQLite opens a connection to 'index.db' under dir, and queries for all IDs registered in it; which are used to
// attach the database partitions under dir.
//
// Note that SQLite sets a maximum number of 10 attached databases by default, but can be modified as a compile-time
// option, up to a maximum of 125 databases.
//
// In the context of modernc.org/sqlite, one can use Go workspaces to modify a local copy of the repository, where the
// appropriate configuration is found:
//
//	const SQLITE_MAX_ATTACHED = 100
//
// Then, it is possible to attach a hundred SQLite databases on the same index, making the partitions usable. The hard
// limit is 125 databases: https://www.sqlite.org/limits.html#max_attached
func AttachSQLite(dir string, pragmas map[string]string, logger *slog.Logger) (*sql.DB, *sql.Conn, error) {
	ctx := context.Background()

	db, err := OpenSQLite(dir+"/index.db", pragmas, logger)
	if err != nil {
		return nil, nil, err
	}

	ids, err := getIDs(ctx, db)
	if err != nil {
		return nil, nil, err
	}

	conn, err := attachDBs(ctx, db, dir, ids)
	if err != nil {
		return nil, nil, err
	}

	return db, conn, nil
}

func attachDBs(ctx context.Context, db *sql.DB, dir string, ids []string) (*sql.Conn, error) {
	conn, err := db.Conn(ctx)
	if err != nil {
		return nil, err
	}

	if _, err = sqlite.Limit(conn, sqlite3.SQLITE_LIMIT_ATTACHED, len(ids)); err != nil {
		return nil, err
	}

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	for i := range ids {
		if _, err := tx.ExecContext(ctx, fmt.Sprintf(queryAttachDB, dir, pathBlock, ids[i], ids[i])); err != nil {
			return nil, err
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return conn, nil
}

func getIDs(ctx context.Context, db *sql.DB) ([]string, error) {
	rows, err := db.QueryContext(ctx, queryPartitionIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ids := make([]string, 0, minAlloc)

	for rows.Next() {
		var id string

		if err = rows.Scan(&id); err != nil {
			return nil, err
		}

		ids = append(ids, id)
	}

	if err = rows.Close(); err != nil {
		return nil, err
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return ids, nil
}

// Partition consumes the data in input, and creates partitioned SQLite databases in dir, with a given blockSize size.
//
// Different blockSize values result in different number of partitions. Note that SQLite has configured a maximum of 10
// attached databases at a time by default. In order to have access to the maximum amount of 125 databases, SQLite must
// be compiled with this maximum.
//
// In the context of modernc.org/sqlite, one can use Go workspaces to modify a local copy of the repository, where the
// appropriate configuration is found:
//
//	const SQLITE_MAX_ATTACHED = 100
//
// Then, it is possible to attach a hundred SQLite databases on the same index, making the partitions usable. The hard
// limit is 125 databases: https://www.sqlite.org/limits.html#max_attached
func Partition(ctx context.Context, blockSize int, input, dir string, logger *slog.Logger) error {
	data, err := readDataDir(ctx, input, logger)
	if err != nil {
		return err
	}

	return partitionData(ctx, data, blockSize, dir, logger)
}

func partitionData(ctx context.Context, data []int, blockSize int, path string, logger *slog.Logger) error {
	idxDB, err := OpenSQLite(path+"/index.db", ReadWritePragmas(), logger)
	if err != nil {
		return err
	}

	if err = runMigrations(ctx, idxDB,
		migration{table: "scopes", create: createScopesTableQuery},
	); err != nil {
		return err
	}

	blocks := prepareBlocks(data[len(data)-1], blockSize)

	if len(blocks) > sqliteAttachHardLimit {
		return fmt.Errorf("number of generated partitions is over the SQLite limit for attaching databases (%d): len: %d", sqliteAttachHardLimit, len(blocks))
	}

	dataMap := mapBlocks(blocks, data)

	for i := range blocks {
		db, err := OpenSQLite(path+pathBlock+blocks[i].id+".db", ReadWritePragmas(), logger)
		if err != nil {
			return err
		}

		if err = runMigrations(ctx, db,
			migration{table: "primes", create: createTableQuery},
		); err != nil {
			return err
		}

		if err = insertData(ctx, db, dataMap[blocks[i]], minBlockSize, logger); err != nil {
			return err
		}

		if err = db.Close(); err != nil {
			return err
		}

		if _, err = idxDB.ExecContext(ctx, insertScopesQuery, blocks[i].id, blocks[i].from, blocks[i].to, len(dataMap[blocks[i]])); err != nil {
			return err
		}
	}

	return idxDB.Close()
}

func mapBlocks(blocks []block, data []int) map[block][]int {
	// data should already be sorted
	var lastIdx int

	dataMap := make(map[block][]int, len(blocks))
	for i := range blocks {
		n, d := mapBlock(blocks[i], data[lastIdx:])

		lastIdx += n

		dataMap[blocks[i]] = d
	}

	return dataMap
}

func mapBlock(b block, data []int) (int, []int) {
	d := make([]int, 0, len(data))

	for i := range data {
		switch {
		case data[i] <= b.to:
			d = append(d, data[i])
		default:
			return i, d
		}
	}

	return -1, d
}

func prepareBlocks(maximum, blockSize int) []block {
	blocks := (maximum / blockSize) + 1

	values := make([]block, 0, blocks)

	id := make([]byte, (blocks/255)+1)

	for i := 0; i < maximum; i += blockSize {
		to := i + blockSize - 1
		if to > maximum {
			to = maximum
		}

		values = append(values, block{
			from: i,
			to:   to,
			id:   hex.EncodeToString(id),
		})

		incID(id)
	}

	return values
}

func incID(id []byte) {
	for i := len(id) - 1; i >= 0; i-- {
		id[i] = (id[i] + 1) % 255

		if id[i] != 0 {
			return
		}
	}
}
