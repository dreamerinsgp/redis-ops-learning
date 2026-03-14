package clients

import (
	"context"
	"fmt"
	"log"

	rdb "github.com/redis/go-redis/v9"
	pkgredis "redis-ops-learning/pkg/redis"
)

// Run executes the clients problem tool: info
func Run(action string) {
	client, err := pkgredis.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()
	ctx := context.Background()

	switch action {
	case "info":
		infoClients(client, ctx)
	default:
		log.Fatalf("Unknown action: %s (use info)", action)
	}
}

func infoClients(client rdb.UniversalClient, ctx context.Context) {
	log.Println("[查看连接] INFO clients")
	info, err := client.Info(ctx, "clients").Result()
	if err != nil {
		log.Fatalf("INFO clients 失败: %v", err)
	}
	fmt.Println(info)
}
