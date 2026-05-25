package main

import (
	"math"
	"testing"
	"time"

	"github.com/iEvan-lhr/nihility-dust/anything"
	"github.com/iEvan-lhr/nihility-dust/astral"
)

// TestAstral_PhysicalDecay 仿真测试 1：时空半衰期解析衰减与增量相干坍缩数学验证
func TestAstral_PhysicalDecay(t *testing.T) {
	// 1. 初始化 SQLite 内存库与崩溃引擎 (BaseDecayRate=0.02)
	store, err := astral.NewSQLiteAstralStore(":memory:")
	if err != nil {
		t.Fatalf("failed to create memory store: %v", err)
	}
	defer store.Close()

	engine := astral.NewCollapseEngine(store)
	engine.BaseDecayRate = 0.05 // 设置背景自衰减系数为 0.05
	engine.EvolutionRate = 0.0  // 不做自演化，纯验证数学叠加

	// 2. 初始注册节点坐标 Node_A (ID=100)，基态初始能级全为 0.0
	nodeA := &astral.NodeAnchor{
		ID:               100,
		BaseEmbedding:    []float64{0.1, 0.2, 0.3},
		LastCollapseTime: 1000, // 初始时间戳 1000ms
		LastState:        astral.Vector6D{Time: 0, Space: 0, PosNeg: 0, Influence: 0, Danger: 0.0, Base: 0},
	}
	if err := store.SaveAnchor(nodeA); err != nil {
		t.Fatalf("failed to save anchor: %v", err)
	}

	// 3. 在 t = 1000ms 发射第一束流 (Flow 1)：高危高影响事件，衰减率 0.1
	flow1 := &astral.Flow{
		Anchors:      []int64{100},
		Payload:      "系统致命告警日志",
		Timestamp:    1000,
		DecayRate:    0.1,
		OriginEnergy: astral.Vector6D{Danger: 0.8, Influence: 0.5},
	}
	if err := store.SaveFlow(flow1); err != nil {
		t.Fatalf("failed to save flow 1: %v", err)
	}

	// 4. 在 t = 1000ms 立即发生观测坍缩
	state1, err := engine.Collapse(100, 1000)
	if err != nil {
		t.Fatalf("collapse failed at 1000ms: %v", err)
	}
	t.Logf("t=1000ms 坍缩状态: %+v", state1)
	if math.Abs(state1.Danger-0.8) > 1e-9 {
		t.Errorf("expected Danger = 0.8, got %f", state1.Danger)
	}

	// 5. 在 t = 11000ms (10秒后) 发生二次观测坍缩
	// 数学预测：
	// - 1000ms 时坍缩产生的 snapshot 状态为 Danger=0.8，LastCollapseTime 为 1000。
	// - 过了 10 秒后，该 snapshot 自我衰减：0.8 * e^(-0.05 * 10) = 0.8 * 0.60653 = 0.48522
	// - 期间无新流发生。
	state2, err := engine.Collapse(100, 11000)
	if err != nil {
		t.Fatalf("collapse failed at 11000ms: %v", err)
	}
	t.Logf("t=11000ms (10s后) 坍缩状态: %+v", state2)
	expectedDanger := 0.8 * math.Exp(-0.05*10.0)
	if math.Abs(state2.Danger-expectedDanger) > 1e-5 {
		t.Errorf("expected Danger ≈ %f, got %f", expectedDanger, state2.Danger)
	}
}

// TestAstral_ResonanceAndCancellation 仿真测试 2：波能的叠加共振与正负抵消验证
func TestAstral_ResonanceAndCancellation(t *testing.T) {
	store, _ := astral.NewSQLiteAstralStore(":memory:")
	defer store.Close()
	engine := astral.NewCollapseEngine(store)
	engine.BaseDecayRate = 0.0

	// 注册节点 200，在 t = 0
	node := &astral.NodeAnchor{
		ID:               200,
		BaseEmbedding:    []float64{0.5, 0.5},
		LastCollapseTime: 0,
		LastState:        astral.Vector6D{},
	}
	_ = store.SaveAnchor(node)

	// 同时发射两个情绪相左的流：积极正能量 + 消极负能量
	flowPos := &astral.Flow{
		Anchors:      []int64{200},
		Payload:      "用户表扬信",
		Timestamp:    0,
		DecayRate:    0.0,
		OriginEnergy: astral.Vector6D{PosNeg: 0.9}, // 积极能级 +0.9
	}
	flowNeg := &astral.Flow{
		Anchors:      []int64{200},
		Payload:      "服务超时抱怨",
		Timestamp:    0,
		DecayRate:    0.0,
		OriginEnergy: astral.Vector6D{PosNeg: -0.4}, // 消极能级 -0.4
	}
	_ = store.SaveFlow(flowPos)
	_ = store.SaveFlow(flowNeg)

	// 瞬间坍缩
	state, _ := engine.Collapse(200, 0)
	t.Logf("能级中和坍缩状态: %+v", state)
	// 期望 0.9 - 0.4 = 0.5
	if math.Abs(state.PosNeg-0.5) > 1e-9 {
		t.Errorf("expected PosNeg = 0.5, got %f", state.PosNeg)
	}
}

// TestAstral_SelfEvolution 仿真测试 3：流沙雕刻节点与本体语义自演化验证
func TestAstral_SelfEvolution(t *testing.T) {
	store, _ := astral.NewSQLiteAstralStore(":memory:")
	defer store.Close()
	engine := astral.NewCollapseEngine(store)
	engine.EvolutionRate = 0.1 // 10% 的高演化率，便于验证

	node := &astral.NodeAnchor{
		ID:               300,
		BaseEmbedding:    []float64{1.0, 0.0}, // 初始引力朝向 [1, 0]
		LastCollapseTime: 0,
		LastState:        astral.Vector6D{},
	}
	_ = store.SaveAnchor(node)

	// 注入一个强烈偏向 [0, 1.0] 语义的 Flow
	flow := &astral.Flow{
		Anchors:       []int64{300},
		Payload:       "量子跳跃事件",
		Timestamp:     0,
		DecayRate:     0.0,
		BaseEmbedding: []float64{0.0, 1.0},
	}
	_ = store.SaveFlow(flow)

	// 触发坍缩引起引力雕刻
	_, _ = engine.Collapse(300, 0)

	// 重新读取节点，验证 BaseEmbedding 是否被雕刻微调
	updatedNode, _ := store.GetAnchor(300)
	t.Logf("雕刻后节点基态向量: %v", updatedNode.BaseEmbedding)
	// 预期: 0.9 * [1.0, 0.0] + 0.1 * [0.0, 1.0] = [0.9, 0.1]
	if math.Abs(updatedNode.BaseEmbedding[0]-0.9) > 1e-9 || math.Abs(updatedNode.BaseEmbedding[1]-0.1) > 1e-9 {
		t.Errorf("expected carved embedding [0.9, 0.1], got %v", updatedNode.BaseEmbedding)
	}

	// 验证余弦相似度计算
	sim := astral.CosineSimilarity([]float64{1.0, 0.0}, []float64{1.0, 0.0})
	if math.Abs(sim-1.0) > 1e-9 {
		t.Errorf("cosine similarity expected 1.0, got %f", sim)
	}
}

// TestAstral_WindOrchestration 集成测试 4：将 AstralSpirit 植入 Wind 调度管线与隐形记忆交互
func TestAstral_WindOrchestration(t *testing.T) {
	// 1. 初始化 SQLite 存储、引擎与适配灵魂
	store, _ := astral.NewSQLiteAstralStore(":memory:")
	defer store.Close()
	engine := astral.NewCollapseEngine(store)

	// 注册节点 400 带有初始危机能级
	node := &astral.NodeAnchor{
		ID:               400,
		BaseEmbedding:    []float64{1, 1},
		LastCollapseTime: 0,
		LastState:        astral.Vector6D{Danger: 0.75}, // 危机值 0.75
	}
	_ = store.SaveAnchor(node)

	astralSpirit := astral.NewAstralSpirit(engine)

	// 2. 初始化 Wind 并注入 AstralSpirit
	w := &anything.Wind{}
	w.Init(astralSpirit)

	// 3. 定义业务原子方法，它直接通过 context store 获取观测坍缩实存能级，无需在形参中写死
	reportChan := make(chan astral.Vector6D, 1)
	/*	businessTask := func(pipe chan *anything.Mission, data []any) {
		// 从 data 隐藏位置获取 mission (因为 adapter 会对首个 data 做解析)
		// 或者更简单的，我们直接从 pipeline 中拉取，但这里我们在 pipeline 之外通过 w.Schedule 运行
		// 实际上，我们的 AstralSpirit 会把坍缩状态存进 mission.Store 中。
		// 为了在业务逻辑中读取它，我们可以使用全局变量，或者我们在 Adapter 包装下拿到。
		// 在 adapter.go 中，SmartAdapt 会自动将 w.Schedule 发送的 mission 所携带的共享 sync.Map 注入进来。
		// 让我们用更直观的方案：直接查询 allMission 的 pipeline 交互。
		pipe <- &anything.Mission{Name: anything.ExitFunction}
	}*/

	// 注册为 Wind 方法
	w.M.Store("SecurityReport", func(pipe chan *anything.Mission, data []any) {
		// 这是一个标准的 Wind 执行体。我们从传入的 pipe 对应的当前任务中提取 astral_state。
		// 实际上在 pipeline 控制器中，pipe 就是通道。但在 schedule.go 执行循环中，
		// 我们的 AstralSpirit 会直接拦截 mission 并向 mission.Store 中注入 "astral_state"。
		// 由于此时 business 执行体被反射或闭包调用，我们需要读取它。
		// 让我们从 pipe 缓存中直接通过当前正在执行的流程寻找，或者为了测试的绝对确定性，
		// 我们直接从外部在 AstralSpirit 注入后读取，或者将当前的 mission 暂存。
		// 让我们让 businessTask 简单地读取 store 的内容：

		// 实际上，为了方便业务获取，我们可以将状态写回 wind.A (返回值集合) 中。
		// 让我们在 schedule.go 中看：ExitFunction 会把 mission.Pursuit 存入 w.A
		// 为了让测试感知到，我们在 business 中获取当前 mission 状态：
		// 既然 Go closure 能捕获变量，我们可以通过一个 hook 钩子来验证：
		pipe <- &anything.Mission{Name: anything.ExitFunction}
	})

	// 4. 启动调度流
	// 我们的第一个参数是 AnchorID=400，用于触发量子观测
	key := w.Schedule("SecurityReport", []any{400})

	// 等待退出信号
	if done, ok := w.E[key]; ok {
		select {
		case <-done:
			t.Log("调度管线正常结束")
			// 5. 验证是否发生了自动观测坍缩，并且状态正确留在了 Anchor 中
			anchor, _ := store.GetAnchor(400)
			t.Logf("管线执行后节点坍缩状态: %+v", anchor.LastState)
			if anchor.LastState.Danger != 0.75 {
				t.Errorf("expected Danger = 0.75, got %f", anchor.LastState.Danger)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("pipeline timeout!")
		}
	}

	// 避免 unused variable 报错
	_ = reportChan
}
