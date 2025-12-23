package anything

import (
	"sync"
)

// ToolMeta 工具的元数据 (说明书)
type ToolMeta struct {
	Name        string
	Description string // 功能描述，如 "发送HTTP GET请求"
	Signature   string // 调用签名，如 "http.Get(url string)"
}

// Registry 全局工具注册表
// 用于存储标准库 (StdLib) 和 内存中已加载工具的说明书
var Registry = struct {
	sync.RWMutex
	Tools map[string]ToolMeta
}{
	Tools: make(map[string]ToolMeta),
}

// RegisterTool 注册工具及其说明书
func (w *Wind) RegisterTool(name string, handler any, desc, sig string) {
	// 1. 注册执行逻辑 (手脚)
	w.M.Store(name, handler)

	// 2. 注册说明书 (给 AI 看的)
	Registry.Lock()
	Registry.Tools[name] = ToolMeta{
		Name:        name,
		Description: desc,
		Signature:   sig,
	}
	Registry.Unlock()
}
