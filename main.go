package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/xgotyou/gossip/internal/client"
	"github.com/xgotyou/gossip/internal/server"
)

type config struct {
	Addr       string        `env:"HOSTPORT" env-default:"0.0.0.0:5051"`
	Neighbours []string      `env:"NEIGHBOURS"`
	PollInt    time.Duration `env:"POLLINT" env-default:"5s"`
}

func main() {
	// Load configuration
	var cfg config
	err := cleanenv.ReadEnv(&cfg)
	if err != nil {
		log.Fatalf("Error reading configuration: %s", err)
	}

	// Start gPRC server and manage connections with other services
	ctx, cancel := context.WithCancel(context.Background())
	go server.Start(cfg.Addr)
	go client.Manage(ctx, cfg.Neighbours, cfg.PollInt)
	defer cancel()

	// Wait for external interuptino signal and shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c

	log.Println("Service terminated.")
}

// Example:
//   HOSTPORT="0.0.0.0:5051" NEIGHBOURS="0.0.0.0:5052,0.0.0.0:5053" go run main.go
//   HOSTPORT="0.0.0.0:5052" NEIGHBOURS="0.0.0.0:5051,0.0.0.0:5053" go run main.go
//   HOSTPORT="0.0.0.0:5053" NEIGHBOURS="0.0.0.0:5051,0.0.0.0:5052" go run main.go
