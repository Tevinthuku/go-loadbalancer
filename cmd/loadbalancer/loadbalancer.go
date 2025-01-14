package main

import (
	"load_balancer/lb"
	"log"
)

func main() {
	lb := lb.NewLoadBalancer(":8080")

	if err := lb.Start(); err != nil {
		log.Fatal(err)
	}

}
