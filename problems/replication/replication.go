package replication

import (
	"context"
	"fmt"
	"log"

	rdb "github.com/redis/go-redis/v9"
	pkgredis "redis-ops-learning/pkg/redis"
)

// Run executes the replication problem tool: info
func Run(action string) {
	client, err := pkgredis.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()
	ctx := context.Background()

	switch action {
	case "info":
		infoReplication(client, ctx)
	default:
		log.Fatalf("Unknown action: %s (use info)", action)
	}
}

func infoReplication(client rdb.UniversalClient, ctx context.Context) {
	log.Println("[查看复制状态] INFO replication")
	info, err := client.Info(ctx, "replication").Result()
	if err != nil {
		log.Fatalf("INFO replication 失败: %v", err)
	}
	fmt.Println(info)
}
