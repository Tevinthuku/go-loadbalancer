package main

import (
	"load_balancer/backend"
	"log"
)

func main() {
	bs := backend.NewServer(":8081")
	if err := bs.Start(); err != nil {
		log.Fatal(err)
	}
}
