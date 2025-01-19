package main

import (
	"fmt"
	"load_balancer/lb"
	"log"
	"os"
	"strconv"
	"strings"
)

func main() {

	lb_port := os.Getenv("LB_PORT")
	if lb_port == "" {
		log.Fatal("LB_PORT is not set")
	}
	if _, err := strconv.Atoi(lb_port); err != nil {
		log.Fatal("LB_PORT must be a valid port number")
	}
	lb_address := fmt.Sprintf(":%s", lb_port)
	backend_ports := os.Getenv("BACKEND_PORTS")
	backend_addresses := []lb.Backend{}
	for _, port := range strings.Split(backend_ports, ",") {
		if _, err := strconv.Atoi(port); err != nil {
			log.Fatal("BACKEND_PORTS must be a valid port number")
		}
		backend_addresses = append(backend_addresses, lb.NewHttpBackend(fmt.Sprintf("http://localhost:%s", port)))
	}
	lb := lb.NewLoadBalancer(lb_address, backend_addresses)

	if err := lb.Start(); err != nil {
		log.Fatal(err)
	}
	defer lb.Close()

}
