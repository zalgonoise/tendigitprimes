package config

import (
	"errors"
	"flag"
	"fmt"
	"strconv"
	"strings"

	"github.com/kelseyhightower/envconfig"
)

const minBlockSize = 100_000_000

var ErrBlockSizeTooLow = errors.New("block size value is too low")

type Build struct {
	Input       string    `envconfig:"PRIMES_BUILD_INPUT"`
	Output      string    `envconfig:"PRIMES_BUILD_OUTPUT"`
	Partitioned bool      `envconfig:"PRIMES_BUILD_IS_PARTITIONED"`
	BlockSize   BlockSize `envconfig:"PRIMES_BUILD_BLOCK_SIZE" `
}

type BlockSize int

func (b *BlockSize) Decode(value string) error {
	n, err := strconv.Atoi(strings.ReplaceAll(value, "_", ""))
	if err != nil {
		return err
	}

	if n < minBlockSize {
		return fmt.Errorf("%w: %d", ErrBlockSizeTooLow, n)
	}

	*b = BlockSize(n)

	return nil
}

func NewBuild(args []string) (*Build, error) {
	flagsConfig, err := flagsBuild(args)
	if err != nil {
		return nil, err
	}

	envConfig, err := envBuild()
	if err != nil {
		return nil, err
	}

	return applyBuildDefaults(mergeBuild(flagsConfig, envConfig)), nil
}

func flagsBuild(args []string) (*Build, error) {
	fs := flag.NewFlagSet("build", flag.ExitOnError)

	input := fs.String("input", "", "path to the input data to consume. Default is './input'")
	output := fs.String("output", "", "path to place the sqlite file in. Default is './sqlite/primes.db'")
	partitioned := fs.Bool("partitioned", false, "partition database in multiple files")
	blockSize := fs.Int("block-size", 0, "value range to set for each partition")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	config := &Build{}

	if *input != "" {
		config.Input = *input
	}

	if *output != "" {
		config.Output = *output
	}

	if *partitioned {
		config.Partitioned = *partitioned
	}

	if *blockSize >= minBlockSize {
		config.BlockSize = BlockSize(*blockSize)
	}

	return config, nil
}

func envBuild() (*Build, error) {
	config := &Build{}

	err := envconfig.Process("", config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func mergeBuild(base, next *Build) *Build {
	if next.Input != "" {
		base.Input = next.Input
	}

	if next.Output != "" {
		base.Output = next.Output
	}

	if next.Partitioned {
		base.Partitioned = true
	}

	return base
}

func applyBuildDefaults(config *Build) *Build {
	if config.Input == "" {
		config.Input = "./raw"
	}

	if config.Output == "" {
		config.Output = "./sqlite/primes.db"
	}

	if config.BlockSize < minBlockSize {
		config.BlockSize = minBlockSize
	}

	return config
}
