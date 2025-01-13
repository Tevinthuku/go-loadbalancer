package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
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

	for {
		conn, err := lb.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) error {
	connBytes, err := readFromConn(conn)
	if err != nil {
		return fmt.Errorf("failed to read from conn: %w", err)
	}
	defer conn.Close()

	backendConn, err := net.DialTimeout("tcp", ":8081", 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to dial backend conn: %w", err)
	}
	defer backendConn.Close()

	_, err = backendConn.Write(connBytes)
	if err != nil {
		return fmt.Errorf("failed to write to backend conn: %w", err)
	}
	backendResponse, err := readFromConn(backendConn)
	if err != nil {
		return fmt.Errorf("failed to read from backend conn: %w", err)
	}
	_, err = conn.Write(backendResponse)
	if err != nil {
		return fmt.Errorf("failed to write to conn: %w", err)
	}

	return nil
}

func readFromConn(conn net.Conn) ([]byte, error) {
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))

	reader := bufio.NewReader(conn)
	var buffer bytes.Buffer

	for {
		chunk := make([]byte, 1024)
		n, err := reader.Read(chunk)
		if n > 0 {
			buffer.Write(chunk[:n])
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
	}

	return buffer.Bytes(), nil
}
