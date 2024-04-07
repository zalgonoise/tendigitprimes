package database

import (
	"context"
	"encoding/hex"
	"log/slog"
)

type block struct {
	from int
	to   int
	id   string
}

func Partition(ctx context.Context, blockSize int, input, dir string, logger *slog.Logger) error {
	data, err := readDataDir(ctx, input, logger)
	if err != nil {
		return err
	}

	return partitionData(ctx, data, blockSize, dir, logger)
}

func partitionData(ctx context.Context, data []int, blockSize int, path string, logger *slog.Logger) error {
	blocks := prepareBlocks(len(data), blockSize)

	// data should already be sorted
	var lastIdx int

	dataMap := make(map[block][]int, len(blocks))
	for i := range blocks {
		d := make([]int, 0, len(data))
		b := blocks[i]

		for idx := lastIdx; idx < len(data); idx++ {
			if data[idx] >= b.from && data[idx] <= b.to {
				d = append(d, data[idx])
			}

			if data[idx] >= b.to {
				lastIdx = idx

				break
			}
		}

		dataMap[blocks[i]] = d
	}

	for i := range blocks {
		db, err := OpenSQLite(path+"/blk_"+blocks[i].id+".db", ReadWritePragmas(), logger)
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
	}

	return nil
}

func prepareBlocks(length int, blockSize int) []block {
	blocks := (length / blockSize) + 1

	values := make([]block, 0, blocks)

	id := make([]byte, (blocks/255)+1)

	for i := 0; i < length; i += blockSize {
		to := i + blockSize - 1
		if to > length {
			to = length
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
