package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/iEvan-lhr/go-llm-client/client"
	"github.com/iEvan-lhr/go-llm-client/llm"
	"github.com/iEvan-lhr/nihility-dust/astral"
)

// ANSI color escape codes for high aesthetic console rendering
const (
	ColorReset  = "\033[0m"
	ColorCyan   = "\033[1;36m"
	ColorGreen  = "\033[1;32m"
	ColorYellow = "\033[1;33m"
	ColorRed    = "\033[1;31m"
	ColorPurple = "\033[1;35m"
	ColorBlue   = "\033[1;34m"
)

// LLMExtractedJSON 大模型提炼提取的拓扑数据映射结构
type LLMExtractedJSON struct {
	NewAnchors []struct {
		TempID string `json:"temp_id"`
		Desc   string `json:"desc"`
	} `json:"new_anchors"`
	Flows []struct {
		Payload            string                     `json:"payload"`
		Anchors            []string                   `json:"anchors"`
		DecayRate          float64                    `json:"decay_rate"`
		Energy             astral.Vector6D            `json:"energy"`
		AsymmetricEnergies map[string]astral.Vector6D `json:"asymmetric_energies"`
	} `json:"flows"`
}

// SystemPrompt 无特定演示实体干扰的纯净星空拓扑六维场提炼提取 System Prompt
const SystemPrompt = `你是一个高维度拓扑空间物理场分析专家。你的任务是将输入的文本（如规程、手册或日志）映射为星空拓扑的 JSON 结构。
你需要提炼出：
1. new_anchors (基态节点坐标)：识别出核心系统、重点设备、人物角色或关键状态本体。新节点必须分配一个以 "$" 开头的唯一临时ID，格式为 "$NEW_KEYWORD"（例如 "$NEW_VALVE"）。
2. flows (流动事件超边)：识别出发生的事件或具体步骤文本作为 Payload。指定关联的 anchors 数组（使用临时 ID 或已知的真实ID），并在 energy 中指定六维初始能谱（danger/pos_neg/time/influence/space/base），打分范围 [-1.0, 1.0]。在 asymmetric_energies 中，为不同受力节点指定非对称能谱值。

必须仅返回符合以下 Schema 的 JSON 字符串，严禁包含任何 Markdown 标记或旁白：
{
  "new_anchors": [{"temp_id": "$NEW_PUMP", "desc": "本体描述"}],
  "flows": [{
    "payload": "事件文本/操作片段",
    "anchors": ["$NEW_PUMP"],
    "decay_rate": 0.02,
    "energy": {"danger": 0.1, "pos_neg": 0.0, "time": 0.2, "influence": 0.5, "space": 0.0, "base": 0.0},
    "asymmetric_energies": {"$NEW_PUMP": {"danger": 0.2}}
  }]
}`

// loadEnv 自动解析 .env 文件并注入到系统环境变量中 (纯 Go 无外部依赖实现)
func loadEnv() {
	content, err := os.ReadFile(".env")
	if err != nil {
		return
	}
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			k := strings.TrimSpace(parts[0])
			v := strings.TrimSpace(parts[1])
			v = strings.Trim(v, `"'`)
			os.Setenv(k, v)
		}
	}
}

// cleanJSON 从大模型的响应中，安全剔除 Markdown 的 ```json 包装标记
func cleanJSON(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		if idx := strings.Index(s, "\n"); idx != -1 {
			s = s[idx+1:]
		}
		if idx := strings.LastIndex(s, "```"); idx != -1 {
			s = s[:idx]
		}
		s = strings.TrimSpace(s)
	}
	return s
}

// Print helper for console meters
func renderMeter(label string, val float64, maxStars int, char rune, color string) {
	normalized := val
	if normalized < 0 {
		normalized = -normalized
	}
	if normalized > 1.0 {
		normalized = 1.0
	}
	active := int(normalized * float64(maxStars))
	if active < 0 {
		active = 0
	}
	bar := strings.Repeat(string(char), active) + strings.Repeat("░", maxStars-active)
	sign := "+"
	if val < 0 {
		sign = "-"
	}
	fmt.Printf("   %-24s: [%s%s%s] (%s%.3f)\n", label, color, bar, ColorReset, sign, val)
}

func printHelp() {
	fmt.Println(ColorCyan + "\n================================================================================")
	fmt.Println("         🌌  星空拓扑六维空间物理场命令行工具 (Astral CLI)  🌌")
	fmt.Println("================================================================================" + ColorReset)
	fmt.Println("本工具使用本地持久化数据库 " + ColorYellow + "dust.db" + ColorReset + "，完全支持自主添加节点内容与物理仿真测试。")
	fmt.Println("\n使用方法 (Commands):")
	fmt.Println(ColorGreen + "  1. 注册基态锚点 (Register Anchor):" + ColorReset)
	fmt.Println("     - [自动分配唯一 ID]: go run astral_cli.go register <本体语义描述>")
	fmt.Println("     - [指定特定 ID]: go run astral_cli.go register <ID> <本体语义描述>")

	fmt.Println(ColorGreen + "\n  2. 查询当前所有的基态锚点 (List Anchors):" + ColorReset)
	fmt.Println("     go run astral_cli.go list")

	fmt.Println(ColorGreen + "\n  3. 发射流动事件 (Emit Flow):" + ColorReset)
	fmt.Println("     go run astral_cli.go emit <IDs_逗号分隔> <事件载荷/文档> <衰减率> <Danger> <PosNeg> <Time> <Influence> [Space]")

	fmt.Println(ColorGreen + "\n  4. 观测坍缩绝对状态 (Collapse Anchor):" + ColorReset)
	fmt.Println("     go run astral_cli.go collapse <ID>")

	fmt.Println(ColorGreen + "\n  5. 引力波语义搜索 (RAG Semantic Search):" + ColorReset)
	fmt.Println("     go run astral_cli.go search <检索文本描述>")

	fmt.Println(ColorGreen + "\n  6. 大模型智能拓扑提取 (AI LLM Extract - ★新增):" + ColorReset)
	fmt.Println("     go run astral_cli.go llm-extract <文本内容或文本文件路径>")

	fmt.Println(ColorGreen + "\n  7. 大模型物理场 RAG 智能对话 (AI LLM Chat - ★新增):" + ColorReset)
	fmt.Println("     go run astral_cli.go llm-chat <智能提问，自动结合检索到的上下文>")

	fmt.Println(ColorCyan + "================================================================================" + ColorReset)
}

func main() {
	if len(os.Args) < 2 {
		printHelp()
		return
	}

	// 初始化持久化 SQLite 数据库 (自动创建/载入当前目录下的 dust.db)
	store, err := astral.NewSQLiteAstralStore("dust.db")
	if err != nil {
		fmt.Printf(ColorRed+"[数据库初始化失败] Error: %v\n"+ColorReset, err)
		return
	}
	defer store.Close()

	engine := astral.NewCollapseEngine(store)
	engine.BaseDecayRate = 0.02
	engine.EvolutionRate = 0.03

	encoder := astral.NewSimpleMockEncoder(128)
	ctx := context.Background()

	pipeline := astral.NewAstralPipeline(store, engine, encoder)

	command := strings.ToLower(os.Args[1])

	switch command {
	case "register":
		if len(os.Args) < 3 {
			fmt.Println(ColorRed + "错误: register 命令参数不足" + ColorReset)
			fmt.Println("格式: go run astral_cli.go register <本体语义描述>  或者  go run astral_cli.go register <指定ID> <描述>")
			return
		}

		var id int64
		var description string

		// 判断是自动分配 ID 还是指定特定 ID
		if len(os.Args) == 3 {
			// 自动分配唯一 ID (传入 0)
			id = 0
			description = os.Args[2]
		} else {
			// 指定特定 ID
			parsedID, err := strconv.ParseInt(os.Args[2], 10, 64)
			if err != nil {
				fmt.Printf(ColorRed+"错误: 指定的 ID 必须是整数, 得到 %q\n"+ColorReset, os.Args[2])
				return
			}
			id = parsedID
			description = os.Args[3]
		}

		// 写入持久化库 (若 id 为 0，RegisterEntity 会自动使用 Snowflake 分配唯一 ID)
		anchor, err := pipeline.RegisterEntity(ctx, id, description)
		if err != nil {
			fmt.Printf(ColorRed+"[注册失败] %v\n"+ColorReset, err)
			return
		}

		fmt.Printf(ColorGreen+"[✔ 注册成功]"+ColorReset+" 坐标 ID: %d 已打入星空拓扑，持久化存于 dust.db。\n", anchor.ID)
		fmt.Printf("   -> 本体语义: %q\n", anchor.Description)

	case "list":
		// 获取持久化库中所有已注册的基态锚点，并显示详细信息
		anchors, err := store.GetAllAnchors()
		if err != nil {
			fmt.Printf(ColorRed+"[查询失败] %v\n"+ColorReset, err)
			return
		}

		fmt.Println(ColorCyan + "\n--- 🌌 [当前已存真空坐标列表 (Existing Anchors)] ---" + ColorReset)
		if len(anchors) == 0 {
			fmt.Println("   [真空状态] 目前没有注册任何基态节点坐标。您可以使用 register 命令添加节点！")
		} else {
			for _, a := range anchors {
				fmt.Printf("  📍 ID: %-19d | 描述: %s\n", a.ID, a.Description)
			}
		}
		fmt.Println()

	case "emit":
		if len(os.Args) < 9 {
			fmt.Println(ColorRed + "错误: emit 命令参数不足" + ColorReset)
			fmt.Println("格式: go run astral_cli.go emit <IDs_逗号分隔> <事件载荷> <衰减率> <Danger> <PosNeg> <Time> <Influence> [Space]")
			return
		}

		idStrParts := strings.Split(os.Args[2], ",")
		var anchors []int64
		for _, s := range idStrParts {
			id, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
			if err != nil {
				fmt.Printf(ColorRed+"错误: 无效的节点 ID %q\n"+ColorReset, s)
				return
			}
			anchors = append(anchors, id)
		}

		payload := os.Args[3]
		decayRate, _ := strconv.ParseFloat(os.Args[4], 64)
		danger, _ := strconv.ParseFloat(os.Args[5], 64)
		posNeg, _ := strconv.ParseFloat(os.Args[6], 64)
		timeVal, _ := strconv.ParseFloat(os.Args[7], 64)
		influence, _ := strconv.ParseFloat(os.Args[8], 64)
		spaceVal := 0.0
		if len(os.Args) >= 10 {
			spaceVal, _ = strconv.ParseFloat(os.Args[9], 64)
		}

		embedding, err := encoder.Encode(ctx, payload)
		if err != nil {
			fmt.Printf(ColorRed+"[文本载荷编码失败] %v\n"+ColorReset, err)
			return
		}

		energy := astral.Vector6D{
			Danger:    danger,
			PosNeg:    posNeg,
			Time:      timeVal,
			Influence: influence,
			Space:     spaceVal,
		}

		flow := &astral.Flow{
			Anchors:       anchors,
			Payload:       payload,
			Timestamp:     time.Now().UnixNano() / 1e6,
			OriginEnergy:  energy,
			DecayRate:     decayRate,
			BaseEmbedding: embedding,
		}

		if err := store.SaveFlow(flow); err != nil {
			fmt.Printf(ColorRed+"[流动发射失败] %v\n"+ColorReset, err)
			return
		}

		fmt.Printf(ColorGreen+"[🚀 流动发射成功]"+ColorReset+" 一股新事件波流已射入星空，持久化关联节点: %v\n", anchors)
		fmt.Printf("   -> 事件载荷: %q\n", payload)
		fmt.Printf("   -> 初始能谱: Danger=%.2f, PosNeg=%+.2f, Time=%.2f, Influence=%.2f, Space=%.2f (半衰期率: %.4f)\n",
			danger, posNeg, timeVal, influence, spaceVal, decayRate)

	case "collapse":
		if len(os.Args) < 3 {
			fmt.Println(ColorRed + "错误: collapse 命令需要指定节点 ID" + ColorReset)
			return
		}

		id, err := strconv.ParseInt(os.Args[2], 10, 64)
		if err != nil {
			fmt.Printf(ColorRed+"错误: 无效的节点 ID %q\n"+ColorReset, os.Args[2])
			return
		}

		nowMs := time.Now().UnixNano() / 1e6

		anchor, err := store.GetAnchor(id)
		if err != nil {
			fmt.Printf(ColorRed+"[未找到节点快照] %v\n"+ColorReset, err)
			return
		}

		fmt.Printf("\n"+ColorCyan+"[量子观测 👁] 触发对节点 %s (ID: %d) 在当前微秒的绝对能级坍缩..."+ColorReset+"\n", anchor.Description, id)
		state, err := engine.Collapse(id, nowMs)
		if err != nil {
			fmt.Printf(ColorRed+"[坍缩计算失败] %v\n"+ColorReset, err)
			return
		}

		renderMeter("时间紧迫压力 (Time)", state.Time, 15, '⚡', ColorYellow)
		renderMeter("跨地域空间阻尼 (Space)", state.Space, 15, '🌐', ColorCyan)
		renderMeter("积极情绪防护 (PosNeg)", state.PosNeg, 15, '♥', ColorGreen)
		renderMeter("核心支配影响力 (Influence)", state.Influence, 15, '🔱', ColorBlue)
		renderMeter("系统断裂危险度 (Danger)", state.Danger, 15, '☣', ColorRed)

		flows, _ := store.GetActiveFlowsForAnchor(id, 0)
		activeCount := 0
		fmt.Println(ColorCyan + "\n   【当前照亮该节点的未衰减事件流】:" + ColorReset)
		for _, f := range flows {
			dtSec := float64(nowMs-f.Timestamp) / 1000.0
			energyRatio := math.Exp(-f.DecayRate * dtSec)
			if energyRatio < 0.02 {
				continue
			}
			activeCount++
			fmt.Printf("     %d. 事件: %q\n        | 时差: %.1f秒前 | 剩余余温能量: %.1f%%\n",
				activeCount, f.Payload, dtSec, energyRatio*100)
		}
		if activeCount == 0 {
			fmt.Println("     [空无一物] 该节点在时空中目前处于纯粹的真空基态，无未衰减的流动穿过。")
		}
		fmt.Println()

	case "search":
		if len(os.Args) < 3 {
			fmt.Println(ColorRed + "错误: search 命令需要提供检索文本" + ColorReset)
			return
		}

		query := os.Args[2]
		embedding, err := encoder.Encode(ctx, query)
		if err != nil {
			fmt.Printf(ColorRed+"[检索向量转化失败] %v\n"+ColorReset, err)
			return
		}

		fmt.Printf("\n"+ColorCyan+"[引力波检索 📡] 正在全网检索与 %q 语义关联的 Top-3 节点坐标..."+ColorReset+"\n", query)
		results, err := store.SearchByRelativity(embedding, 3)
		if err != nil {
			fmt.Printf(ColorRed+"[检索失败] %v\n"+ColorReset, err)
			return
		}

		if len(results) == 0 {
			fmt.Println("   未检索到任何拥有有效引力向量的基态节点坐标。")
		} else {
			for idx, r := range results {
				fmt.Printf("   📍 [%d] 相似度: %.4f | ID: %d | %s\n", idx+1, r.Similarity, r.Anchor.ID, r.Anchor.Description)
				fmt.Printf("       | 坍缩能级: Danger=%.2f, PosNeg=%+.2f, Time=%.2f, Space=%.2f, Influence=%.2f\n",
					r.Anchor.LastState.Danger, r.Anchor.LastState.PosNeg, r.Anchor.LastState.Time, r.Anchor.LastState.Space, r.Anchor.LastState.Influence)
			}
		}

		fmt.Printf("\n" + ColorCyan + "[知识库检索 📚] 正在检索最相关的 Top-3 事件文本/文档 (供 AI 上下文知识)..." + ColorReset + "\n")
		knowledgeItems, err := pipeline.SearchKnowledge(ctx, query, 3)
		if err != nil {
			fmt.Printf(ColorRed+"[知识检索失败] %v\n"+ColorReset, err)
			return
		}

		if len(knowledgeItems) == 0 {
			fmt.Println("   未检索到任何相关的知识或事件文档。")
		} else {
			for idx, item := range knowledgeItems {
				fmt.Printf("   💡 [%d] 语义相似度: %.4f | FlowID: %d | 关联节点: %v\n", idx+1, item.Similarity, item.FlowID, item.Anchors)
				fmt.Println(ColorGreen + "       >>> 检索到的 AI 上下文知识载荷: >>>" + ColorReset)
				lines := strings.Split(item.Payload, "\n")
				for _, line := range lines {
					if strings.TrimSpace(line) != "" {
						fmt.Printf("           %s\n", line)
					}
				}
				fmt.Println(ColorGreen + "       <<< ========================== <<<" + ColorReset)
			}
		}
		fmt.Println()

	case "llm-extract":
		if len(os.Args) < 3 {
			fmt.Println(ColorRed + "错误: llm-extract 命令需要提供文本内容或文本文件路径" + ColorReset)
			return
		}

		inputText := os.Args[2]
		// 尝试作为文件读取，如果文件存在则使用文件内容
		if _, err := os.Stat(inputText); err == nil {
			content, err := os.ReadFile(inputText)
			if err == nil {
				inputText = string(content)
			}
		}

		// 1. 初始化大模型客户端
		loadEnv()
		apiKey := os.Getenv("DASHSCOPE_API_KEY")
		if apiKey == "" {
			apiKey = os.Getenv("QWEN_API_KEY")
		}
		if apiKey == "" {
			fmt.Println(ColorRed + "错误: 未在 .env 或环境变量中找到 QWEN_API_KEY 或 DASHSCOPE_API_KEY" + ColorReset)
			return
		}

		modelName := os.Getenv("QWEN_MODEL")
		if modelName == "" {
			modelName = "qwen-plus"
		}

		fmt.Printf("\n"+ColorCyan+"[大模型交互 🤖] 正在使用模型 %q 进行高维拓扑与能谱关系提取..."+ColorReset+"\n", modelName)

		c, err := client.New(llm.Config{
			Provider: "dashscope",
			Model:    modelName,
			APIKey:   apiKey,
		})
		if err != nil {
			fmt.Printf(ColorRed+"[初始化大模型失败] %v\n"+ColorReset, err)
			return
		}

		// 构建提取 prompt
		prompt := fmt.Sprintf("%s\n\n请提取以下文本：\n%s", SystemPrompt, inputText)
		fmt.Println(prompt)
		resp, err := c.SendNoHistory(ctx, prompt)
		if err != nil {
			fmt.Printf(ColorRed+"[大模型提取失败] %v\n"+ColorReset, err)
			return
		}
		responseText := resp.Message.Content
		fmt.Println(responseText)
		// 提取并清洗 JSON
		jsonStr := cleanJSON(responseText)
		fmt.Println(jsonStr)
		// 反序列化并自动入库
		var extracted LLMExtractedJSON
		if err := json.Unmarshal([]byte(jsonStr), &extracted); err != nil {
			fmt.Printf(ColorRed+"[JSON 解析失败] 原始响应中未找到合法 JSON，或格式错误。\n报错: %v\n"+ColorReset, err)
			fmt.Println("大模型原始输出:")
			fmt.Println(responseText)
			return
		}

		// 收集所有以 $ 开头的临时 ID 并分配 Snowflake ID 映射
		var tempIDs []string
		for _, na := range extracted.NewAnchors {
			tempIDs = append(tempIDs, na.TempID)
		}
		idMapping := pipeline.GenerateTempIDMapping(tempIDs)

		fmt.Println(ColorGreen + "\n--- 📍 自动物理坐标注册 (Auto Snowflake ID mapping) ---" + ColorReset)
		for _, na := range extracted.NewAnchors {
			realID := idMapping[na.TempID]
			_, err := pipeline.RegisterEntity(ctx, realID, na.Desc)
			if err != nil {
				fmt.Printf(ColorRed+"    [注册失败] %s: %v\n"+ColorReset, na.TempID, err)
				return
			}
			fmt.Printf("   📍 [新坐标注入] %s ➔ 物理 ID: %d | 描述: %s\n", na.TempID, realID, na.Desc)
		}

		fmt.Println(ColorGreen + "\n--- 🚀 自动事件流动发射 (Auto Flow Emission) ---" + ColorReset)
		for _, f := range extracted.Flows {
			var resolvedAnchors []int64
			for _, aStr := range f.Anchors {
				if strings.HasPrefix(aStr, "$") {
					resolvedAnchors = append(resolvedAnchors, idMapping[aStr])
				} else {
					if id, err := strconv.ParseInt(aStr, 10, 64); err == nil {
						resolvedAnchors = append(resolvedAnchors, id)
					}
				}
			}

			resolvedAsymmetric := make(map[int64]astral.Vector6D)
			for aStr, energy := range f.AsymmetricEnergies {
				if strings.HasPrefix(aStr, "$") {
					resolvedAsymmetric[idMapping[aStr]] = energy
				} else {
					if id, err := strconv.ParseInt(aStr, 10, 64); err == nil {
						resolvedAsymmetric[id] = energy
					}
				}
			}

			flow, err := pipeline.EmitEvent(ctx, resolvedAnchors, f.Payload, f.DecayRate, f.Energy, resolvedAsymmetric)
			if err != nil {
				fmt.Printf(ColorRed+"    [发射失败] Payload: %q | 错误: %v\n"+ColorReset, f.Payload, err)
				return
			}
			fmt.Printf("   🚀 [流动发射] FlowID: %d | 照亮节点: %v\n      | 载荷: %q\n      | 初始能谱: Danger=%.2f, PosNeg=%+.2f, Time=%.2f\n",
				flow.ID, resolvedAnchors, f.Payload, f.Energy.Danger, f.Energy.PosNeg, f.Energy.Time)
		}

		fmt.Println(ColorCyan + "\n[✔ AI 拓扑关系注入完毕] 所有节点坐标与流动事件已成功持久化存于 dust.db。\n" + ColorReset)

	case "llm-chat":
		if len(os.Args) < 3 {
			fmt.Println(ColorRed + "错误: llm-chat 命令需要提供问答内容" + ColorReset)
			return
		}

		query := os.Args[2]

		// 1. 从本地物理场执行 RAG 检索，提取最相关的 Top-3 事件文档作为上下文
		fmt.Printf(ColorCyan+"[物理检索 📡] 正在从本地星空场匹配与 %q 最相关的时空上下文记忆..."+ColorReset+"\n", query)
		knowledgeItems, err := pipeline.SearchKnowledge(ctx, query, 3)
		if err != nil {
			fmt.Printf(ColorRed+"[检索记忆失败] %v\n"+ColorReset, err)
			return
		}

		var contextSB strings.Builder
		if len(knowledgeItems) == 0 {
			contextSB.WriteString("- 目前本地物理场中没有匹配到相关的活跃流文本记录。\n")
		} else {
			for idx, item := range knowledgeItems {
				contextSB.WriteString(fmt.Sprintf("- 记忆事件 [%d] (引力相似度: %.4f): %s\n", idx+1, item.Similarity, item.Payload))
			}
		}

		// 2. 初始化大模型客户端
		loadEnv()
		apiKey := os.Getenv("DASHSCOPE_API_KEY")
		if apiKey == "" {
			apiKey = os.Getenv("QWEN_API_KEY")
		}
		if apiKey == "" {
			fmt.Println(ColorRed + "错误: 未在 .env 或环境变量中找到 QWEN_API_KEY 或 DASHSCOPE_API_KEY" + ColorReset)
			return
		}

		modelName := os.Getenv("QWEN_MODEL")
		if modelName == "" {
			modelName = "qwen-plus"
		}

		c, err := client.New(llm.Config{
			Provider: "dashscope",
			Model:    modelName,
			APIKey:   apiKey,
		})
		if err != nil {
			fmt.Printf(ColorRed+"[初始化大模型失败] %v\n"+ColorReset, err)
			return
		}

		// 3. 拼接带有 RAG 上下文的系统提问 Prompt
		prompt := fmt.Sprintf(`你是一个亲切、务实且极其专业的智能助理。
以下是系统为您匹配并检索出来的实时参考信息：
[参考信息]
%s[参考信息结束]

请结合上述参考信息，用自然、真实且贴近日常生活/真实工作场景的口吻，完整且详尽地回答用户的问题。
【绝对禁止】在回答中提及或使用任何关于“星空拓扑”、“高维物理场”、“力场”、“坍缩”、“能谱”、“坐标锚点”、“流动事件”等本项目的物理场专属名词或科幻/学术风格语调，请以通俗、专业的普通人语言进行沟通。

用户的问题是：%s`, contextSB.String(), query)

		fmt.Println(ColorCyan + "\n[大模型思考 🤖] 以下是来自物理场 RAG 增强的流式回答：\n" + ColorReset)
		fmt.Print(ColorGreen + "AI: " + ColorReset)

		_, err = c.SendStreamNoHistory(ctx, prompt, func(ctx context.Context, chunk string) error {
			fmt.Print(chunk)
			return nil
		})
		if err != nil {
			fmt.Printf(ColorRed+"\n[流式问答输出中断] Error: %v\n"+ColorReset, err)
		}
		fmt.Println("\n")

	default:
		fmt.Printf(ColorRed+"未知命令 %q\n"+ColorReset, command)
		printHelp()
	}
}
