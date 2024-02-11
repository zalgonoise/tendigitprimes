package config

import "github.com/kelseyhightower/envconfig"

type Primes struct {
	LogLevel string `envconfig:"PRIMES_LOG_LEVEL" default:"info"`

	Database Database
	Server   Server
}

type Database struct {
	URI string `envconfig:"PRIMES_DB_URI"`
}

type Server struct {
	HTTPPort int `envconfig:"PRIMES_HTTP_PORT" default:"8080"`
	GRPCPort int `envconfig:"PRIMES_GRPC_PORT" default:"8081"`
}

func New() (*Primes, error) {
	var config Primes

	err := envconfig.Process("", &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
