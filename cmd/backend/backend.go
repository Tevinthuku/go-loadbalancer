package main

import (
	"fmt"
	"load_balancer/backend"
	"log"
	"os"
)

func main() {
	port := os.Getenv("BACKEND_PORT")
	if port == "" {
		log.Fatal("BACKEND_PORT is not set")
	}
	address := fmt.Sprintf(":%s", port)
	bs := backend.NewServer(address)
	if err := bs.Start(); err != nil {
		log.Fatal(err)
	}
}
