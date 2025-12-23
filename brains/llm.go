package brains

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// LLMBrain 基于 HTTP 的通用 LLM 实现
type LLMBrain struct {
	ApiKey  string
	BaseURL string // e.g., "https://api.deepseek.com/v1"
	Model   string // e.g., "deepseek-chat"
	System  string // 系统提示词 (人设)
}

func NewLLMBrain(apiKey, baseURL, model string) *LLMBrain {
	return &LLMBrain{
		ApiKey:  apiKey,
		BaseURL: baseURL,
		Model:   model,
		System:  "You are a helpful AI assistant.", // 默认人设
	}
}

// SetSystem 设置系统提示词
func (b *LLMBrain) SetSystem(prompt string) {
	b.System = prompt
}

// Cognize 实现 Brain 接口
func (b *LLMBrain) Cognize(ctx context.Context, prompt string) (string, error) {
	// 构造 OpenAI 格式的请求体
	reqBody := map[string]any{
		"model": b.Model,
		"messages": []map[string]string{
			{"role": "system", "content": b.System},
			{"role": "user", "content": prompt},
		},
		"temperature": 0.1, // 低温度，保证指令生成的稳定性
		"stream":      false,
	}

	jsonData, _ := json.Marshal(reqBody)

	// 发起请求
	req, _ := http.NewRequestWithContext(ctx, "POST", b.BaseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+b.ApiKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("LLM API Error: %s", string(body))
	}

	// 解析响应
	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("empty response from LLM")
	}

	return result.Choices[0].Message.Content, nil
}
