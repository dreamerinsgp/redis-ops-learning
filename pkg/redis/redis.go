package redis

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	rdb "github.com/redis/go-redis/v9"
)

// Open creates a Redis client from REDIS_ADDR, REDIS_PASSWORD, REDIS_USERNAME env vars.
// REDIS_ADDR format: host:port (default 127.0.0.1:6379)
// REDIS_PASSWORD: optional
// REDIS_USERNAME: optional, for Redis 6+ ACL (e.g. Aliyun 普通账号)
func Open() (rdb.UniversalClient, error) {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "127.0.0.1:6379"
	}
	// Parse host:port
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, fmt.Errorf("invalid REDIS_ADDR %q: %w", addr, err)
	}
	if port == "" {
		port = "6379"
	}

	password := os.Getenv("REDIS_PASSWORD")
	username := os.Getenv("REDIS_USERNAME")
	dbStr := os.Getenv("REDIS_DB")
	db := 0
	if dbStr != "" {
		if n, err := strconv.Atoi(dbStr); err == nil {
			db = n
		}
	}

	opts := &rdb.Options{
		Addr:     net.JoinHostPort(host, port),
		Password: password,
		DB:       db,
	}
	if username != "" {
		opts.Username = username
	}

	client := rdb.NewClient(opts)
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}
	return client, nil
}

// ParseAddrFromDSN parses "redis://user:pass@host:port" or "host:port" format.
func ParseAddrFromDSN(dsn string) (addr, username, password string, err error) {
	if dsn == "" {
		return "127.0.0.1:6379", "", "", nil
	}
	if strings.HasPrefix(dsn, "redis://") || strings.HasPrefix(dsn, "rediss://") {
		opt, err := rdb.ParseURL(dsn)
		if err != nil {
			return "", "", "", err
		}
		return opt.Addr, opt.Username, opt.Password, nil
	}
	// Plain host:port
	return dsn, "", "", nil
}
