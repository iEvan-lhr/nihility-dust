package anything

import (
	"context"
)

// Brain (大脑接口)
// 负责思考、推理和生成
type Brain interface {
	// Cognize (认知/思考)
	// ctx: 上下文
	// prompt: 用户的输入或当前的系统状态
	// schema: (可选) 期望输出的格式约束，比如要求返回特定的 JSON 结构
	Cognize(ctx context.Context, prompt string) (string, error)

	// ExtractKeywords 从用户意图中提取 3-5 个搜索关键词
	ExtractKeywords(ctx context.Context, prompt string) ([]string, error)
}
