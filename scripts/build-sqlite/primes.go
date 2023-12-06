package main

import (
	"bufio"
	"context"
	"errors"
	"log/slog"
	"os"
	"path"
	"strconv"
)

const maxAlloc = 5_000_000

func readPrimes(ctx context.Context, logger *slog.Logger, dir string) ([]int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	logger.InfoContext(ctx, "scanned directory", slog.Int("num_files", len(entries)))

	total := make([]int, 0, len(entries)*maxAlloc)
	errs := make([]error, 0, len(entries))

	for i := range entries {
		logger.InfoContext(ctx, "extracting primes from file", slog.String("filename", entries[i].Name()))

		data, extractErr := extract(ctx, logger, path.Join(dir, entries[i].Name()))
		if extractErr != nil {
			errs = append(errs, err)

			continue
		}

		total = append(total, data...)
	}

	logger.InfoContext(ctx, "extracted primes from input file(s)",
		slog.Int("num_primes", len(total)),
		slog.Int("num_errors", len(errs)),
	)

	return total, errors.Join(errs...)
}

func extract(ctx context.Context, logger *slog.Logger, path string) ([]int, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			logger.WarnContext(ctx, "error closing file",
				slog.String("error", err.Error()),
				slog.String("filename", file.Name()),
			)
		}
	}()

	scanner := bufio.NewScanner(file)

	values := make([]int, 0, maxAlloc)
	errs := make([]error, 0, maxAlloc)

	for scanner.Scan() {
		value, convErr := strconv.Atoi(scanner.Text())
		if convErr != nil {
			errs = append(errs, convErr)

			continue
		}

		values = append(values, value)
	}

	return values, errors.Join(errs...)
}
