package cachepenetration

import (
	"context"
	"fmt"
	"log"
	"time"

	rdb "github.com/redis/go-redis/v9"
	pkgredis "redis-ops-learning/pkg/redis"
)

// Run executes the cache penetration problem tool
func Run(action string) {
	client, err := pkgredis.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()
	ctx := context.Background()

	switch action {
	case "info":
		showInfo(client, ctx)
	case "simulate":
		simulate(client, ctx)
	case "bloom":
		showBloomFilter(client, ctx)
	default:
		log.Fatalf("Unknown action: %s (use info, simulate, or bloom)", action)
	}
}

// showInfo demonstrates cache penetration concept with real Redis data
func showInfo(client rdb.UniversalClient, ctx context.Context) {
	fmt.Println("=== 缓存穿透 (Cache Penetration) 演示 ===")
	fmt.Println()
	fmt.Println("【场景】查询不存在的商品 key")
	fmt.Println("【问题】每次查询都穿透到数据库，数据库压力大")
	fmt.Println()

	// 准备真实测试数据：存在 vs 不存在的 key 对比
	demoPrefix := "demo:cp:"
	productExists := demoPrefix + "product:1001"
	productNotExists := demoPrefix + "product:9999"

	log.Println("[准备数据] 写入存在的商品 product:1001，product:9999 不写入（模拟不存在）")
	client.Set(ctx, productExists, `{"id":1001,"name":"商品A","price":99}`, 0)
	client.Del(ctx, productNotExists) // 确保不存在

	// Check if keys exist
	exists, _ := client.Exists(ctx, productExists).Result()
	notExists, _ := client.Exists(ctx, productNotExists).Result()

	fmt.Printf("product:1001 (存在)   EXISTS: %d\n", exists)
	fmt.Printf("product:9999 (不存在) EXISTS: %d\n", notExists)
	fmt.Println()

	// Simulate penetration: query non-existent key
	log.Println("[模拟穿透] GET product:9999 (不存在的 key，每次都会穿透)")
	val, err := client.Get(ctx, productNotExists).Result()
	if err == rdb.Nil {
		fmt.Printf("→ 返回 nil，缓存未命中 (Cache Miss)\n")
		fmt.Println("→ 如果没有空值缓存，每次都会查询数据库 = 穿透！\n")
	} else if err != nil {
		fmt.Printf("→ 错误: %v\n", err)
	} else {
		fmt.Printf("→ 值: %s\n", val)
	}

	// Show statistics
	fmt.Println("【相关统计指标】")
	info, _ := client.Info(ctx, "stats").Result()
	lines := getStatsLines(info, []string{"keyspace_hits", "keyspace_misses"})
	for _, line := range lines {
		fmt.Println(line)
	}

	hitRate := getKeyspaceHitRate(info)
	fmt.Printf("缓存命中率: %.2f%%\n", hitRate)

	// 清理演示数据
	client.Del(ctx, productExists)
	fmt.Println()
	fmt.Println("[清理] 演示 key 已删除")
}

// simulate demonstrates the penetration pattern
func simulate(client rdb.UniversalClient, ctx context.Context) {
	fmt.Println("=== 模拟缓存穿透场景 ===")
	fmt.Println()

	testKey := "test:penetration:product:8888"

	// Clean up first
	client.Del(ctx, testKey)

	// Step 1: Normal query (cache miss -> DB)
	log.Printf("[Step 1] GET %s (首次查询，缓存不存在)", testKey)
	val1, err := client.Get(ctx, testKey).Result()
	if err == rdb.Nil {
		fmt.Println("→ Cache Miss! (需要查询数据库)")
	}
	_ = val1
	fmt.Println()

	// Step 2: Set empty cache to prevent penetration
	log.Println("[Step 2] SETEX 设置空值缓存 (防止穿透)")
	client.SetEx(ctx, testKey, "NULL", 30*time.Second)
	fmt.Println("→ 设置空值，TTL=30秒")
	fmt.Println()

	// Step 3: Query again (cache hit with null value)
	log.Printf("[Step 3] GET %s (已有空值缓存)", testKey)
	val2, err := client.Get(ctx, testKey).Result()
	if err == rdb.Nil {
		fmt.Println("→ 缓存过期或不存在")
	} else {
		fmt.Printf("→ 命中空值缓存，值: %s\n", val2)
		fmt.Println("→ 无需查数据库，穿透已防止！")
	}

	// Show keys
	keys, _ := client.Keys(ctx, "test:penetration:*").Result()
	fmt.Printf("\n当前测试 keys: %v\n", keys)

	// Clean up
	client.Del(ctx, keys...)
	fmt.Println("测试 key 已清理")
}

// showBloomFilter shows how bloom filter helps prevent penetration
func showBloomFilter(client rdb.UniversalClient, ctx context.Context) {
	fmt.Println("=== 布隆过滤器防穿透方案 ===")
	fmt.Println()

	bfKey := "bloom:products"

	// Clean up
	client.Del(ctx, bfKey)

	// Create bloom filter using Redis module (simulate with SET)
	// Note: In production, use RedisBloom module (BF.ADD)
	log.Println("[Step 1] 初始化布隆过滤器 (模拟)")
	fmt.Println("→ 添加存在的商品 ID 到布隆过滤器")
	
	// Simulate adding existing product IDs
	products := []string{"1001", "1002", "1003", "1004", "1005"}
	for _, p := range products {
		fmt.Printf("→ BF.ADD %s product:%s\n", bfKey, p)
	}
	fmt.Println()

	// Check existing product
	log.Println("[Step 2] 查询存在的商品 (product:1001)")
	fmt.Println("→ 先检查布隆过滤器 → 可能存在 → 查询缓存/数据库")
	
	// Check non-existing product
	log.Println("[Step 3] 查询不存在的商品 (product:9999)")
	fmt.Println("→ 先检查布隆过滤器 → 不存在 → 直接返回 null")
	fmt.Println("→ 无需查询数据库，完全防止穿透！")
	fmt.Println()

	fmt.Println("【布隆过滤器方案优势】")
	fmt.Println("1. 极低的内存占用 (位图)")
	fmt.Println("2. 快速判断 key 是否可能存在")
	fmt.Println("3. 完全防止不存在 key 的穿透")
	fmt.Println()
	fmt.Println("【注意事项】")
	fmt.Println("- 布隆过滤器有误判率 (false positive)")
	fmt.Println("- 不存在的 key 一定返回不存在")
	fmt.Println("- 存在的 key 可能误判为存在 (可接受)")
}

func getStatsLines(info string, keys []string) []string {
	var result []string
	lines := getInfoLines(info)
	for _, line := range lines {
		for _, k := range keys {
			if len(line) > len(k) && line[:len(k)+1] == k+":" {
				result = append(result, line)
				break
			}
		}
	}
	return result
}

func getKeyspaceHitRate(info string) float64 {
	var hits, misses int64
	lines := getInfoLines(info)
	for _, line := range lines {
		if len(line) > 14 && line[:14] == "keyspace_hits:" {
			fmt.Sscanf(line, "keyspace_hits:%d", &hits)
		}
		if len(line) > 15 && line[:15] == "keyspace_misses:" {
			fmt.Sscanf(line, "keyspace_misses:%d", &misses)
		}
	}
	total := hits + misses
	if total == 0 {
		return 0
	}
	return float64(hits) / float64(total) * 100
}

func getInfoLines(info string) []string {
	var lines []string
	start := 0
	for i, c := range info {
		if c == '\n' {
			lines = append(lines, info[start:i])
			start = i + 1
		}
	}
	return lines
}
