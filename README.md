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

## auto reload

```
air
```

## manual reload

```
go run main.go
```

## migration

```
go run main.go -migrate
```

## seed data

```
go run main.go -seed
```

# docker compose

```
docker compose up -d
```

### if up some container

```
docker compose up -d [app, postgres, redis]
```
