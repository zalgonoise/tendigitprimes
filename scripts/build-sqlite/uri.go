package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
)

func validateURI(ctx context.Context, logger *slog.Logger, uri string) error {
	stat, err := os.Stat(uri)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			logger.InfoContext(ctx, "file does not exist; creating one",
				slog.String("path", uri),
				slog.String("error", err.Error()),
			)

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
