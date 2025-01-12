package main

import (
	"load_balancer/backend"
)

func main() {
	bs := backend.NewServer(":8081")
	bs.Start()
}
