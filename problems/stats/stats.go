package stats

import (
	"context"
	"fmt"
	"log"

	rdb "github.com/redis/go-redis/v9"
	pkgredis "redis-ops-learning/pkg/redis"
)

// Run executes the stats problem tool: info, stats
func Run(action string) {
	client, err := pkgredis.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()
	ctx := context.Background()

	switch action {
	case "info":
		infoAll(client, ctx)
	case "stats":
		infoStats(client, ctx)
	default:
		log.Fatalf("Unknown action: %s (use info or stats)", action)
	}
}

func infoAll(client rdb.UniversalClient, ctx context.Context) {
	log.Println("[INFO 全量] server, clients, memory, stats")
	for _, section := range []string{"server", "clients", "memory", "stats"} {
		info, err := client.Info(ctx, section).Result()
		if err != nil {
			log.Printf("INFO %s 失败: %v", section, err)
			continue
		}
		fmt.Printf("# %s\n%s\n", section, info)
	}
}

func infoStats(client rdb.UniversalClient, ctx context.Context) {
	log.Println("[命令统计] INFO stats")
	info, err := client.Info(ctx, "stats").Result()
	if err != nil {
		log.Fatalf("INFO stats 失败: %v", err)
	}
	fmt.Println(info)
}
