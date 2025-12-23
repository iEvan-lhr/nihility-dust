package anything

import (
	"context"
	"fmt"
	"strings"
)

// SystemPromptTemplate 动态提示词模板
// 使用 %s 占位符来动态插入工具列表
const SystemPromptTemplate = `
你是一个 Wind 系统的任务编排架构师。你的唯一任务是将用户的自然语言意图转换为 JSON 格式的任务脚本。

### 1. 当前可用工具 (Tools Available):
%s

### 2. 输出严格要求:
- 必须是纯 JSON 字符串。
- 严禁包含 Markdown 标记（如 '''json）。
- 不要包含任何解释性文字。

### 3. JSON 结构模板:
{
  "name": "GeneratedTask",
  "vars": {},
  "steps": [
    {"call": "方法名", "args": ["参数", "$引用变量"], "save": "结果变量"}
  ]
}
`

// Wish 智能许愿 (RAG 动态组装版)
func (w *Wind) Wish(userPrompt string) {
	ctx := context.Background()
	fmt.Printf(">>> User Wish: %s\n", userPrompt)

	// ================================================
	// 1. 意图分析与关键词提取
	// ================================================
	keywords, err := w.B.ExtractKeywords(ctx, userPrompt)
	if err != nil {
		fmt.Printf("[Wind] Keyword Error: %v\n", err)
		keywords = []string{} // 降级：无关键词
	}
	fmt.Printf("    -> Keywords: %v\n", keywords)

	// ================================================
	// 2. 动态检索工具 (RAG Retrieval)
	// ================================================
	// 2.1 获取标准库工具 (L1)
	stdToolsDesc := GetStdToolsPrompt()

	// 2.2 获取扩展工具 (L2 - 从数据库根据关键词捞)
	var customToolsDesc string
	if dslSpirit, ok := w.S.(*DSLSpirit); ok {
		tools, _ := dslSpirit.Store.SearchByKeywords(keywords)
		if len(tools) > 0 {
			var sb strings.Builder
			sb.WriteString("\n>>> 扩展能力 (Retrieved Skills):\n")
			for _, t := range tools {
				// 拼接格式： - 方法名: 签名 [描述]
				sb.WriteString(fmt.Sprintf("- %s: %s [%s]\n", t.Name, t.Signature, t.Description))

				// ⚡️ 重要：检索到的工具必须预加载到内存，否则 AI 生成了也无法执行
				// 这里做一个静默预加载
				content, _ := dslSpirit.Store.Get(t.Name)
				w.M.Store(t.Name, func(p chan *Mission, d []any) {
					dslSpirit.ExecuteDSL(w, t.Name, content, p, d)
				})
			}
			customToolsDesc = sb.String()
		}
	}

	// 2.3 组合所有工具描述
	allToolsDescription := stdToolsDesc + "\n" + customToolsDesc

	// ================================================
	// 3. 动态组装 Prompt (Prompt Engineering)
	// ================================================
	// 这里将工具列表注入到模板中
	finalSystemInstruction := fmt.Sprintf(SystemPromptTemplate, allToolsDescription)

	// 构造最终发给 LLM 的内容： 系统指令 + 用户需求
	fullPayload := fmt.Sprintf("%s\n\n### 用户请求:\n%s", finalSystemInstruction, userPrompt)

	// ================================================
	// 4. 调用大脑生成 (Generation)
	// ================================================
	fmt.Println(">>> Brain Generating Plan...")
	dslJson, err := w.B.Cognize(ctx, fullPayload)
	if err != nil {
		fmt.Printf("[Wind] Brain Error: %v\n", err)
		return
	}

	// 清洗 JSON
	dslJson = cleanJson(dslJson)
	fmt.Printf(">>> Generated DSL:\n%s\n", dslJson)

	// ================================================
	// 5. 执行 (Execution)
	// ================================================
	if dslSpirit, ok := w.S.(*DSLSpirit); ok {
		// 异步执行
		go dslSpirit.ExecuteDSL(w, "AI.GeneratedTask", dslJson, nil, nil)
	}
}

// 辅助方法：获取标准库描述
func GetStdToolsPrompt() string {
	Registry.RLock()
	defer Registry.RUnlock()

	var sb strings.Builder
	sb.WriteString(">>> 基础能力 (Standard Libs):\n")
	for _, t := range Registry.Tools {
		// 过滤掉那些没有签名的（非公开工具）
		if t.Signature != "" {
			sb.WriteString(fmt.Sprintf("- %s: %s [%s]\n", t.Name, t.Signature, t.Description))
		}
	}
	return sb.String()
}

func cleanJson(str string) string {
	str = strings.TrimSpace(str)
	str = strings.ReplaceAll(str, "```json", "")
	str = strings.ReplaceAll(str, "```", "")
	return str
}
