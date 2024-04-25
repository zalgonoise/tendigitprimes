package main

import (
	"context"
	"log/slog"

	"github.com/zalgonoise/tendigitprimes/config"
	"github.com/zalgonoise/tendigitprimes/database"
)

func ExecBuild(ctx context.Context, logger *slog.Logger, args []string) (int, error) {
	c, err := config.NewBuild(args)
	if err != nil {
		return 1, err
	}

	if !c.Partitioned {
		logger.InfoContext(ctx, "validating output URI")
		db, err := database.OpenSQLite(c.Output, database.ReadWritePragmas(), logger)
		if err != nil {
			return 1, err
		}

		if err = database.MigrateSQLite(ctx, db, c.Input, logger); err != nil {
			return 1, err
		}

		if err = db.Close(); err != nil {
			return 1, err
		}

		return 0,

			nil
	}

	if err := database.Partition(ctx, int(c.BlockSize), c.Input, c.Output, logger); err != nil {
		return 1, err
	}

	return 0, nil
}
