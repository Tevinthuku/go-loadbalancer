package main

import (
	"fmt"
	"load_balancer/lb"
	"log"
	"os"
	"strings"
)

func main() {

	lb_port := os.Getenv("LB_PORT")
	lb_address := fmt.Sprintf(":%s", lb_port)
	backend_ports := os.Getenv("BACKEND_PORTS")
	backend_addresses := []string{}
	for _, port := range strings.Split(backend_ports, ",") {
		backend_addresses = append(backend_addresses, fmt.Sprintf("http://localhost:%s", port))
	}
	lb := lb.NewLoadBalancer(lb_address, backend_addresses)

	if err := lb.Start(); err != nil {
		log.Fatal(err)
	}

}
