package main

import (
	"bufio"
	"fmt"
	"load_balancer/lb"
	"log"
	"net"
	"net/http"
	"net/url"
	"time"
)

func main() {
	lb, err := lb.NewLoadBalancer(":8080")
	if err != nil {
		log.Fatal(err)
	}
	defer lb.Close()

	for {
		conn, err := lb.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go handleHTTPConnection(conn)
	}
}

func handleHTTPConnection(conn net.Conn) error {

	defer conn.Close()

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.ReadRequest(bufio.NewReader(conn))
	if err != nil {
		return fmt.Errorf("failed to read request: %w", err)
	}
	defer req.Body.Close()

	// forward request to backend
	backendURL := fmt.Sprintf("http://localhost:8081%s", req.URL.Path)
	req.URL, err = url.Parse(backendURL)
	if err != nil {
		return fmt.Errorf("failed to parse backend URL: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to forward request to backend: %w", err)
	}
	defer resp.Body.Close()

	// forward response to client
	err = resp.Write(conn)
	if err != nil {
		return fmt.Errorf("failed to write response to client: %w", err)
	}

	return nil
}
