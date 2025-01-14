package lb

import (
	"fmt"
	"net/http"
	"time"
)

type LoadBalancer struct {
	port string
}

func NewLoadBalancer(port string) *LoadBalancer {
	return &LoadBalancer{port: port}
}

func (lb *LoadBalancer) Start() error {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if err := lb.handleHTTPRequest(w, r); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	return http.ListenAndServe(lb.port, nil)
}

func (lb *LoadBalancer) handleHTTPRequest(w http.ResponseWriter, req *http.Request) error {

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	newReq, err := newRequestToForwardToBackend(req)
	if err != nil {
		return fmt.Errorf("failed to create backend request: %w", err)
	}

	resp, err := client.Do(newReq)
	if err != nil {
		return fmt.Errorf("failed to forward request to backend: %w", err)
	}
	defer resp.Body.Close()

	err = resp.Write(w)
	if err != nil {
		return fmt.Errorf("failed to write response to client: %w", err)
	}

	return nil
}

func newRequestToForwardToBackend(req *http.Request) (*http.Request, error) {
	backendURL := fmt.Sprintf("http://localhost:8081%s", req.URL.Path)
	newReq, err := http.NewRequest(req.Method, backendURL, req.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to create backend request: %w", err)
	}

	// Copy headers from original request
	for key, values := range req.Header {
		for _, value := range values {
			newReq.Header.Add(key, value)
		}
	}
	return newReq, nil
}
