package sqlite

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestContains(t *testing.T) {
	for _, testcase := range []struct {
		name      string
		part      partition
		min       int64
		max       int64
		isPresent bool
		isOver    bool
	}{
		{
			name: "WithinBounds/OnStart",
			part: partition{
				from:  1_000_000_000,
				to:    1_000_999_999,
				total: 1_000_000,
				id:    "0a",
			},
			min:       1_000_000_000,
			max:       5_000_000_000,
			isPresent: true,
		},
		{
			name: "WithinBounds/OnMiddle",
			part: partition{
				from:  3_000_000_000,
				to:    3_000_999_999,
				total: 1_000_000,
				id:    "0a",
			},
			min:       1_000_000_000,
			max:       5_000_000_000,
			isPresent: true,
		},
		{
			name: "OutOfBounds/OnEnd",
			part: partition{
				from:  5_000_000_000,
				to:    5_000_999_999,
				total: 1_000_000,
				id:    "0a",
			},
			min:       1_000_000_000,
			max:       5_000_000_000,
			isPresent: false,
			isOver:    true,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			isPresent, isOver := contains(testcase.part, testcase.min, testcase.max)
			require.Equal(t, testcase.isPresent, isPresent)
			require.Equal(t, testcase.isOver, isOver)
		})
	}
}
