package main

import (
	"fmt"
	"github.com/iEvan-lhr/nihility-dust/anything"
	"github.com/iEvan-lhr/nihility-dust/brains"
	"github.com/joho/godotenv"
	"os"
	"strings"
	"testing"
	"time"
)

// =================================================================
// 场景 2: 真实集成测试 (需要 .env 配置，调用真实 Qwen API)
// =================================================================

func TestWind_Wish_Real_Qwen(t *testing.T) {
	// 尝试加载 .env，如果没 Key 就跳过测试
	_ = godotenv.Load(".env") // 假设 .env 在上级目录，根据实际位置调整
	if os.Getenv("QWEN_API_KEY") == "" {
		t.Skip("Skipping real integration test: QWEN_API_KEY not found in env")
	}

	// 1. 初始化
	store, _ := anything.NewSQLiteStore(":memory:") // 依然用内存库，保持清洁
	dslSpirit := anything.NewDSLSpirit(store)
	qwenBrain := brains.NewQwenBrainFromEnv()

	w := &anything.Wind{}
	w.Init(dslSpirit, qwenBrain)

	// 2. 注册基础工具
	w.RegisterTool("fmt.Println", fmt.Println, "打印日志", "fmt.Println(msg...)")
	w.RegisterTool("ToString", func(b []byte) string { return string(b) }, "转字符串", "ToString(any)")

	// 3. 注册一个结果捕获工具
	resultChan := make(chan string, 1)
	w.RegisterTool("Test.Report", func(answer string) {
		resultChan <- answer
	}, "报告最终结果", "Test.Report(answer string)")

	// 4. 存入一个数学计算的 DSL 到数据库 (扩展技能)
	// 让 AI 去检索这个技能
	mathScript := `
	{
		"name": "Skill.MathDouble",
		"steps": [
			{"call": "Test.Report", "args": ["Result is 200"]}
		]
	}`
	// 这里的逻辑有点绕：通常 Skill 是用来被调用的。
	// 为了简单验证，我们存一个带有 "math" 标签的工具描述，看 AI 能不能搜到并调用它，
	// 或者 AI 可能会直接用 Test.Report。
	// 为了强迫 AI 使用 RAG，我们定义一个它"不具备"的能力。
	store.Save(
		"Skill.MagicBox",
		mathScript,
		"magic,secret",
		"一个神奇的盒子，能直接给出 200 的结果",
		"Skill.MagicBox()",
	)

	// 预加载这个技能的逻辑
	w.M.Store("Skill.MagicBox", func(p chan *anything.Mission, d []any) {
		resultChan <- "Magic 200"
	})

	// 5. 发起请求
	prompt := "我想使用神奇盒子(magic box)来获取结果"
	t.Logf(">>> Asking Qwen: %s", prompt)
	w.Wish(prompt)

	// 6. 等待真实 AI 的反应
	select {
	case res := <-resultChan:
		t.Logf(">>> AI Execution Result: %s", res)
		if res != "Magic 200" && !strings.Contains(res, "200") {
			t.Errorf("Unexpected result: %s", res)
		}
	case <-time.After(15 * time.Second): // 给大模型多一点时间
		t.Fatal("Real AI Test Timed out!")
	}
}
