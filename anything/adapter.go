package anything

import (
	"context"
	"fmt"
	"reflect"
	"sync"
)

// Adapt 将任意强类型函数转换为 Wind 标准执行体
// fn: 用户的业务函数，例如 func(req UserReq) (UserResp, error)
// 返回值: Wind 标准闭包 func(chan *Mission, []any)
func (w *Wind) Adapt(fn any) func(chan *Mission, []any) {
	fnVal := reflect.ValueOf(fn)
	fnType := fnVal.Type()

	if fnType.Kind() != reflect.Func {
		panic(fmt.Sprintf("Adapt: expected a function, got %T", fn))
	}

	return func(pipeline chan *Mission, data []any) {
		// --- 1. 准备入参 (Input Binding) ---
		inArgs := make([]reflect.Value, 0, fnType.NumIn())
		interpreter := NewInterpreter(data, pipeline)

		// 追踪使用了多少个 data 里的元素，用于 Direct Mapping
		dataIndex := 0

		for i := 0; i < fnType.NumIn(); i++ {
			argType := fnType.In(i)

			// Case A: 自动注入管道
			if argType == reflect.TypeOf((chan *Mission)(nil)) {
				inArgs = append(inArgs, reflect.ValueOf(pipeline))
				continue
			}

			// Case B: 结构体注入 (DTO 模式 - 推荐)
			// 如果参数是结构体（非指针），我们创建一个指针去绑定，然后取值传进去
			if argType.Kind() == reflect.Struct {
				argPtr := reflect.New(argType) // 创建 *Struct
				if err := interpreter.Bind(argPtr.Interface()); err != nil {
					fmt.Printf("[Adapt] Bind error: %v\n", err)
				}
				inArgs = append(inArgs, argPtr.Elem()) // 传入 Struct 值
				continue
			}

			// Case C: 指针结构体注入
			if argType.Kind() == reflect.Ptr && argType.Elem().Kind() == reflect.Struct {
				argPtr := reflect.New(argType.Elem())
				if err := interpreter.Bind(argPtr.Interface()); err != nil {
					fmt.Printf("[Adapt] Bind error: %v\n", err)
				}
				inArgs = append(inArgs, argPtr) // 传入 *Struct
				continue
			}

			// Case D: 直接参数映射 (按顺序取 []any)
			// 比如函数 func(id int, name string)
			if dataIndex < len(data) {
				val := reflect.ValueOf(data[dataIndex])
				// 类型转换检查
				if val.Type().ConvertibleTo(argType) {
					inArgs = append(inArgs, val.Convert(argType))
				} else {
					// 类型不匹配，传零值以防 Panic
					inArgs = append(inArgs, reflect.Zero(argType))
					fmt.Printf("[Adapt] Type mismatch at index %d: need %v, got %v\n", dataIndex, argType, val.Type())
				}
				dataIndex++
			} else {
				// 数据不够了，传零值
				inArgs = append(inArgs, reflect.Zero(argType))
			}
		}

		// --- 2. 执行逻辑 (Execution) ---
		results := fnVal.Call(inArgs)

		// --- 3. 结果处理 (Output Assembly) ---
		// 默认情况下，Wind 的任务是不返回值的。
		// 但如果我们在做“代码组合”或“热加载流程”，我们可能需要这些返回值传递给下一个步骤。
		if len(results) > 0 {
			assembler := NewAssembler()
			outputData := assembler.Pack(results)

			// 【策略选择】
			// 这里有两个选择：
			// 1. 什么都不做，假设业务函数内部已经操作了 pipeline。
			// 2. (更智能的) 如果业务函数有返回值，我们默认认为这是要传给"下一个节点"的数据。
			//    但问题是我们不知道"下一个节点"是谁 (Mission Name)。

			// 为了通用性，我们目前仅打印或保留扩展点。
			// 如果配合"BluePrint" (蓝图) 模式，这里会将 outputData 返回给蓝图执行器。

			// 简单的调试输出，证明封装成功
			fmt.Printf("[Adapt] Function returned %d values: %v\n", len(outputData), outputData)
		}
	}
}

// SmartAdapt 智能适配器：将任意 Go 函数转换为 Wind 执行体
func (w *Wind) SmartAdapt(fn any) func(chan *Mission, []any) {
	fnVal := reflect.ValueOf(fn)
	fnType := fnVal.Type()

	return func(pipeline chan *Mission, data []any) {
		// 1. 获取当前任务的上下文 (Hack: 我们需要从 pipeline 的上游获取 mission 里的 store)
		// 但这里的闭包签名限制了我们拿不到 mission 对象，只拿到了 pipeline 和 data。
		// 为了解决这个问题，我们需要约定：Wind 在调用这个闭包前，
		// 如果需要 Context，我们得通过某种方式传递。
		//
		// **修正方案**：为了完美支持您的需求，我们修改 Spirit 调用逻辑，
		// 让 data 的第一个元素如果是 *Mission，则认为是上下文传递 (或者修改闭包签名)。
		//
		// 但为了不改动 func(chan, []any) 这个核心签名，
		// 我们假设 Spirit 在调用时，已经把 *Mission 放进了 data 的隐藏位置，或者我们只处理纯函数逻辑。

		// 这里的实现专注于：将 data ([]any) 映射到 fn 的入参

		inArgs := make([]reflect.Value, 0, fnType.NumIn())

		// 数据游标
		dataIdx := 0

		for i := 0; i < fnType.NumIn(); i++ {
			argType := fnType.In(i)

			// --- A. 智能注入基础设施 ---

			// 1. 注入 Pipeline
			if argType == reflect.TypeOf((chan *Mission)(nil)) {
				inArgs = append(inArgs, reflect.ValueOf(pipeline))
				continue
			}

			// 2. 注入 Context (标准库)
			if argType.Implements(reflect.TypeOf((*context.Context)(nil)).Elem()) {
				// 暂时传 Background，除非我们能在 data 里找到 Context
				inArgs = append(inArgs, reflect.ValueOf(context.Background()))
				continue
			}

			// 3. 注入 Memory (sync.Map)
			if argType == reflect.TypeOf((*sync.Map)(nil)) {
				// 需要从 data 里找，或者新建
				inArgs = append(inArgs, reflect.ValueOf(&sync.Map{}))
				continue
			}

			// --- B. 处理变长参数 (fmt.Println) ---
			if fnType.IsVariadic() && i == fnType.NumIn()-1 {
				// 将剩余的所有 data 放入
				for j := dataIdx; j < len(data); j++ {
					inArgs = append(inArgs, reflect.ValueOf(data[j]))
				}
				break
			}

			// --- C. 智能参数映射 (http.Get, time.Sleep) ---
			if dataIdx < len(data) {
				rawVal := data[dataIdx]
				targetVal := reflect.ValueOf(rawVal)

				// 类型转换黑魔法：尽可能让它通
				// 比如 data 是 int(1000)，目标是 time.Duration(int64)
				if rawVal != nil && targetVal.Type().ConvertibleTo(argType) {
					inArgs = append(inArgs, targetVal.Convert(argType))
				} else if rawVal == nil {
					// 允许传 nil
					inArgs = append(inArgs, reflect.Zero(argType))
				} else {
					// 类型硬不匹配，尝试强制兼容 (如 json number float64 -> int)
					// 这里简单处理：直接传，panic 了就让 Spirit 捕获
					inArgs = append(inArgs, targetVal)
				}
				dataIdx++
			} else {
				// 数据不够，补零值
				inArgs = append(inArgs, reflect.Zero(argType))
			}
		}

		// --- 执行 ---
		results := fnVal.Call(inArgs)

		// --- 结果处理 ---
		// 标准库通常返回 (res, error) 或 (int, error)
		// 我们需要把 error 剥离，或者把 res 传递给下一个节点
		var output []any
		for _, r := range results {
			// 如果是 Error 类型，并且不为 nil，可以记录日志或中断
			if r.Type().Implements(reflect.TypeOf((*error)(nil)).Elem()) {
				if !r.IsNil() {
					fmt.Printf("[SmartAdapt] System Error: %v\n", r.Interface())
					// 可以选择 pipe <- ExitFunction
				}
				continue // Error 不向下传递
			}
			output = append(output, r.Interface())
		}

		// 如果有输出，怎么传给下一个？
		// 在 Wind 模式下，除非显式发送 Mission，否则不会自动流转。
		// 但如果您使用 "Flow/JSON" 模式，那个 Flow Director 会捕获这里的 output。
	}
}
