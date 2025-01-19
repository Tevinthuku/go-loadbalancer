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
 go run cmd/loadbalancer/loadbalancer.go
```

#### Test request distribution by the load balancer to the different backends

```
curl http://localhost:8080/test
```
