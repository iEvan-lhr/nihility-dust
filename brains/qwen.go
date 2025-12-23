package brains

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

// QwenBrain 适配阿里云通义千问
type QwenBrain struct {
	ApiKey  string
	BaseURL string
	Model   string
	System  string // 默认的系统人设 (用于生成 DSL)
}

// NewQwenBrainFromEnv 从环境变量自动初始化
func NewQwenBrainFromEnv() *QwenBrain {
	return &QwenBrain{
		ApiKey:  os.Getenv("QWEN_API_KEY"),
		BaseURL: os.Getenv("QWEN_BASE_URL"),
		Model:   os.Getenv("QWEN_MODEL"),
		//System:  SystemPromptDSL, // 默认关联 DSL 生成规则
	}
}

// Cognize 实现 Brain 接口：生成 DSL
func (b *QwenBrain) Cognize(ctx context.Context, prompt string) (string, error) {
	// 复用底层请求逻辑，使用 Struct 中存储的默认 System Prompt (即架构师人设)
	return b.doChatRequest(ctx, b.System, prompt)
}

// ExtractKeywords 实现 Brain 接口：提取关键词
func (b *QwenBrain) ExtractKeywords(ctx context.Context, prompt string) ([]string, error) {
	// 1. 定义提取关键词专用的 System Prompt
	extractSysPrompt := `你是一个精准的搜索引擎助手。
任务：分析用户的输入，提取 1 到 5 个核心关键词（Tag）。
要求：
1. 关键词用于数据库模糊检索。
2. 请提取**单个词**，例如 "爬虫"、"百度"，不要提取长句子。
3. 只返回关键词，用英文逗号 "," 分隔。
4. 如果有英文，请转为小写

示例输入：帮我爬取百度首页
示例输出：crawl,baidu,network
`

	// 2. 调用底层请求
	// 2. 调用底层请求
	resultStr, err := b.doChatRequest(ctx, extractSysPrompt, prompt)
	if err != nil {
		return nil, err
	}

	// 3. 健壮的切分逻辑 (Fix: 同时支持逗号、空格、换行符切分)
	// 这样 "magic box result" 会被切分成 ["magic", "box", "result"]
	// 从而命中数据库里的 "magic" 标签
	f := func(c rune) bool {
		return c == ',' || c == ' ' || c == '\n' || c == '，'
	}

	rawParts := strings.FieldsFunc(resultStr, f)
	keywords := make([]string, 0, len(rawParts))

	for _, p := range rawParts {
		clean := strings.ToLower(strings.TrimSpace(p))
		clean = strings.Trim(clean, `"' .`)
		if clean != "" && len(clean) > 1 { // 过滤掉太短的杂质
			keywords = append(keywords, clean)
		}
	}

	return keywords, nil
}

// ==========================================
// 私有方法：封装通用的 HTTP 请求逻辑
// ==========================================

func (b *QwenBrain) doChatRequest(ctx context.Context, sysPrompt, userPrompt string) (string, error) {
	if b.ApiKey == "" {
		return "", fmt.Errorf("QWEN_API_KEY not set")
	}

	// 构造请求体
	reqBody := map[string]any{
		"model": b.Model,
		"messages": []map[string]string{
			{"role": "system", "content": sysPrompt},
			{"role": "user", "content": userPrompt},
		},
		"temperature": 0.1, // 保持低温度，确保结果稳定
		"top_p":       0.8,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	// 发起 HTTP 请求
	req, err := http.NewRequestWithContext(ctx, "POST", b.BaseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+b.ApiKey)

	client := &http.Client{Timeout: 30 * time.Second} // 设置超时
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// 错误处理
	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("Qwen API Error [%d]: %s", resp.StatusCode, string(body))
	}

	// 解析响应结构
	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("json decode error: %v", err)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("empty response from LLM")
	}

	return result.Choices[0].Message.Content, nil
}
