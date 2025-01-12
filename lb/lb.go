package lb

import "net"

type LoadBalancer struct {
	ln net.Listener
}

func NewLoadBalancer(address string) (*LoadBalancer, error) {
	ln, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}
	return &LoadBalancer{ln: ln}, nil
}

func (lb *LoadBalancer) Close() {
	lb.ln.Close()
}

func (lb *LoadBalancer) Accept() (net.Conn, error) {
	return lb.ln.Accept()
}
