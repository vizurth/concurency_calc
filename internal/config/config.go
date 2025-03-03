package config

import (
	"github.com/vizurth/concurency_calc/internal/server/router"
	"os"
)

type Config struct {
	ServerConfig router.Config
}

func NewConfig() *Config {
	srvPort := os.Getenv("SERVER_PORT")
	if srvPort == "" {
		srvPort = "8080"
	}
	srvHost := os.Getenv("SERVER_HOST")
	if srvHost == "" {
		srvHost = "localhost"
	}
	return &Config{
		ServerConfig: router.Config{
			Host: srvHost,
			Port: srvPort,
		},
	}
}
