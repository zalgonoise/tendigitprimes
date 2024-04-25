package config

import (
	"flag"
	"strings"

	"github.com/kelseyhightower/envconfig"
)

type Primes struct {
	LogLevel string `envconfig:"PRIMES_LOG_LEVEL"`

	Database Database
	Server   Server
}

type Database struct {
	URI         string `envconfig:"PRIMES_DB_URI"`
	Partitioned bool   `envconfig:"PRIMES_DB_IS_PARTITIONED"`
}

type Server struct {
	HTTPPort int `envconfig:"PRIMES_HTTP_PORT"`
	GRPCPort int `envconfig:"PRIMES_GRPC_PORT"`
}

func NewPrimes(args []string) (*Primes, error) {
	flagsConfig, err := flagsPrimes(args)
	if err != nil {
		return nil, err
	}

	envConfig, err := envPrimes()
	if err != nil {
		return nil, err
	}

	return applyPrimesDefaults(mergePrimes(flagsConfig, envConfig)), nil
}

func flagsPrimes(args []string) (*Primes, error) {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)

	logLevel := fs.String("log-level", "", "defines the minimum log level when registering events [one of: 'debug', 'info', 'warn', 'error']")

	dbURI := fs.String("db.uri", "", "the URI for the database file or partitions directory")
	dbIsPartitioned := fs.Bool("db.partitioned", false, "setup SQLite with partitioned database files")

	serverHTTPPort := fs.Int("server.http-port", 0, "web server's HTTP port")
	serverGRPCPort := fs.Int("server.grpc-port", 0, "web server's gRPC port")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	config := &Primes{}

	l := strings.ToLower(*logLevel)

	switch l {
	case "debug", "info", "warn", "error":
		config.LogLevel = l
	}

	if *dbURI != "" {
		config.Database.URI = *dbURI
	}

	if *dbIsPartitioned {
		config.Database.Partitioned = true
	}

	if *serverHTTPPort > 0 {
		config.Server.HTTPPort = *serverHTTPPort
	}

	if *serverGRPCPort > 0 {
		config.Server.GRPCPort = *serverGRPCPort
	}

	return config, nil
}

func envPrimes() (*Primes, error) {
	config := &Primes{}

	err := envconfig.Process("", config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func mergePrimes(base, next *Primes) *Primes {
	if next.LogLevel != "" {
		base.LogLevel = next.LogLevel
	}

	// Database configuration (path + is_partitioned) is coupled tightly
	if next.Database.URI != "" {
		base.Database = next.Database
	}

	if next.Server.HTTPPort > 0 {
		base.Server.HTTPPort = next.Server.HTTPPort
	}

	if next.Server.GRPCPort > 0 {
		base.Server.GRPCPort = next.Server.GRPCPort
	}

	return base
}

func applyPrimesDefaults(config *Primes) *Primes {
	if config.LogLevel == "" {
		config.LogLevel = "info"
	}

	if config.Server.HTTPPort == 0 {
		config.Server.HTTPPort = 8080
	}

	if config.Server.GRPCPort == 0 {
		config.Server.GRPCPort = 8081
	}

	return config
}
