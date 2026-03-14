package main

import (
	"fmt"
	"os"

	cachepenetration "redis-ops-learning/problems/cache_penetration"
	"redis-ops-learning/problems/clients"
	"redis-ops-learning/problems/memory"
	"redis-ops-learning/problems/replication"
	"redis-ops-learning/problems/slowlog"
	"redis-ops-learning/problems/stats"
)

func main() {
	if len(os.Args) < 3 {
		printUsage()
		os.Exit(1)
	}
	subcmd := os.Args[1]
	problem := os.Args[2]
	action := ""
	if len(os.Args) >= 4 {
		action = os.Args[3]
	}

	if subcmd != "run" {
		printUsage()
		os.Exit(1)
	}

	switch problem {
	case "01-memory":
		memory.Run(action)
	case "02-clients":
		clients.Run(action)
	case "03-slowlog":
		slowlog.Run(action)
	case "04-replication":
		replication.Run(action)
	case "05-stats":
		stats.Run(action)
	case "06-cache-penetration":
		cachepenetration.Run(action)
	default:
		fmt.Printf("Unknown problem: %s\n", problem)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`Usage: go run ./cmd run <problem> [action]

Problems and actions:
  01-memory            info | bigkeys
  02-clients           info
  03-slowlog           info | slowlog
  04-replication       info
  05-stats             info | stats
  06-cache-penetration info | simulate | bloom

Set env:
  REDIS_ADDR=host:port     (default 127.0.0.1:6379)
  REDIS_PASSWORD=          (optional)
  REDIS_USERNAME=          (optional, for Redis 6+ ACL)`)
}
