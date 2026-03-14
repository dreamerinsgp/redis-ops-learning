package bigkey

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	rdb "github.com/redis/go-redis/v9"
	pkgredis "redis-ops-learning/pkg/redis"
)

// Run executes the bigkey problem tool
func Run(action string) {
	client, err := pkgredis.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()
	ctx := context.Background()

	switch action {
	case "info":
		showBigKeyInfo(ctx, client)
	case "scan":
		scanBigKeys(ctx, client)
	case "demo":
		demoBigKeyProblem(ctx, client)
	case "split":
		splitBigHash(ctx, client)
	default:
		log.Fatalf("Unknown action: %s (use info, scan, demo, split)", action)
	}
}

// showBigKeyInfo shows basic big key information
func showBigKeyInfo(ctx context.Context, client rdb.UniversalClient) {
	log.Println("[大key信息] INFO memory + keyspace")

	// Get memory info
	info, err := client.Info(ctx, "memory").Result()
	if err != nil {
		log.Fatalf("INFO memory 失败: %v", err)
	}
	fmt.Println("=== MEMORY INFO ===")
	fmt.Println(info)

	// Get keyspace info
	keysace, err := client.Info(ctx, "keyspace").Result()
	if err != nil {
		log.Printf("INFO keyspace 失败: %v", err)
	} else {
		fmt.Println("=== KEYSPACE INFO ===")
		fmt.Println(keysace)
	}
}

// scanBigKeys scans for big keys in the database
func scanBigKeys(ctx context.Context, client rdb.UniversalClient) {
	log.Println("[大key扫描] SCAN 扫描 + MEMORY USAGE 检测（采样 500 个 key）")

	var cursor uint64
	sampled := 0
	maxSamples := 500
	type keySize struct {
		key     string
		size    int64
		ttl     time.Duration
		keyType string
	}
	var largeKeys []keySize

	for {
		keys, next, err := client.Scan(ctx, cursor, "*", 100).Result()
		if err != nil {
			log.Printf("SCAN 失败: %v", err)
			break
		}
		for _, k := range keys {
			// Get key type
			keyType, err := client.Type(ctx, k).Result()
			if err != nil {
				continue
			}

			// Get memory usage
			size, err := client.MemoryUsage(ctx, k).Result()
			if err != nil {
				continue
			}

			// Get TTL
			ttl, _ := client.TTL(ctx, k).Result()

			// Collect keys > 10KB
			if size > 10240 {
				largeKeys = append(largeKeys, keySize{k, size, ttl, keyType})
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
		fmt.Println("(未发现 > 10KB 的大 key)")
		fmt.Println("提示：运行 demo 动作可创建演示用大 key，再运行 scan 查看")
		return
	}

	// Sort by size descending
	sort.Slice(largeKeys, func(i, j int) bool { return largeKeys[i].size > largeKeys[j].size })

	fmt.Println("\n=== 发现的大 Key (Top 20) ===")
	fmt.Printf("%-40s %15s %10s %s\n", "KEY", "SIZE", "TYPE", "TTL")
	fmt.Println(strings.Repeat("-", 85))

	for i, ks := range largeKeys {
		if i >= 20 {
			break
		}
		ttlStr := "-"
		if ks.ttl > 0 {
			ttlStr = ks.ttl.Round(time.Second).String()
		}
		fmt.Printf("%-40s %15d %10s %s\n", ks.key, ks.size, ks.keyType, ttlStr)
	}
}

// demoBigKeyProblem demonstrates the big key problem
func demoBigKeyProblem(ctx context.Context, client rdb.UniversalClient) {
	log.Println("[大key问题演示] 创建大 Hash 并演示阻塞")

	// Clean up first
	keysToClean := []string{
		"demo:big_hash",
		"demo:big_list",
		"demo:big_string",
	}
	for _, k := range keysToClean {
		client.Del(ctx, k)
	}

	// Create a big hash (10000 fields)
	fmt.Println("1. 创建大 Hash (10000 个字段)...")
	bigHash := make(map[string]interface{})
	for i := 0; i < 10000; i++ {
		bigHash[fmt.Sprintf("field:%d", i)] = fmt.Sprintf("value:%d", i)
	}
	err := client.HSet(ctx, "demo:big_hash", bigHash).Err()
	if err != nil {
		log.Fatalf("HSet 失败: %v", err)
	}

	// Check size
	size, _ := client.MemoryUsage(ctx, "demo:big_hash").Result()
	fmt.Printf("   demo:big_hash 大小: %d bytes (%.2f KB)\n", size, float64(size)/1024)

	// Create a big string (1MB)
	fmt.Println("2. 创建大 String (1MB)...")
	bigString := strings.Repeat("x", 1024*1024)
	err = client.Set(ctx, "demo:big_string", bigString, 0).Err()
	if err != nil {
		log.Fatalf("Set 失败: %v", err)
	}

	size, _ = client.MemoryUsage(ctx, "demo:big_string").Result()
	fmt.Printf("   demo:big_string 大小: %d bytes (%.2f KB)\n", size, float64(size)/1024)

	// Create a big list (10000 elements)
	fmt.Println("3. 创建大 List (10000 个元素)...")
	var list []interface{}
	for i := 0; i < 10000; i++ {
		list = append(list, fmt.Sprintf("item:%d", i))
	}
	err = client.RPush(ctx, "demo:big_list", list...).Err()
	if err != nil {
		log.Fatalf("RPush 失败: %v", err)
	}

	size, _ = client.MemoryUsage(ctx, "demo:big_list").Result()
	fmt.Printf("   demo:big_list 大小: %d bytes (%.2f KB)\n", size, float64(size)/1024)

	// Now demonstrate the problem - HGETALL on big hash is slow
	fmt.Println("\n4. 演示问题：HGETALL vs HSCAN...")
	fmt.Println("   执行 HGETALL demo:big_hash (阻塞)...")
	start := time.Now()
	result, err := client.HGetAll(ctx, "demo:big_hash").Result()
	elapsed := time.Since(start)
	if err != nil {
		log.Printf("HGETALL 失败: %v", err)
	} else {
		fmt.Printf("   HGETALL 返回 %d 个字段，耗时 %v\n", len(result), elapsed)
		if elapsed > time.Millisecond*100 {
			fmt.Println("   ⚠️ 注意：HGETALL 耗时超过 100ms，对线上有风险！")
		}
	}

	fmt.Println("\n5. 演示数据已保留（供 scan 动作检测）")
	fmt.Println("   提示：再次运行 demo 将自动清理并重建；或手动 DEL demo:big_*")
}

// splitBigHash demonstrates how to split a big hash
func splitBigHash(ctx context.Context, client rdb.UniversalClient) {
	log.Println("[大key拆分方案] 演示 Hash 拆分方法")

	// Clean up first
	keysToClean := []string{
		"demo:user:1001",
		"demo:user:1001:profile",
		"demo:user:1001:attrs",
	}
	for _, k := range keysToClean {
		client.Del(ctx, k)
	}

	// Create original big hash
	fmt.Println("1. 创建原始 Hash (模拟用户画像，1000 字段)...")
	bigHash := make(map[string]interface{})
	for i := 0; i < 1000; i++ {
		bigHash[fmt.Sprintf("attr:%d", i)] = fmt.Sprintf("value:%d", i)
	}
	err := client.HSet(ctx, "demo:user:1001", bigHash).Err()
	if err != nil {
		log.Fatalf("HSet 失败: %v", err)
	}

	originalSize, _ := client.MemoryUsage(ctx, "demo:user:1001").Result()
	fmt.Printf("   原始 Hash 大小: %d bytes\n", originalSize)

	// Split into smaller hashes
	fmt.Println("2. 拆分为多个小 Hash (每 100 字段一个 Hash)...")
	fieldCount := 1000
	chunkSize := 100
	for chunk := 0; chunk*chunkSize < fieldCount; chunk++ {
		start := chunk * chunkSize
		end := start + chunkSize
		if end > fieldCount {
			end = fieldCount
		}

		chunkKey := fmt.Sprintf("demo:user:1001:attrs:%d", chunk)
		chunkData := make(map[string]interface{})
		for i := start; i < end; i++ {
			chunkData[fmt.Sprintf("attr:%d", i)] = fmt.Sprintf("value:%d", i)
		}

		err := client.HSet(ctx, chunkKey, chunkData).Err()
		if err != nil {
			log.Fatalf("HSet chunk 失败: %v", err)
		}
	}

	// Check new size
	var totalSize int64 = 0
	for chunk := 0; chunk*chunkSize < fieldCount; chunk++ {
		chunkKey := fmt.Sprintf("demo:user:1001:attrs:%d", chunk)
		size, _ := client.MemoryUsage(ctx, chunkKey).Result()
		totalSize += size
	}

	fmt.Printf("   拆分后总大小: %d bytes\n", totalSize)
	fmt.Printf("   节省空间: %d bytes (%.1f%%)\n", originalSize-totalSize,
		float64(originalSize-totalSize)/float64(originalSize)*100)

	// Demonstrate reading with HSCAN
	fmt.Println("3. 使用 HSCAN 分批读取...")
	var cursor uint64
	fieldsRead := 0
	for {
		var keys []string
		keys, cursor, err = client.HScan(ctx, "demo:user:1001", cursor, "*", 100).Result()
		if err != nil {
			break
		}
		fieldsRead += len(keys) / 2 // each field has a value
		if cursor == 0 {
			break
		}
	}
	fmt.Printf("   HSCAN 遍历 %d 个字段（不阻塞）\n", fieldsRead)

	// Clean up
	fmt.Println("\n4. 清理演示数据...")
	for _, k := range keysToClean {
		client.Del(ctx, k)
	}
	// Also clean up chunk keys
	for chunk := 0; chunk*chunkSize < fieldCount; chunk++ {
		client.Del(ctx, fmt.Sprintf("demo:user:1001:attrs:%d", chunk))
	}
	fmt.Println("   清理完成")
}
