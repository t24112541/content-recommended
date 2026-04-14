# app

## project structure

```
# > tree ./
├───config          # server and router config
├───database        # database connections
│   ├───postgres
│   └───redis
├───handler         # all handlers
├───model           # all models
│   ├───orm         # model for db
│   ├───request     # model for request
│   └───response    # model for response
├───router          # all routers
└───service         # all service

```

## setup

this app require:

1. Go version go1.25.0 windows/amd64
2. Docker version 29.3.1

#### setup step

1. copy env.example to .env then setup env
2. run command for up container

   ```
   docker compose up -d postgres redis
   go run main.go -migrate -seed

   ```

   note: command `go run main.go` can run app but if want to run on container `ctrl + c` then `docker compose up -d app`

---

### API Overview

- Base URL: `http://localhost:8000`
- Main endpoint under test: `GET /users/{userId}/recommendations?limit={n}`
- Batch endpoint under test: `GET /recommendations/batch?page={p}&limit={n}`

### Performance Testing Requirements

#### k6 Test Scenarios

1. **Single User Load Test**  
   100 requests/second for 1 minute
2. **Batch Endpoint Stress Test**  
   Concurrent batch requests with varying page sizes
3. **Cache Effectiveness Test**  
   Repeated identical requests to measure cache hit ratio

#### k6 Commands

click [command](./test/readme.md)

#### k6 result

click [result](./test/result.md)

---

#### migration

```
go run main.go -migrate
```

#### seed data

```
go run main.go -seed
```

#### auto reload

```
air
```

#### manual reload

```
go run main.go
```

#### docker compose

```
docker compose up -d
```

#### if up some container

```
docker compose up -d [app, postgres, redis]
```
