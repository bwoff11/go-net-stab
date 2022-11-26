package main

import (
	"log"

	"github.com/bwoff11/go-net-stab/internal/config"
	"github.com/bwoff11/go-net-stab/internal/listener"
	"github.com/bwoff11/go-net-stab/internal/registry"
	"github.com/bwoff11/go-net-stab/internal/reporting"
)

func main() {
	if err := config.LoadConfig(); err != nil {
		log.Fatal("Failed to load config:", err)
	}
	if err := listener.Start(); err != nil {
		log.Fatal("Failed to start listener:", err)
	}

	registry := registry.Create()
	for _, endpoint := range config.Config.Endpoints {
		registry.AddEndpoint(endpoint)
	}
	registry.Run()

	reporting.ServeMetrics()
}
