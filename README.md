# Redis Ops Learning Project

A Go-based project for learning common Redis operations issues. Each problem provides tools to inspect, monitor, and diagnose Redis.

## Prerequisites

- Go 1.21+
- Redis (local or cloud, e.g. Aliyun Redis)

## Setup

1. Copy `.env.example` to `.env`
2. Fill `REDIS_ADDR` (host:port), `REDIS_PASSWORD`, `REDIS_USERNAME` (if Redis 6+ ACL)
3. Load env: `source .env` or `export $(cat .env | xargs)`
4. Add your IP to Redis whitelist (for cloud Redis)

## Usage

```bash
# Memory: view memory info, sample big keys
go run ./cmd run 01-memory info
go run ./cmd run 01-memory bigkeys

# Clients: view connection stats
go run ./cmd run 02-clients info

# Slow log: view config, recent slow commands
go run ./cmd run 03-slowlog info
go run ./cmd run 03-slowlog slowlog

# Replication: view replication status
go run ./cmd run 04-replication info

# Stats: full info, command statistics
go run ./cmd run 05-stats info
go run ./cmd run 05-stats stats
```

## Structure

```
redis-ops-learning/
├── cmd/main.go           # CLI entry, dispatches to problems
├── pkg/redis/            # Shared Redis client (REDIS_ADDR, REDIS_PASSWORD, REDIS_USERNAME)
└── problems/
    ├── memory/           # 01-memory: memory usage, big keys
    ├── clients/          # 02-clients: connection stats
    ├── slowlog/          # 03-slowlog: slow command config and log
    ├── replication/      # 04-replication: replication status
    └── stats/            # 05-stats: server stats
```

## DEX Ops Dashboard Integration

The Performance dashboard **Redis Ops** tab can run these tools when integrated:

1. Save Infra config with Redis (host, port, password, username)
2. Ensure `redis-ops-learning` is deployable alongside `mysql-ops-learning`
3. Backend can invoke `go run ./cmd run <problem> <action>` with REDIS_* env from config

## Environment Variables

| Variable       | Description                          | Default          |
|----------------|--------------------------------------|------------------|
| REDIS_ADDR     | host:port                             | 127.0.0.1:6379   |
| REDIS_PASSWORD | Password (optional)                  | -                |
| REDIS_USERNAME | Username for Redis 6+ ACL (optional) | -                |
| REDIS_DB       | Database number (0-15)               | 0                |
# redis-ops-learning
