package main

import (
	"github.com/vizurth/concurency_calc/internal/config"
	"github.com/vizurth/concurency_calc/internal/server/router"
)

func main() {
	cfg := config.NewConfig()
	r := router.NewRouter(cfg.ServerConfig)
	r.Start()
}
