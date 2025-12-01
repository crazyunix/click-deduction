package main

import (
	"fmt"
	"log"

	"github.com/crazyunix/click-deduction/deduction"

	"github.com/redis/go-redis/v9"
)

func main() {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// 启动时初始化策略
	err := deduction.UpdateDeductionPlan(rdb, "plan_p", 30, deduction.DefaultGroupSize)
	if err != nil {
		log.Fatalf("init error: %v", err)
	}

	var reported int
	total := 100
	for i := 0; i < total; i++ {
		ok, _ := deduction.ShouldReport(rdb, "plan_p", deduction.DefaultGroupSize)
		if ok {
			fmt.Printf("Click %d: ✅ 回传\n", i+1)
			reported++
		} else {
			fmt.Printf("Click %d: ❌ 扣量\n", i+1)
		}
	}
	fmt.Printf("点击总数: %d, 回传数: %d, 回传比例: %.2f%%, 扣量比例: %.2f%%", total, reported, float64(100*reported)/float64(total), float64(100*(total-reported))/float64(total))
}
