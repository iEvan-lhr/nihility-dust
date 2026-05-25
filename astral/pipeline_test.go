package astral

import (
	"context"
	"encoding/json"
	"math"
	"strings"
	"testing"
	"time"
)

// TestAstralPipeline_FullWorkflow 验证 Ingest 坐标、发射 Flows 并通过新接口坍缩为通用 Context
func TestAstralPipeline_FullWorkflow(t *testing.T) {
	ctx := context.Background()

	store, err := NewSQLiteAstralStore(":memory:")
	if err != nil {
		t.Fatalf("failed to open store: %v", err)
	}
	defer store.Close()

	engine := NewCollapseEngine(store)
	engine.BaseDecayRate = 0.02

	encoder := NewSimpleMockEncoder(3)
	pipeline := NewAstralPipeline(store, engine, encoder)

	// 1. 注册基态节点 (ID 自定义)
	t.Log("--- 注册基态节点 ---")
	evanNode, err := pipeline.RegisterEntity(ctx, ID_Evan, "Evan, 资深 Go 架构师")
	if err != nil {
		t.Fatalf("failed to register evan node: %v", err)
	}
	t.Logf("注册成功，ID: %d, 描述: %s", evanNode.ID, evanNode.Description)

	// 2. 发射事件
	_, err = pipeline.EmitEvent(
		ctx,
		[]int64{ID_Evan},
		"进行日常维护",
		0.1,
		Vector6D{Danger: 0.2},
		nil,
	)
	if err != nil {
		t.Fatalf("failed to emit event: %v", err)
	}

	// 3. 量子坍缩观测
	collapsedCtx, err := pipeline.CollapseToContext(ctx, ID_Evan)
	if err != nil {
		t.Fatalf("failed to collapse to context: %v", err)
	}

	// 序列化为通用 JSON
	jsonBytes, err := json.MarshalIndent(collapsedCtx, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal collapsed context: %v", err)
	}

	t.Logf("这一微秒坍缩输出的通用时空上下文 JSON:\n%s", string(jsonBytes))

	// 断言验证
	if collapsedCtx.AnchorID != ID_Evan {
		t.Errorf("expected AnchorID = %d, got %d", ID_Evan, collapsedCtx.AnchorID)
	}
	if len(collapsedCtx.ActiveEvents) != 1 {
		t.Errorf("expected 1 active event, got %d", len(collapsedCtx.ActiveEvents))
	}
}

// TestAstralPipeline_TempIDMapping_AutoSnowflake 仿真测试：验证外部大模型生成的临时 ID 能被系统零摩擦动态映射为雪花唯一 ID 并完成物理持久化
func TestAstralPipeline_TempIDMapping_AutoSnowflake(t *testing.T) {
	ctx := context.Background()

	store, err := NewSQLiteAstralStore(":memory:")
	if err != nil {
		t.Fatalf("failed to open store: %v", err)
	}
	defer store.Close()

	engine := NewCollapseEngine(store)
	encoder := NewSimpleMockEncoder(8)
	pipeline := NewAstralPipeline(store, engine, encoder)

	// 模拟外部大模型对原始文本（例如 LOTO 步骤）进行语义理解后，产生的通用 JSON。
	// 大模型不需要知道物理 ID，它使用临时符号（如 "$NEW_VALVE"）进行拓扑超边描述，并做去重。
	type LLMExtractedJSON struct {
		NewAnchors []struct {
			TempID string `json:"temp_id"`
			Desc   string `json:"desc"`
		} `json:"new_anchors"`
		Flows []struct {
			Payload            string              `json:"payload"`
			Anchors            []string            `json:"anchors"`
			DecayRate          float64             `json:"decay_rate"`
			Energy             Vector6D            `json:"energy"`
			AsymmetricEnergies map[string]Vector6D `json:"asymmetric_energies"`
		} `json:"flows"`
	}

	// 模拟大模型输出的文本 JSON
	rawLLMOutput := `{
		"new_anchors": [
			{"temp_id": "$NEW_MACHINE", "desc": "L13一次包装1#包装机"},
			{"temp_id": "$NEW_VALVE", "desc": "1#包装机锁定电能与气能隔离点"}
		],
		"flows": [
			{
				"payload": "执行LOTO上锁步骤4：关闭气源释放余气，泄压到0",
				"anchors": ["$NEW_MACHINE", "$NEW_VALVE"],
				"decay_rate": 0.15,
				"energy": {"danger": 0.0, "time": 0.0},
				"asymmetric_energies": {
					"$NEW_MACHINE": {"danger": -0.40, "time": -0.20},
					"$NEW_VALVE": {"influence": 0.60}
				}
			}
		]
	}`

	// --- 外部大模型逻辑开始（由 Go Pipeline 提供纯数据层面的 ID 映射与持久化） ---
	var parsed LLMExtractedJSON
	if err := json.Unmarshal([]byte(rawLLMOutput), &parsed); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// 1. 自动收集所有临时 ID 并提取 Snowflake 全局唯一 ID 映射表
	var tempIDs []string
	for _, a := range parsed.NewAnchors {
		tempIDs = append(tempIDs, a.TempID)
	}
	tempToRealMapping := pipeline.GenerateTempIDMapping(tempIDs)

	// 2. 将新注册的节点持久化导入 `dust.db`
	for _, na := range parsed.NewAnchors {
		realID := tempToRealMapping[na.TempID]
		_, err := pipeline.RegisterEntity(ctx, realID, na.Desc)
		if err != nil {
			t.Fatalf("failed to register entity: %v", err)
		}
	}

	// 3. 将流动 Flow 发射并转换临时 ID 连线
	for _, f := range parsed.Flows {
		var resolvedAnchors []int64
		for _, aStr := range f.Anchors {
			if strings.HasPrefix(aStr, "$") {
				resolvedAnchors = append(resolvedAnchors, tempToRealMapping[aStr])
			}
		}

		resolvedAsymmetric := make(map[int64]Vector6D)
		for aStr, energy := range f.AsymmetricEnergies {
			if strings.HasPrefix(aStr, "$") {
				resolvedAsymmetric[tempToRealMapping[aStr]] = energy
			}
		}

		_, err = pipeline.EmitEvent(ctx, resolvedAnchors, f.Payload, f.DecayRate, f.Energy, resolvedAsymmetric)
		if err != nil {
			t.Fatalf("failed to emit: %v", err)
		}
	}

	// --- 4. 验证数据物理入库与半衰对冲 ---
	realMachineID := tempToRealMapping["$NEW_MACHINE"]
	anchor, err := store.GetAnchor(realMachineID)
	if err != nil {
		t.Fatalf("node not persisted: %v", err)
	}
	t.Logf("★ [自动 Snowflake ID 验证] $NEW_MACHINE ➔ 物理 ID: %d, 描述: %s", realMachineID, anchor.Description)

	// 坍缩验证
	state, _ := engine.Collapse(realMachineID, time.Now().UnixNano()/1e6)
	t.Logf("★ [能能对冲验证] 机器当前坍缩 Danger 值: %.4f (成功受到步骤 4 的负能对冲！)", state.Danger)
	if math.Abs(state.Danger-0.00) > 1e-9 {
		t.Errorf("expected Danger = 0.00 (clamped from -0.40), got %f", state.Danger)
	}
}

// TestAstralPipeline_SearchKnowledge 验证通过高维语义检索，提取相关的事件文本载荷作为 AI 上下文知识
func TestAstralPipeline_SearchKnowledge(t *testing.T) {
	ctx := context.Background()

	store, err := NewSQLiteAstralStore(":memory:")
	if err != nil {
		t.Fatalf("failed to open store: %v", err)
	}
	defer store.Close()

	engine := NewCollapseEngine(store)
	encoder := NewSimpleMockEncoder(32) // 使用 32 维特征哈希
	pipeline := NewAstralPipeline(store, engine, encoder)

	// 1. 发射两条语义迥异的流动事件 (承载不同的知识文本)
	payloadLOTO := "L13一次包装1#包装机LOTO指引，操作步骤第一步：确认电能，气能锁定点并通知所有相关人员"
	payloadCat := "宠物猫 Seven Happy 的日常陪伴：喂食猫粮与修剪指甲"

	_, err = pipeline.EmitEvent(ctx, []int64{1001}, payloadLOTO, 0.02, Vector6D{Danger: 0.1}, nil)
	if err != nil {
		t.Fatalf("failed to emit flow 1: %v", err)
	}

	_, err = pipeline.EmitEvent(ctx, []int64{1002}, payloadCat, 0.01, Vector6D{PosNeg: 0.8}, nil)
	if err != nil {
		t.Fatalf("failed to emit flow 2: %v", err)
	}

	// 2. 使用自然语言搜索“包装机 LOTO”
	t.Log("--- 检索自然语言 query: \"包装机 LOTO\" ---")
	items, err := pipeline.SearchKnowledge(ctx, "包装机 LOTO", 2)
	if err != nil {
		t.Fatalf("failed to search knowledge: %v", err)
	}

	if len(items) != 2 {
		t.Fatalf("expected 2 search results, got %d", len(items))
	}

	// 验证排在第一位的是更相关的 LOTO 指引流
	t.Logf("检索结果 1 (相似度 %.4f): %q", items[0].Similarity, items[0].Payload)
	t.Logf("检索结果 2 (相似度 %.4f): %q", items[1].Similarity, items[1].Payload)

	if !strings.Contains(items[0].Payload, "包装机") {
		t.Errorf("expected top result to be the LOTO manual, got %q", items[0].Payload)
	}

	if items[0].Similarity <= items[1].Similarity {
		t.Errorf("expected top result similarity (%f) to be strictly greater than second result similarity (%f)",
			items[0].Similarity, items[1].Similarity)
	}
}
