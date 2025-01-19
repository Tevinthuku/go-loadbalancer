package lb

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
)

type LoadBalancer struct {
	port        string
	backends    []*Backend
	requestChan chan *RequestContext
}

func NewLoadBalancer(port string, backends []string) *LoadBalancer {
	lb := &LoadBalancer{port: port, requestChan: make(chan *RequestContext), backends: make([]*Backend, 0, len(backends))}
	for _, backend := range backends {
		lb.backends = append(lb.backends, newBackend(backend))
	}
	go lb.worker()
	return lb
}

func (lb *LoadBalancer) Start() error {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		lb.resolve(w, r)
	})

	return http.ListenAndServe(lb.port, nil)
}

func (lb *LoadBalancer) resolve(w http.ResponseWriter, req *http.Request) {
	responseChan := make(chan *http.Response, 1) // buffered channel to avoid goroutine leak
	lb.requestChan <- &RequestContext{request: req, responseChan: responseChan}
	response := <-responseChan
	if response != nil {
		if err := response.Write(w); err != nil {
			fmt.Println("failed to write response to client", err)
		}
		if response.Body != nil {
			response.Body.Close()
		}
	}
	close(responseChan)
}

func (lb *LoadBalancer) worker() {
	healthyBackendIdx := 0
	for req := range lb.requestChan {
		backend := lb.backends[healthyBackendIdx]
		if backend.isHealthy() {
			go backend.handleRequest(req)
			healthyBackendIdx = (healthyBackendIdx + 1) % len(lb.backends)
		} else {
			// check if the next backends are healthy or not
			healthyBackendFound := false
			// loop through all backends to find a healthy one
			for i := 1; i < len(lb.backends); i++ {
				absoluteIdx := (healthyBackendIdx + i) % len(lb.backends)
				if lb.backends[absoluteIdx].isHealthy() {
					healthyBackendFound = true
					go lb.backends[absoluteIdx].handleRequest(req)
					// we set this so that the next request will be sent to the next backend
					healthyBackendIdx = (absoluteIdx + 1) % len(lb.backends)
					break
				}
			}
			if !healthyBackendFound {
				response := http.Response{
					StatusCode: http.StatusServiceUnavailable,
					Body:       io.NopCloser(strings.NewReader("Service Unavailable")),
				}
				req.responseChan <- &response
			}
		}
	}
}

type Backend struct {
	address string
	health  atomic.Bool
	client  *http.Client
}

func newBackend(address string) *Backend {
	backend := &Backend{address: address, client: &http.Client{}}
	backend.health.Store(true)
	go backend.checkHealth()
	return backend
}

func (b *Backend) checkHealth() {
	for {
		resp, err := http.Get(fmt.Sprintf("%s/health", b.address))
		if err != nil {
			b.health.Store(false)
		} else if resp.StatusCode != http.StatusOK {
			b.health.Store(false)
		} else {
			b.health.Store(true)
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(2 * time.Second)
	}
}

func (b *Backend) handleRequest(c *RequestContext) {
	fmt.Println("backend ", b.address, " handling request")
	newReq, err := b.generateNewRequest(c.request)
	if err != nil {
		newResponse := http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       io.NopCloser(strings.NewReader(fmt.Sprintf("Internal Server Error: %s", err.Error()))),
		}
		c.responseChan <- &newResponse
		return
	}

	resp, err := b.client.Do(newReq)
	if err != nil {
		newResponse := http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       io.NopCloser(strings.NewReader(fmt.Sprintf("Internal Server Error: %s", err.Error()))),
		}
		c.responseChan <- &newResponse
		return
	}

	c.responseChan <- resp
}

func (b *Backend) generateNewRequest(req *http.Request) (*http.Request, error) {
	backendURL := fmt.Sprintf("%s%s?%s", b.address, req.URL.Path, req.URL.RawQuery)
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

func (b *Backend) isHealthy() bool {
	return b.health.Load()
}

type RequestContext struct {
	request      *http.Request
	responseChan chan *http.Response
}
