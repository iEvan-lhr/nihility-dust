package main

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/iEvan-lhr/nihility-dust/anything"
)

// DynamicDust 包含动态路由逻辑
type DynamicDust struct{}

// 1. [入口任务] Analyzer
// 它的作用是分析数据，然后"动态"决定下一步去哪里
func (d *DynamicDust) Analyzer(mission chan *anything.Mission, data []any) {
	fmt.Println("[Analyzer] 正在分析输入数据...", data)

	inputVal := data[0].(int)

	// --- 模拟动态决策逻辑 ---
	// 场景：根据输入值的大小，去不同的部门处理
	// 实际业务中，这里可以是 "select next_step from workflows where condition=..."

	var nextTaskName string
	var processedData []any

	if inputVal > 100 {
		// 如果金额 > 100，走“VIP通道”
		nextTaskName = "VipService"
		processedData = []any{"尊敬的VIP用户", inputVal}
	} else {
		// 否则，走“普通通道”
		nextTaskName = "NormalService"
		processedData = []any{"普通用户", inputVal}
	}

	fmt.Printf("[Analyzer] 决策结果: 下一步跳转到 -> %s\n", nextTaskName)

	// --- 关键点：Name 是变量，不是写死的字符串 ---
	mission <- &anything.Mission{
		Name:    nextTaskName,
		Pursuit: processedData,
	}
}

// 2. [分支任务 A] VipService
func (d *DynamicDust) VipService(mission chan *anything.Mission, data []any) {
	fmt.Printf("[VipService] 处理中: %v, 金额: %v\n", data[0], data[1])
	// VIP 流程结束
	mission <- &anything.Mission{Name: anything.ExitFunction}
}

// 3. [分支任务 B] NormalService
func (d *DynamicDust) NormalService(mission chan *anything.Mission, data []any) {
	fmt.Printf("[NormalService] 处理中: %v, 金额: %v\n", data[0], data[1])
	// 普通流程结束
	mission <- &anything.Mission{Name: anything.ExitFunction}
}

// ----------------------------------------------------------------
// 测试代码
// ----------------------------------------------------------------

func TestDynamicRouting(t *testing.T) {
	// 初始化
	w := &anything.Wind{}
	w.Register(&DynamicDust{})
	w.Init()

	rand.Seed(time.Now().UnixNano())

	// 场景 1: 输入 50 (应该触发 NormalService)
	t.Log("\n--- 测试场景 1: 输入 50 ---")
	done1, _ := anything.DoChanTemp("Analyzer", []any{50}, 0)
	<-done1

	// 场景 2: 输入 150 (应该触发 VipService)
	t.Log("\n--- 测试场景 2: 输入 150 ---")
	done2, _ := anything.DoChanTemp("Analyzer", []any{150}, 0)
	<-done2
}
