package main

import (
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/xgotyou/gossip/internal/client"
	"github.com/xgotyou/gossip/internal/server"
)

func main() {
	addr, ok := os.LookupEnv("HOSTPORT")
	if !ok {
		addr = "0.0.0.0:5051"
	}
	go server.Start(addr)

	neighbours, ok := os.LookupEnv("NEIGHBOURS")
	var nbs []string
	if ok {
		nbs = strings.Split(neighbours, " ")
	}

	pollIntStr, ok := os.LookupEnv("POLLINT")
	var pollInt time.Duration
	if ok {
		var err error
		pollInt, err = time.ParseDuration(pollIntStr)
		if err != nil {
			log.Fatalf("Can't parse POLLINT value: %v", err)
		}
	} else {
		pollInt = 5 * time.Second
	}

	go client.Manage(nbs, pollInt)

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c

	log.Println("Service terminated.")
}

// Example:
//   HOSTPORT="0.0.0.0:5051" NEIGHBOURS="0.0.0.0:5052 0.0.0.0:5053" go run main.go
//   HOSTPORT="0.0.0.0:5052" NEIGHBOURS="0.0.0.0:5051 0.0.0.0:5053" go run main.go
//   HOSTPORT="0.0.0.0:5053" NEIGHBOURS="0.0.0.0:5051 0.0.0.0:5052" go run main.go
