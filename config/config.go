package config

import (
	"github.com/caarlos0/env"
)

type Config struct {
	Service     ServiceConfig
	Server      ServerConfig
	Postgres    PostgresConfig
	OnDisk      OnDiskConfig
	UserService UserServiceConfig
	//AWS      AWSConfig
}

type ServerConfig struct {
	Host string `env:"HTTP_HOST,required"`
	Port string `env:"HTTP_PORT,required"`
}

type ServiceConfig struct {
}

type PostgresConfig struct {
	Addr string `env:"POSTGRES_URL,required"`
}

type OnDiskConfig struct {
	BaseURL       string `env:"ONDISK_BASE_URL" envDefault:"http://localhost:80/csv/"`
	DirectoryPath string `evn:"ONDISK_DIRECTORY_PATH" envDefault:"csv/"`
}

type UserServiceConfig struct {
	BaseURL string `env:"USER_SERVICE_BASE_URL,required"`
}

func Parse() (*Config, error) {
	cfg := Config{}

	if err := env.Parse(&cfg.Server); err != nil {
		return nil, err
	}

	if err := env.Parse(&cfg.Service); err != nil {
		return nil, err
	}

	if err := env.Parse(&cfg.Postgres); err != nil {
		return nil, err
	}

	if err := env.Parse(&cfg.OnDisk); err != nil {
		return nil, err
	}

	if err := env.Parse(&cfg.UserService); err != nil {
		return nil, err
	}

	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
