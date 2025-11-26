package deduction

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

// 生成需要回传的索引
func generateIndexesToReport(ratio int, groupSize int) (result []int) {
	seed := time.Now().UnixNano()
	if ratio == 50 {
		for i := 0; i < groupSize; i++ {
			if i%2 == 0 {
				result = append(result, i)
			}
		}
	} else {
		target := groupSize * (100 - ratio) / 100
		result, _, _ = generateRandomByTarget(groupSize, target, seed)
	}
	return
}

// 获取redis key
func getRedisKeys(adGroupID string) (ratioKey, indexKey, indexesKey string) {
	prefix := fmt.Sprintf("deduct:%s", adGroupID)
	return prefix + ":ratio", prefix + ":index", prefix + ":indexes"
}

// 从redis读取扣量策略
func getRedisDeductData(rdb *redis.Client, key string) []int {
	data, err := rdb.Get(ctx, key).Bytes()
	if err != nil {
		return nil
	}
	var reportIndexes []int
	if err := json.Unmarshal(data, &reportIndexes); err != nil {
		return nil
	}
	return reportIndexes
}

// 初始化或更新某个广告组的扣量策略
func InitDeductionPlan(rdb *redis.Client, adGroupID string, ratio int, groupSize int) error {
	ratioKey, _, indexesKey := getRedisKeys(adGroupID)
	ratioStr := rdb.Get(ctx, ratioKey).Val()
	if ratioStr == fmt.Sprintf("%d", ratio) {
		return nil
	}
	ratioInt, _ := strconv.Atoi(ratioStr)
	indexes := getRedisDeductData(rdb, indexesKey)
	if indexes == nil || ratioInt != ratio {
		indexes = generateIndexesToReport(ratio, groupSize)
		// 把前三个索引替换成 0,1,2
		if ratio != 50 {
			size := len(indexes)
			if size == 1 {
				indexes[0] = 0
			} else if size == 2 {
				indexes[0] = 0
				indexes[1] = 1
			} else if size > 2 {
				indexes[0] = 0
				indexes[1] = 1
				indexes[2] = 2
			}
		}
	}
	data, _ := json.Marshal(indexes)
	pipe := rdb.TxPipeline()
	pipe.Set(ctx, ratioKey, ratio, 0)
	pipe.Set(ctx, indexesKey, data, 0)
	_, err := pipe.Exec(ctx)
	return err
}

// 更新某个广告组的扣量策略
func UpdateDeductionPlan(rdb *redis.Client, adGroupID string, ratio int, groupSize int) error {
	// 删除所有的 key
	ratioKey, indexKey, indexesKey := getRedisKeys(adGroupID)
	// 获取当前的扣量策略
	oldRatio := rdb.Get(ctx, ratioKey).Val()
	if oldRatio == fmt.Sprintf("%d", ratio) {
		return nil
	}
	rdb.Del(ctx, ratioKey, indexKey, indexesKey)
	return InitDeductionPlan(rdb, adGroupID, ratio, groupSize)
}

// 判断当前点击是否需要回传
func ShouldReport(rdb *redis.Client, adGroupID string, groupSize int) (bool, error) {
	_, indexKey, indexesKey := getRedisKeys(adGroupID)

	// 当前点击计数器自增
	index, err := rdb.Incr(ctx, indexKey).Result()
	if err != nil {
		return false, err
	}
	if index == 101 {
		ResetDeductionPlan(rdb, adGroupID, groupSize)
	}
	groupIndex := int((index - 1) % int64(groupSize))

	// 读取当前的回传索引数组
	indexData, err := rdb.Get(ctx, indexesKey).Bytes()
	if err != nil {
		return false, err
	}
	var reportIndexes []int
	if err := json.Unmarshal(indexData, &reportIndexes); err != nil {
		return false, err
	}
	if len(reportIndexes) == 0 {
		return false, nil
	}
	deduct := slices.Contains(reportIndexes, groupIndex)
	return deduct, nil
}

// 如果当前计数为 100，需要重新生成
func ResetDeductionPlan(rdb *redis.Client, adGroupID string, groupSize int) error {
	// 删除所有的 key
	ratioKey, _, indexesKey := getRedisKeys(adGroupID)
	Ratio := rdb.Get(ctx, ratioKey).Val()
	if Ratio == "50" {
		return nil
	}
	fmt.Println("reset", adGroupID)
	RatioInt, _ := strconv.Atoi(Ratio)
	rdb.Del(ctx, ratioKey, indexesKey)
	indexes := generateIndexesToReport(RatioInt, groupSize)
	data, _ := json.Marshal(indexes)
	pipe := rdb.TxPipeline()
	pipe.Set(ctx, ratioKey, RatioInt, 0)
	pipe.Set(ctx, indexesKey, data, 0)
	_, err := pipe.Exec(ctx)
	return err
}
