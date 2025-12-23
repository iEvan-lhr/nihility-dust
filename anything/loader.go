package anything

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
)

// === 数据结构定义 ===

// FlowBlueprint 对应 JSON 结构
type FlowBlueprint struct {
	Name  string     `json:"name"`  // 新方法的注册名称
	Steps []FlowStep `json:"steps"` // 步骤列表
}

type FlowStep struct {
	Action   string            `json:"action"`    // 要调用的原子方法名 (e.g. "Inventory.Check")
	InputMap map[string]string `json:"input_map"` // 参数映射规则
}

// === 动态加载器 ===

// RegisterByJSON 通过 JSON 字符串注册新方法
func (w *Wind) RegisterByJSON(jsonStr string) error {
	// 1. 解析 JSON
	var bp FlowBlueprint
	if err := json.Unmarshal([]byte(jsonStr), &bp); err != nil {
		return fmt.Errorf("invalid json: %v", err)
	}

	// 2. 构建闭包 (The Director)
	// 这个闭包就是运行时实际执行的逻辑
	dynamicFunc := func(pipeline chan *Mission, data []any) {
		log.Printf("[Flow] Start executing flow: %s", bp.Name)

		// 创建上下文 (Context)，用于在步骤间传递数据
		// 初始数据来自入参 data
		ctx := make(map[string]any)
		for i, v := range data {
			ctx[fmt.Sprintf("data.%d", i)] = v
		}

		// 3. 遍历执行步骤
		for i, step := range bp.Steps {
			// A. 查找原子方法 (Atom)
			// 原子方法必须已经存在于 Wind.M 中
			atom, ok := w.M.Load(step.Action)
			if !ok {
				log.Printf("[Flow] Error: Step %d action '%s' not found", i, step.Action)
				return // 或者 panic
			}

			// B. 准备原子方法的参数
			// 这一步比较 tricky，因为原子方法签名可能是 func(chan, []any) 或 func(chan, Struct)
			// 为了通用，我们假设原子方法已经通过 Adapt 适配成了 func(chan, []any)
			// 或者我们需要在此处构造 []any 传给它

			stepInput := make([]any, 0)

			// 这里做一个简化假设：
			// 如果 InputMap 为空，直接透传原始 data
			if len(step.InputMap) == 0 {
				stepInput = data
			} else {
				// 如果有映射，我们需要根据规则构造参数
				// 注意：这里需要知道目标函数需要多少参数，由于 []any 是变长的，
				// 我们这里简化为：将 map 的值解析出来放入切片 (实际生产需要更有序的映射)
				// 更好的做法是：原子方法接受一个 Context Map，而不是 []any

				// 为了演示，我们把 context 注入进去
				stepInput = append(stepInput, ctx)
			}

			// C. 执行原子方法
			// 我们需要兼容 reflect.Value 和 原生 func
			switch v := atom.(type) {
			case func(chan *Mission, []any):
				v(pipeline, stepInput)
			case reflect.Value:
				v.Call([]reflect.Value{reflect.ValueOf(pipeline), reflect.ValueOf(stepInput)})
			default:
				log.Printf("[Flow] Unknown handler type for %s", step.Action)
			}

			// D. (可选) 获取结果并更新 Context
			// 如果需要拿到上一步的结果给下一步用，需要原子方法配合写入 context
			log.Printf("[Flow] Step %d [%s] completed", i, step.Action)
		}

		log.Printf("[Flow] Flow %s finished", bp.Name)
	}

	// 4. 注册到 Wind
	w.M.Store(bp.Name, dynamicFunc)
	log.Printf("✅ Dynamic Method Registered: %s", bp.Name)

	return nil
}
