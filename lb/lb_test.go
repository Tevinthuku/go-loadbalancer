package lb_test

import (
	"fmt"
	"io"
	"load_balancer/lb"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

func TestLoadBalancer(t *testing.T) {

	expected_exec_order := []int{0, 1}
	exec_order := []int{}
	b1 := &TestBackend{idx: 0, healthy: true, onHandleRequest: func() {
		exec_order = append(exec_order, 0)
	}}
	b2 := &TestBackend{idx: 1, healthy: true, onHandleRequest: func() {
		exec_order = append(exec_order, 1)
	}}
	lb := lb.NewLoadBalancer(":8080", []lb.Backend{b1, b2})

	fmt.Println("starting client")
	defer lb.Close()
	go lb.Start()

	client := http.Client{}
	req, err := http.NewRequest("GET", "http://localhost:8080/test", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	for i := 0; i < 2; i++ {
		_, err = client.Do(req)
		if err != nil {
			t.Fatalf("failed to send first request: %v", err)
		}

	}

	if !reflect.DeepEqual(exec_order, expected_exec_order) {
		t.Fatalf("expected execution order %v, got %v", expected_exec_order, exec_order)
	}
}

func TestUnhealthyBackendIsSkipped(t *testing.T) {
	expected_exec_order := []int{0, 2, 0, 1}
	exec_order := []int{}
	b1 := &TestBackend{idx: 0, healthy: true, onHandleRequest: func() {
		exec_order = append(exec_order, 0)
	}}
	b2 := &TestBackend{idx: 1, healthy: false, onHandleRequest: func() {
		exec_order = append(exec_order, 1)
	}}
	b3 := &TestBackend{idx: 2, healthy: true, onHandleRequest: func() {
		exec_order = append(exec_order, 2)
	}}
	lb := lb.NewLoadBalancer(":8080", []lb.Backend{b1, b2, b3})
	defer lb.Close()
	go lb.Start()

	client := http.Client{}
	for i := 0; i < 3; i++ {
		req, err := http.NewRequest("GET", "http://localhost:8080/test", nil)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		_, err = client.Do(req)
		if err != nil {
			t.Fatalf("failed to send request: %v", err)
		}
	}
	b2.healthy = true
	req, err := http.NewRequest("GET", "http://localhost:8080/test", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	_, err = client.Do(req)
	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}

	if !reflect.DeepEqual(exec_order, expected_exec_order) {
		t.Fatalf("expected execution order %v, got %v", expected_exec_order, exec_order)
	}
}

type TestBackend struct {
	idx             int
	healthy         bool
	onHandleRequest func()
}

func (b *TestBackend) HandleRequest(c *lb.RequestContext) {
	res := http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader("OK")),
	}
	c.ResponseChan <- &res
	b.onHandleRequest()
}

func (b *TestBackend) IsHealthy() bool {
	return b.healthy
}
