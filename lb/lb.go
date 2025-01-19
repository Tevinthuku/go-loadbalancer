package lb

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync/atomic"
	"syscall"
	"time"
)

type LoadBalancer struct {
	port        string
	backends    []Backend
	requestChan chan *RequestContext
	svr         *http.Server
}

func NewLoadBalancer(port string, backends []Backend) *LoadBalancer {
	lb := &LoadBalancer{port: port, requestChan: make(chan *RequestContext), backends: make([]Backend, 0, len(backends))}
	lb.backends = append(lb.backends, backends...)
	svr := http.Server{
		Addr:    lb.port,
		Handler: http.HandlerFunc(lb.resolve),
	}
	lb.svr = &svr
	go lb.worker()
	return lb
}

func (lb *LoadBalancer) Start() error {

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := lb.svr.Shutdown(ctx); err != nil {
			fmt.Printf("HTTP server shutdown error: %v\n", err)
		}
		close(lb.requestChan)
	}()

	return lb.svr.ListenAndServe()
}

func (lb *LoadBalancer) resolve(w http.ResponseWriter, req *http.Request) {
	responseChan := make(chan *http.Response, 1) // buffered channel to avoid goroutine leak
	lb.requestChan <- &RequestContext{request: req, ResponseChan: responseChan}
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
		if backend.IsHealthy() {
			go backend.HandleRequest(req)
			healthyBackendIdx = (healthyBackendIdx + 1) % len(lb.backends)
		} else {
			// check if the next backends are healthy or not
			healthyBackendFound := false
			// loop through all backends to find a healthy one
			for i := 1; i < len(lb.backends); i++ {
				absoluteIdx := (healthyBackendIdx + i) % len(lb.backends)
				if lb.backends[absoluteIdx].IsHealthy() {
					healthyBackendFound = true
					go lb.backends[absoluteIdx].HandleRequest(req)
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
				req.ResponseChan <- &response
			}
		}
	}
}

func (lb *LoadBalancer) Close() {
	close(lb.requestChan)
	if err := lb.svr.Shutdown(context.Background()); err != nil {
		fmt.Println("error during shutdown: ", err)
	}
}

type Backend interface {
	HandleRequest(c *RequestContext)
	IsHealthy() bool
}

type HttpBackend struct {
	address string
	health  atomic.Bool
	client  *http.Client
}

func NewHttpBackend(address string) *HttpBackend {
	backend := &HttpBackend{address: address, client: &http.Client{}}
	backend.health.Store(true)
	go backend.checkHealth()
	return backend
}

func (b *HttpBackend) checkHealth() {
	for {
		resp, err := b.client.Get(fmt.Sprintf("%s/health", b.address))
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

func (b *HttpBackend) HandleRequest(c *RequestContext) {
	fmt.Println("backend ", b.address, " handling request")
	newReq, err := b.generateNewRequest(c.request)
	if err != nil {
		newResponse := http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       io.NopCloser(strings.NewReader(fmt.Sprintf("Internal Server Error: %s", err.Error()))),
		}
		c.ResponseChan <- &newResponse
		return
	}

	resp, err := b.client.Do(newReq)
	if err != nil {
		newResponse := http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       io.NopCloser(strings.NewReader(fmt.Sprintf("Internal Server Error: %s", err.Error()))),
		}
		c.ResponseChan <- &newResponse
		return
	}

	c.ResponseChan <- resp
}

func (b *HttpBackend) generateNewRequest(req *http.Request) (*http.Request, error) {
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

func (b *HttpBackend) IsHealthy() bool {
	return b.health.Load()
}

type RequestContext struct {
	request      *http.Request
	ResponseChan chan *http.Response
}
