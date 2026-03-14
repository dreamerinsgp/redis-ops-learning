package memory

import (
	"context"
	"fmt"
	"log"
	"sort"

	rdb "github.com/redis/go-redis/v9"
	pkgredis "redis-ops-learning/pkg/redis"
)

// Run executes the memory problem tool: info, bigkeys
func Run(action string) {
	client, err := pkgredis.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()
	ctx := context.Background()

	switch action {
	case "info":
		infoMemory(client, ctx)
	case "bigkeys":
		bigKeys(client, ctx)
	default:
		log.Fatalf("Unknown action: %s (use info or bigkeys)", action)
	}
}

func infoMemory(client rdb.UniversalClient, ctx context.Context) {
	log.Println("[查看内存] INFO memory")
	info, err := client.Info(ctx, "memory").Result()
	if err != nil {
		log.Fatalf("INFO memory 失败: %v", err)
	}
	fmt.Println(info)
}

func bigKeys(client rdb.UniversalClient, ctx context.Context) {
	log.Println("[大 key 采样] SCAN + MEMORY USAGE (采样前 50 个 key)")
	var cursor uint64
	sampled := 0
	maxSamples := 50
	type keySize struct {
		key  string
		size int64
	}
	var largeKeys []keySize
	for {
		keys, next, err := client.Scan(ctx, cursor, "*", 100).Result()
		if err != nil {
			log.Printf("SCAN 失败: %v", err)
			break
		}
		for _, k := range keys {
			size, err := client.MemoryUsage(ctx, k).Result()
			if err != nil {
				continue
			}
			if size > 1024 {
				largeKeys = append(largeKeys, keySize{k, size})
			}
			sampled++
			if sampled >= maxSamples {
				break
			}
		}
		cursor = next
		if cursor == 0 || sampled >= maxSamples {
			break
		}
	}
	if len(largeKeys) == 0 {
		fmt.Println("(未发现明显大 key，或 MEMORY USAGE 不可用。可尝试 redis-cli --bigkeys)")
		return
	}
	sort.Slice(largeKeys, func(i, j int) bool { return largeKeys[i].size > largeKeys[j].size })
	for _, ks := range largeKeys {
		fmt.Printf("%s: %d bytes\n", ks.key, ks.size)
	}
}
