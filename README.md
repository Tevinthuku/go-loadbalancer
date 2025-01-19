## Load balancer

### Setup

#### Run 3 backend servers:

```bash
 BACKEND_PORT=8081 go run cmd/backend/backend.go
```

```bash
 BACKEND_PORT=8082 go run cmd/backend/backend.go
```

```bash
 BACKEND_PORT=8083 go run cmd/backend/backend.go
```

#### Run the load balancer

```bash
LB_PORT=8080 BACKEND_PORTS=8081,8082,8083 go run cmd/loadbalancer/loadbalancer.go
```

#### Test request distribution by the load balancer to the different backends

```bash
curl http://localhost:8080/test
```

```bash
backend  http://localhost:8081  handling request
backend  http://localhost:8082  handling request
backend  http://localhost:8083  handling request
```
