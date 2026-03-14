package slowlog

import (
	"context"
	"fmt"
	"log"

	rdb "github.com/redis/go-redis/v9"
	pkgredis "redis-ops-learning/pkg/redis"
)

// Run executes the slowlog problem tool: info, slowlog
func Run(action string) {
	client, err := pkgredis.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()
	ctx := context.Background()

	switch action {
	case "info":
		infoConfig(client, ctx)
	case "slowlog":
		viewSlowlog(client, ctx)
	default:
		log.Fatalf("Unknown action: %s (use info or slowlog)", action)
	}
}

func infoConfig(client rdb.UniversalClient, ctx context.Context) {
	log.Println("[查看慢日志配置] CONFIG GET slowlog*")
	cfg, err := client.ConfigGet(ctx, "slowlog*").Result()
	if err != nil {
		log.Printf("CONFIG 可能被禁用: %v", err)
		return
	}
	for k, v := range cfg {
		fmt.Printf("%s: %s\n", k, v)
	}
}

func viewSlowlog(client rdb.UniversalClient, ctx context.Context) {
	log.Println("[最近慢命令] SLOWLOG GET 10")
	logs, err := client.SlowLogGet(ctx, 10).Result()
	if err != nil {
		log.Fatalf("SLOWLOG GET 失败: %v", err)
	}
	if len(logs) == 0 {
		fmt.Println("(无慢命令记录)")
		return
	}
	for i, e := range logs {
		cmd := ""
		for _, c := range e.Args {
			cmd += fmt.Sprintf("%v ", c)
		}
		if len(cmd) > 80 {
			cmd = cmd[:80] + "..."
		}
		fmt.Printf("%d. %d us - %s\n", i+1, e.Duration.Microseconds(), cmd)
	}
}
