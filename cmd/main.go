package main

import (
	"fmt"
	"log"

	"github.com/dielit66/cloud-camp-tt/internal/config"
)

func main() {
	cfg, err := config.LoadConfig("config/config.yaml")

	if err != nil {
		log.Fatalf("Error while reading config, err: %v", err)
	}

	fmt.Printf("port from config: %s", cfg.Port)
}
