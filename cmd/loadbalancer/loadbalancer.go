package main

import (
	"load_balancer/lb"
	"log"
)

func main() {
	lb := lb.NewLoadBalancer(":8080", []string{"http://localhost:8081", "http://localhost:8082", "http://localhost:8083"})

	if err := lb.Start(); err != nil {
		log.Fatal(err)
	}

}
