package lb

import (
	"fmt"
	"net/http"
	"time"
)

type LoadBalancer struct {
	port        string
	reqResolver *RequestResolver
}

func NewLoadBalancer(port string, backends []string) *LoadBalancer {
	lb := &LoadBalancer{port: port}
	lb.reqResolver = newRequestResolver(backends)
	return lb
}

func (lb *LoadBalancer) Start() error {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		lb.reqResolver.resolve(w, r)
	})

	return http.ListenAndServe(lb.port, nil)
}

type RequestResolver struct {
	backends    []*Backend
	requestChan chan *RequestWithResponseWriter
}

func newRequestResolver(backends []string) *RequestResolver {
	resolver := &RequestResolver{}
	for _, backend := range backends {
		resolver.backends = append(resolver.backends, newBackend(backend))
	}
	go resolver.worker()
	return resolver
}

func (rr *RequestResolver) resolve(w http.ResponseWriter, req *http.Request) {
	rr.requestChan <- &RequestWithResponseWriter{request: req, writer: w}
}

func (rr *RequestResolver) worker() {
	healthyBackendIdx := 0
	for req := range rr.requestChan {
		backend := rr.backends[healthyBackendIdx]
		go backend.handleRequest(req)
		healthyBackendIdx = (healthyBackendIdx + 1) % len(rr.backends)
	}
}

type Backend struct {
	address   string
	isHealthy bool
	client    *http.Client
}

func newBackend(address string) *Backend {
	backend := &Backend{address: address, isHealthy: true, client: &http.Client{}}
	go backend.checkHealth()
	return backend
}

func (b *Backend) checkHealth() {
	for {
		resp, err := http.Get(fmt.Sprintf("%s/health", b.address))
		if err != nil {
			b.isHealthy = false
		}
		if resp.StatusCode != http.StatusOK {
			b.isHealthy = false
		} else {
			b.isHealthy = true
		}
		resp.Body.Close()
		time.Sleep(1 * time.Second)
	}
}

func (b *Backend) handleRequest(c *RequestWithResponseWriter) {

	newReq, err := b.generateNewRequest(c.request)
	if err != nil {
		newResponse := http.Response{
			StatusCode: http.StatusInternalServerError,
			Status:     fmt.Sprintf("Internal Server Error: %s", err.Error()),
		}
		newResponse.Write(c.writer)
		return
	}

	resp, err := b.client.Do(newReq)
	if err != nil {
		newResponse := http.Response{
			StatusCode: http.StatusInternalServerError,
			Status:     fmt.Sprintf("Internal Server Error: %s", err.Error()),
		}
		newResponse.Write(c.writer)
		return
	}

	resp.Write(c.writer)
}

func (b *Backend) generateNewRequest(req *http.Request) (*http.Request, error) {
	backendURL := fmt.Sprintf("%s%s", b.address, req.URL.Path)
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

type RequestWithResponseWriter struct {
	request *http.Request
	writer  http.ResponseWriter
}
