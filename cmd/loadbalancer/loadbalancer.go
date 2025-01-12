package main

import (
	"fmt"
	"load_balancer/lb"
	"log"
	"net"
	"time"
)

func main() {
	lb, err := lb.NewLoadBalancer(":8080")
	if err != nil {
		log.Fatal(err)
	}
	defer lb.Close()

	// get connection to backend server
	clientConn, err := net.Dial("tcp", ":8081")
	if err != nil {
		log.Fatal(err)
	}
	defer clientConn.Close()
	for {
		conn, err := lb.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go handleConnection(conn, clientConn)
	}
}

func handleConnection(conn net.Conn, backendConn net.Conn) error {
	connBytes, err := readFromConn(conn)
	defer conn.Close()
	if err != nil {
		return fmt.Errorf("failed to read from conn: %w", err)
	}
	_, err = backendConn.Write(connBytes)
	if err != nil {
		return fmt.Errorf("failed to write to backend conn: %w", err)
	}
	backendConnBytes, err := readFromConn(backendConn)
	if err != nil {
		return fmt.Errorf("failed to read from backend conn: %w", err)
	}
	_, err = conn.Write(backendConnBytes)
	if err != nil {
		return fmt.Errorf("failed to write to conn: %w", err)
	}

	return nil
}

func readFromConn(conn net.Conn) ([]byte, error) {
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return nil, err
	}
	return buf[:n], nil
}
