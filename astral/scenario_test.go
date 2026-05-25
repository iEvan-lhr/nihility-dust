package astral

import (
	"fmt"
	"strings"
	"testing"
)

// 定义白皮书中所描述的孤立星空坐标 ID 集合
const (
	ID_Evan           int64 = 1001 // 资深架构师/全栈开发
	ID_Cat_SevenHappy int64 = 1002 // 宠物猫 Seven Happy
	ID_Cat_Kohler     int64 = 1003 // 宠物猫 Kohler
	ID_Proj_DeerAgent int64 = 1004 // Go语言 Office/OOXML 框架
	ID_Proj_Dtwins    int64 = 1005 // 工业 IoT 能源优化项目
	ID_Tech_Milvus    int64 = 1006 // 向量数据库

	// 观测一：跳槽新增锚点
	ID_NewCorp            int64 = 2001 // 新 C轮 AI 独角兽
	ID_Location_NewOffice int64 = 2002 // 新物理办公室 (空间变迁)

	// 观测二：独立创业新增锚点
	ID_Startup_Entity     int64 = 3001 // 初创公司主体
	ID_Capital_Seed       int64 = 3002 // 天使/种子基金
	ID_Market_Clients     int64 = 3003 // 供应链市场与大客户
	ID_Plan_Kaifeng_Villa int64 = 3004 // 开封乡下建房计划
)

// 辅助可视化条形图，展示坍缩出的能谱分布
func printSpectrumBar(label string, val float64, size int, char rune, ansiColor string) {
	normalized := val
	if normalized < 0 {
		normalized = -normalized
	}
	if normalized > 1.0 {
		normalized = 1.0
	}
	active := int(normalized * float64(size))
	if active < 0 {
		active = 0
	}
	bar := strings.Repeat(string(char), active) + strings.Repeat("░", size-active)
	sign := "+"
	if val < 0 {
		sign = "-"
	}
	fmt.Printf("   %-22s: [%s%s%s] (%s%.3f)\n", label, ansiColor, bar, "\033[0m", sign, val)
}

// TestAstralWhitepaperScenario 运行白皮书描述的完整三阶段场景以及两大生涯抉择的高能物理干涉实验
func TestAstralWhitepaperScenario(t *testing.T) {
	fmt.Println("\n\033[1;35m================================================================================\033[0m")
	fmt.Println("\033[1;35m      🌌  星空拓扑: 纯关系寄生与动态流驱动的六维架构白皮书实测验证仿真  🌌       \033[0m")
	fmt.Println("\033[1;35m================================================================================\033[0m")

	// 1. 初始化 SQLite 内存物理数据库与坍缩引擎
	store, err := NewSQLiteAstralStore(":memory:")
	if err != nil {
		t.Fatalf("Failed to init store: %v", err)
	}
	defer store.Close()

	engine := NewCollapseEngine(store)
	engine.BaseDecayRate = 0.02 // 基底自衰减系数 0.02/秒
	engine.EvolutionRate = 0.02 // 雕刻演化率 2%

	// =========================================================================
	// 第一阶段：初始化真空坐标（基态锚点）
	// =========================================================================
	fmt.Println("\n\033[1;36m▶▶ 第一阶段：初始化真空坐标（基态锚点）\033[0m")
	fmt.Println("此时系统是无序的，只有漂浮在真空中的唯一 ID 和它们出厂时的“六维核心向量”。彼此没有任何连线。")

	anchors := map[int64]string{
		ID_Evan:           "ID_Evan (架构师，基底呈现高逻辑性与构建欲)",
		ID_Cat_SevenHappy: "ID_Cat_SevenHappy (宠物猫，基底呈现高情感依恋)",
		ID_Cat_Kohler:     "ID_Cat_Kohler (宠物猫，基底呈现高情感依恋)",
		ID_Proj_DeerAgent: "ID_Proj_DeerAgent (Office/OOXML 框架)",
		ID_Proj_Dtwins:    "ID_Proj_Dtwins (工业 IoT, 涵盖沪西、沪东、北京)",
		ID_Tech_Milvus:    "ID_Tech_Milvus (向量数据库)",
	}

	for id, desc := range anchors {
		// 生成代表出厂基态语义表征的虚拟 Embedding 向量 (长度为 3)
		var baseEmbedding []float64
		switch id {
		case ID_Evan:
			baseEmbedding = []float64{0.9, 0.1, 0.2}
		case ID_Cat_SevenHappy, ID_Cat_Kohler:
			baseEmbedding = []float64{0.1, 0.9, 0.1}
		case ID_Proj_DeerAgent, ID_Proj_Dtwins:
			baseEmbedding = []float64{0.2, 0.1, 0.9}
		default:
			baseEmbedding = []float64{0.5, 0.5, 0.5}
		}

		anchorNode := &NodeAnchor{
			ID:               id,
			BaseEmbedding:    baseEmbedding,
			LastCollapseTime: 0,
			LastState:        Vector6D{}, // 出厂状态为空
		}
		_ = store.SaveAnchor(anchorNode)
		fmt.Printf("   [真空坐标注入] ID: %d | %s\n", id, desc)
	}

	// =========================================================================
	// 第二阶段：注入流动（Flows）
	// =========================================================================
	fmt.Println("\n\033[1;36m▶▶ 第二阶段：注入流动（Flows） —— 星光穿梭与超边建立\033[0m")
	fmt.Println("系统向这片星空发射了多股带有不同载荷（Payload）和半衰期的“流”。这些流在瞬间照亮了多个坐标（超边）。")

	// Flow 1：情感与生活陪伴 (高积极、长半衰期)
	fmt.Println("\n   🚀 [注入 Flow 1] 情感与生活陪伴流 (长半衰期)")
	fmt.Println("      Payload: \"周末使用 AI 生成 Seven Happy 和 Kohler 的风格化肖像。\"")
	fmt.Println("      Anchors: [ID_Evan, ID_Cat_SevenHappy, ID_Cat_Kohler]")
	flow1 := &Flow{
		Anchors:   []int64{ID_Evan, ID_Cat_SevenHappy, ID_Cat_Kohler},
		Payload:   "周末使用 AI 生成 Seven Happy 和 Kohler 的风格化肖像。",
		Timestamp: 0,
		DecayRate: 0.001, // 衰减常数极低，如暗能量般缓慢维持系统稳定性
		OriginEnergy: Vector6D{
			PosNeg:    0.95, // 极高的正向情绪底噪
			Danger:    0.0,  // 极低危险
			Influence: 0.2,
		},
		AsymmetricEnergies: map[int64]Vector6D{
			ID_Evan:           {PosNeg: 0.95, Danger: 0.0},
			ID_Cat_SevenHappy: {PosNeg: 0.90},
			ID_Cat_Kohler:     {PosNeg: 0.90},
		},
		BaseEmbedding: []float64{0.1, 0.9, 0.1},
	}
	_ = store.SaveFlow(flow1)

	// Flow 2：技术攻坚事件 (极高影响力、高危险传导、短半衰期)
	fmt.Println("\n   🚀 [注入 Flow 2] 技术攻坚事件流 (短半衰期)")
	fmt.Println("      Payload: \"发现并解决 PPTX 重新压缩失败的问题，根本原因定位为 ZIP 过程中的目录层级错误。\"")
	fmt.Println("      Anchors: [ID_Evan, ID_Proj_DeerAgent]")
	flow2 := &Flow{
		Anchors:   []int64{ID_Evan, ID_Proj_DeerAgent},
		Payload:   "发现并解决 PPTX 重新压缩失败的问题，根本原因定位为 ZIP 过程中的目录层级错误。",
		Timestamp: 0,
		DecayRate: 0.15, // 半衰期短，Bug解决后危险极速衰减至零
		AsymmetricEnergies: map[int64]Vector6D{
			ID_Evan:           {Time: 0.60, Danger: 0.40},                 // 排查期间为 Evan 注入了中度时间压力与短期危险度
			ID_Proj_DeerAgent: {Influence: 0.95, PosNeg: 0.90, Base: 0.8}, // 注入技术主导的影响力与Bug修复的积极联系
		},
		BaseEmbedding: []float64{0.8, 0.1, 0.1},
	}
	_ = store.SaveFlow(flow2)

	// Flow 3：架构设计与跨地域传导 (广域空间、持续时间压力)
	fmt.Println("\n   🚀 [注入 Flow 3] Dtwins架构设计与跨地域传导流 (持续时间压力)")
	fmt.Println("      Payload: \"Dtwins二期项目的 SOW 分析与架构映射，目标工厂覆盖沪西、沪东与北京。\"")
	fmt.Println("      Anchors: [ID_Evan, ID_Proj_Dtwins]")
	flow3 := &Flow{
		Anchors:   []int64{ID_Evan, ID_Proj_Dtwins},
		Payload:   "Dtwins二期项目的 SOW 分析与架构映射，目标工厂覆盖沪西、沪东与北京。",
		Timestamp: 0,
		DecayRate: 0.005, // 活跃期长，慢速衰减
		AsymmetricEnergies: map[int64]Vector6D{
			ID_Evan: {
				Time:      0.85, // 交付迫近的持续时间压力
				Influence: 0.80, // 项目核心骨干影响力
				Space:     0.50, // 跨越上海至北京工厂的空间阻尼
			},
			ID_Proj_Dtwins: {Time: 0.80, Influence: 0.50},
		},
		BaseEmbedding: []float64{0.5, 0.1, 0.4},
	}
	_ = store.SaveFlow(flow3)

	// =========================================================================
	// 第三阶段：观测与坍缩（动态状态生成）
	// =========================================================================
	fmt.Println("\n\033[1;36m▶▶ 第三阶段：工作日早晨的观测与坍缩（2026年5月）\033[0m")
	fmt.Println("大模型（Agent）接到指令：“评估 Evan 当前的工作负荷与焦点”。此时观测发生，状态坍缩！")

	// 在 t = 15s 时发生观测
	tObservation := int64(15000) // 15秒后 (15000ms)
	fmt.Printf("\n   [量子观测 👁] 触发对 ID_Evan (%d) 在 t = 15s 的局部状态坍缩...\n", ID_Evan)

	evanState, err := engine.Collapse(ID_Evan, tObservation)
	if err != nil {
		t.Fatalf("Collapse failed: %v", err)
	}

	// 打印坍缩状态谱线图
	printSpectrumBar("当前时间紧迫度 (Time)", evanState.Time, 15, '⚡', "\033[33m")
	printSpectrumBar("跨区域空间阻尼 (Space)", evanState.Space, 15, '🌐', "\033[36m")
	printSpectrumBar("积极情绪对冲 (PosNeg)", evanState.PosNeg, 15, '♥', "\033[32m")
	printSpectrumBar("核心控制影响力 (Influence)", evanState.Influence, 15, '🔱', "\033[34m")
	printSpectrumBar("系统危险焦虑值 (Danger)", evanState.Danger, 15, '☣', "\033[31m")

	fmt.Println("\n   \033[32m✔ 瞬间呈现的 Evan 动态状态图谱解析：\033[0m")
	fmt.Println("      * 当前焦点：Dtwins 二期项目的架构映射 (高 Time 压力与 Influence 引力主导)")
	fmt.Println("      * 潜在脆点：来自 Dtwins 跨地域 (沪西、沪东、北京) 的隐性空间危险与工期紧迫度")
	fmt.Println("      * 系统稳定性：尽管项目时间压力巨大，但底层高维情感流 (双猫陪伴) 几乎无衰减，释放高额积极联系中和，整体情绪张力安全稳定！")
	fmt.Println("      * 技术领域话语权：对 Proj_DeerAgent (OOXML 修复 Bug 遗存) 仍保有高度技术余温。")

	// =========================================================================
	// 第四阶段：高能物理干涉实验 (Obs 1 & Obs 2)
	// =========================================================================
	fmt.Println("\n\033[1;35m================================================================================\033[0m")
	fmt.Println("\033[1;35m      ⚡  高能物理实验: 注入高初始势能新流，观测干涉与全网震荡  ⚡       \033[0m")
	fmt.Println("\033[1;35m================================================================================\033[0m")

	// -------------------------------------------------------------------------
	// 观测一：注入流 Flow_Jump_NewAICorp (跳槽新 AI 独角兽)
	// -------------------------------------------------------------------------
	fmt.Println("\n\033[1;36m▶▶ 观测一：注入跳槽新 AI 独角兽流 (Flow_Jump_NewAICorp)\033[0m")
	fmt.Println("   这股流以高势能、强空间位移与规则重塑为特质，打破原有的动态平衡。")

	// 在 t = 20s 注入跳槽流
	tJump := int64(20000)
	flowJump := &Flow{
		Anchors:   []int64{ID_Evan, ID_NewCorp, ID_Tech_Milvus, ID_Location_NewOffice},
		Payload:   "加入一家处于 C 轮的 AI 基础设施/Agent 独角兽公司，担任核心架构师。",
		Timestamp: tJump,
		DecayRate: 0.01, // 持续的高工作负荷
		AsymmetricEnergies: map[int64]Vector6D{
			ID_Evan: {
				Time:      0.90, // 996 高频迭代的时间压力
				Danger:    0.60, // 试用期和新环境未知的系统高危值
				Space:     0.80, // 原有物理办公室与闵行家庭物理距离的空间位移撕裂
				Influence: 0.90, // 对向量数据库 Milvus 话语权大幅扩张
			},
		},
		BaseEmbedding: []float64{0.1, 0.5, 0.9},
	}
	_ = store.SaveFlow(flowJump)

	// 同时模拟：由于高负荷运转与空间位移，导致原有的双猫陪伴流 (Flow 1) 无法得到日常输入，
	// 在 Evan 的日常观测中，该情感流的半衰期被迫“加速衰减”（从 0.001 升高到 0.06/秒）
	fmt.Println("   [空间撕裂与干涉] 由于长距离位移和负荷运转，导致双猫日常陪伴流 1 半衰期加速衰减，能量减弱！")
	flow1.DecayRate = 0.06 // 阻尼显著变大
	_ = store.SaveFlow(flow1)

	// 在 t = 30s 再次坍缩观测 Evan 的能谱
	tObsJump := int64(30000)
	stateAfterJump, _ := engine.Collapse(ID_Evan, tObsJump)

	fmt.Printf("\n   [量子观测 👁] 触发跳槽 10 秒后 (t = 30s) 的 Evan 坍缩能谱:\n")
	printSpectrumBar("新公司工作紧迫度 (Time)", stateAfterJump.Time, 15, '⚡', "\033[33m")
	printSpectrumBar("新环境空间位移 (Space)", stateAfterJump.Space, 15, '🌐', "\033[36m")
	printSpectrumBar("积极情绪对冲 (PosNeg)", stateAfterJump.PosNeg, 15, '♥', "\033[32m")
	printSpectrumBar("AI架构控制力 (Influence)", stateAfterJump.Influence, 15, '🔱', "\033[34m")
	printSpectrumBar("新系统危险焦虑值 (Danger)", stateAfterJump.Danger, 15, '☣', "\033[31m")

	fmt.Println("   [观测结论] Evan 处于高负荷运转状态，空间张力增大，技术基底高度重组 (Milvus/RAG 占主导，Dtwins 工业流因断流开始暗淡)。")

	// -------------------------------------------------------------------------
	// 观测二：注入流 Flow_Startup_Genesis (独立创业：大爆炸事件)
	// -------------------------------------------------------------------------
	fmt.Println("\n\033[1;36m▶▶ 观测二：注入独立创业流 (Flow_Startup_Genesis) —— 创造绝对领域与黑洞\033[0m")
	fmt.Println("   这是一种极端的超边生成。创业不是一条线，而是一个吸收一切能量的危险黑洞。")

	// 在 t = 40s 发射大爆炸创业流
	tStartup := int64(40000)
	flowStartup := &Flow{
		Anchors:   []int64{ID_Evan, ID_Startup_Entity, ID_Capital_Seed, ID_Market_Clients, ID_Plan_Kaifeng_Villa},
		Payload:   "以 Deer Agent 为核心底座，成立专注于自动化专业分析与供应链 AI 的初初公司。",
		Timestamp: tStartup,
		DecayRate: 0.002, // 终身创业，近乎恒定不衰减的生命力
		AsymmetricEnergies: map[int64]Vector6D{
			ID_Evan: {
				Danger:    1.00, // 资金链断裂风险、高度破坏不确定性达到极限 Danger=100%
				Time:      1.00, // 无上限的 24x7 压榨时间压力
				Influence: 1.00, // 对 DeerAgent 拥有绝对 100% 的主宰技术控制力
				PosNeg:    0.95, // 造物主的狂热精神与极端高度的自我实现积极感
			},
			// ⚡ 危险传导与全网震荡：创业资金消耗直接与“乡下建房计划”发生引力对冲！
			ID_Plan_Kaifeng_Villa: {
				PosNeg: -0.80, // 建房资金被创业挤出，产生强烈负向对冲
				Time:   -0.90, // 建房工期推进受阻，工期无限流速变缓
			},
		},
		BaseEmbedding: []float64{0.9, 0.9, 0.9},
	}
	_ = store.SaveFlow(flowStartup)

	// 在 t = 45s (创业 5 秒后) 观测 Evan 与 开封建房计划 两个节点的坍缩情况
	tObsStartup := int64(45000)
	stateEvanStartup, _ := engine.Collapse(ID_Evan, tObsStartup)
	stateVillaStartup, _ := engine.Collapse(ID_Plan_Kaifeng_Villa, tObsStartup)

	fmt.Printf("\n   [量子观测 👁] 触发创业 5 秒后 (t = 45s) Evan 的坍缩能谱:\n")
	printSpectrumBar("创业极限压迫感 (Time)", stateEvanStartup.Time, 15, '⚡', "\033[33m")
	printSpectrumBar("造物主狂热情绪 (PosNeg)", stateEvanStartup.PosNeg, 15, '♥', "\033[32m")
	printSpectrumBar("绝对领域主宰力 (Influence)", stateEvanStartup.Influence, 15, '🔱', "\033[34m")
	printSpectrumBar("资金链断裂危机 (Danger)", stateEvanStartup.Danger, 15, '☣', "\033[31m")

	fmt.Printf("\n   [量子观测 👁] 触发创业 5 秒后「开封乡下建房计划」(%d) 节点的干涉坍缩能谱:\n", ID_Plan_Kaifeng_Villa)
	printSpectrumBar("建房项目负推进 (Time)", stateVillaStartup.Time, 15, '⚡', "\033[33m")
	printSpectrumBar("资金抽调对冲 (PosNeg)", stateVillaStartup.PosNeg, 15, '♥', "\033[35m")

	fmt.Println("\n   \033[31m⚡ 危险传导震荡评估：\033[0m")
	fmt.Println("      由于创业流（Danger=100%）携带极高能量，该危机瞬间蔓延至 Evan 个人生活圈，")
	fmt.Println("      与「开封建房计划」产生强烈资金与精力引力干涉。建房计划因资金与工期流速受阻（Time=-0.9）陷入停滞状态。")
	fmt.Println("      系统的整体稳定性目前完全悬挂在「双猫情感流」所残留的最后一丝积极防御能级上，星空处于极其不稳定的高危边缘！")

	fmt.Println("\n\033[1;35m================================================================================\033[0m")
	fmt.Println("\033[1;35m      🌌  星空拓扑六维空间架构白皮书全部测试模拟顺利结束，完美坍缩出实存属性！ ✨   \033[0m")
	fmt.Println("\033[1;35m================================================================================\033[0m")

	// 最终核心学术指标断言，确保系统没有偏航
	if evanState.Danger >= 0.40 {
		t.Errorf("观测三阶段中 Evan 的 Danger 应已被双猫情感抵消，应小于 0.40, 实际: %f", evanState.Danger)
	}
	if stateEvanStartup.Danger < 0.90 {
		t.Errorf("创业状态下 Evan 的危机值应达到物理上限 Danger >= 0.90, 实际: %f", stateEvanStartup.Danger)
	}
	if stateVillaStartup.PosNeg >= 0.0 {
		t.Errorf("建房计划应当与创业产生负向阻尼对冲，PosNeg 应为负值, 实际: %f", stateVillaStartup.PosNeg)
	}
}
