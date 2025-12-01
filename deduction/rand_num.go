package deduction

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"time"
)

// 自动根据目标数量 total 计算 groupSize 和 pickPerGroup
// 然后从 base 范围内进行分组采样
func generateRandomByTarget(base, target int, seed int64) ([]int, int, int) {
	if target <= 0 || base <= 0 {
		return nil, 0, 0
	}
	rnd := rand.New(rand.NewSource(seed))

	// 自动计算 groupSize 和 pickPerGroup
	// 枚举合理的 groupSize（比如 5 ~ 20）
	bestGroupSize := 0
	bestPickPerGroup := 0
	bestDiff := base // 差值最小优先

	for groupSize := 5; groupSize <= 20; groupSize++ {
		groupCount := int(math.Ceil(float64(base) / float64(groupSize)))
		pickPerGroup := int(math.Ceil(float64(target) / float64(groupCount)))
		actualTotal := groupCount * pickPerGroup
		diff := int(math.Abs(float64(actualTotal - target)))
		if diff < bestDiff {
			bestDiff = diff
			bestGroupSize = groupSize
			bestPickPerGroup = pickPerGroup
		}
	}

	// 开始按照 groupSize 和 pickPerGroup 进行分组随机抽样
	var result []int
	groupCount := int(math.Ceil(float64(base) / float64(bestGroupSize)))

	for i := 0; i < groupCount; i++ {
		start := i * bestGroupSize
		end := start + bestGroupSize
		if end > base {
			end = base
		}
		group := []int{}
		for j := start; j < end; j++ {
			group = append(group, j)
		}
		perm := rnd.Perm(len(group))
		pick := bestPickPerGroup
		if pick > len(perm) {
			pick = len(perm)
		}
		for k := 0; k < pick; k++ {
			result = append(result, group[perm[k]])
		}
	}

	// 截断结果
	if len(result) > target {
		perm := rnd.Perm(len(result))
		truncated := []int{}
		for i := 0; i < target; i++ {
			truncated = append(truncated, result[perm[i]])
		}
		result = truncated
	}

	sort.Ints(result)
	return result, bestGroupSize, bestPickPerGroup
}

// 扣量，保证 index=0 不扣
func generateRemoveIndexes(total, removeCount int) []int {
	result := []int{}

	if removeCount <= 0 {
		return result
	}

	// 有效区间 1~(total-1)
	effectiveTotal := total - 1 // 99

	if removeCount >= effectiveTotal {
		// 全部扣（除了 index=0）
		for i := 1; i < total; i++ {
			result = append(result, i)
		}
		return result
	}

	// gcd 计算块数
	g := gcd(effectiveTotal, removeCount)
	blockSize := effectiveTotal / g
	removePerBlock := removeCount / g

	// 均匀扣点
	for b := 0; b < g; b++ {
		start := 1 + b*blockSize
		interval := float64(blockSize) / float64(removePerBlock)

		for i := 0; i < removePerBlock; i++ {
			idx := start + int(float64(i)*interval)
			if idx >= total {
				idx = total - 1
			}
			result = append(result, idx)
		}
	}

	return result
}

// 求最大公约数
func gcd(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

func RunTest() {
	base := 100
	seed := time.Now().UnixNano()

	plans := map[string]int{
		"plan_A": 90,
		"plan_B": 80,
		"plan_C": 70,
	}

	for planID, target := range plans {
		// result, groupSize, pickCount := generateRandomByTarget(base, target, seed)
		result := generateRemoveIndexes(base, target)
		fmt.Printf("%s: 每 %d 个取 %d，共生成 %d 个\n", planID, base, target, len(result))
		fmt.Println(result)
		seed++ // 改变 seed 确保不同结果
	}
}
