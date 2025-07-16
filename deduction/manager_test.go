package deduction

import (
	"testing"

	"github.com/redis/go-redis/v9"
)

var rdb = redis.NewClient(&redis.Options{
	Addr: "localhost:6379",
})

func TestDeductionPlan(t *testing.T) {
	adGroupID := "test_plan"
	ratio := 70 // 扣量70%，只回传30%
	groupSize := 10
	
	err := InitDeductionPlan(rdb, adGroupID, ratio, groupSize)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	count := 120
	reported := 0

	for i := 0; i < count; i++ {
		ok, err := ShouldReport(rdb, adGroupID, groupSize)
		if err != nil {
			t.Fatalf("Report check failed: %v", err)
		}
		if ok {
			reported++
		}
	}

	t.Logf("点击总数: %d, 回传数: %d, 扣量比例: %.2f%%", count, reported, float64(100*reported)/float64(count))
}
