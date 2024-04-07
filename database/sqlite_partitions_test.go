package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zalgonoise/tendigitprimes/log"
)

func TestBlocks(t *testing.T) {
	logger := log.New("debug")

	data, err := readDataDir(context.Background(), "testdata/raw", logger)
	require.NoError(t, err)

	blocks := prepareBlocks(data[len(data)-1], 50_000_000)

	dataMap := mapBlocks(blocks, data)

	for i := range blocks {
		d := dataMap[blocks[i]]
		require.GreaterOrEqual(t, d[0], blocks[i].from)
		require.LessOrEqual(t, d[len(d)-1], blocks[i].to)
	}
}

func TestPrepareBlocks(t *testing.T) {
	for _, testcase := range []struct {
		name  string
		len   int
		size  int
		idLen int
	}{
		{
			name:  "4B/BlocksOf400M",
			len:   4_000_000_000,
			size:  400_000_000,
			idLen: 2,
		},
		{
			name:  "4B/BlocksOf40M",
			len:   4_000_000_000,
			size:  40_000_000,
			idLen: 2,
		},
		{
			name:  "4B/BlocksOf10M",
			len:   4_000_000_000,
			size:  10_000_000,
			idLen: 4,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			b := prepareBlocks(testcase.len, testcase.size)

			if len(b) > 0 {
				require.Len(t, b[0].id, testcase.idLen, "unexpected ID size", "id", b[0].id)

				last := b[len(b)-1].id

				if len(last) > 2 && last[0] == '0' && last[1] == '0' {
					t.Fatal("generating block IDs that are excessively large", "id:", last)
				}
			}

			for i := 0; i < len(b)-1; i++ {
				cur := b[i]
				next := b[i+1]

				require.Less(t, cur.to, next.from)
			}

		})
	}

}
