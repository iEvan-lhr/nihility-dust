package anything

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strings"
	"sync"
)

// DSLSpirit 核心灵体：支持原生代码 + 数据库 DSL 的混合调度
type DSLSpirit struct {
	DefaultSpirit             // 继承默认 Spirit (用于执行 Native 方法)
	Store         ScriptStore // 数据库存储接口
}

// NewDSLSpirit 初始化
func NewDSLSpirit(store ScriptStore) *DSLSpirit {
	return &DSLSpirit{Store: store}
}

// Materialize 显形：Wind 的核心入口
// 调度优先级：内存(Native/Cached) > 数据库(DSL Cold Load)
func (s *DSLSpirit) Materialize(id int64, w *Wind, mission *Mission, pipeline chan *Mission) {
	// A. 尝试内存方法
	// 1. 原生注册的代码 (手脚)
	// 2. 通过 LoadByTag 预加载的 DSL (缓存闭包)
	if _, ok := w.M.Load(mission.Name); ok {
		s.DefaultSpirit.Materialize(id, w, mission, pipeline)
		return
	}

	// B. 尝试数据库冷加载 (JIT Compile & Run)
	// 当内存中找不到时，去数据库读取 JSON 并解释执行
	go func() {
		defer func() {
			if err := recover(); err != nil {
				fmt.Printf("[DSL Panic] %v\n", err)
			}
			// 发送任务结束信号
			if ch, ok := w.E[id]; ok {
				select {
				case ch <- struct{}{}:
				default:
				}
			}
		}()

		// 1. 从 DB 读取代码
		content, err := s.Store.Get(mission.Name)
		if err != nil {
			log.Printf("[Spirit] 404: Mission '%s' not found in Memory or DB\n", mission.Name)
			return
		}

		// 2. 执行解释逻辑
		// 将 Mission 携带的数据 (Pursuit) 作为输入
		s.ExecuteDSL(w, mission.Name, content, pipeline, mission.Pursuit)
	}()
}

// LoadByTag 通过标签模糊搜索并批量注册到 Wind 内存中
// 返回成功加载的数量
func (s *DSLSpirit) LoadByTag(w *Wind, tag string) int {
	// 1. 搜索数据库
	scripts, err := s.Store.Search(tag)
	if err != nil {
		fmt.Printf("[DSLSpirit] Search Error: %v\n", err)
		return 0
	}

	count := 0
	for name, content := range scripts {
		// 变量捕获：确保闭包里用的是当前循环的变量
		scriptName := name
		scriptContent := content

		// 2. 注册闭包到 Wind.M (预热缓存)
		// 这样下次调度时，直接命中 Materialize 的 A 分支，无需查库
		w.M.Store(scriptName, func(pipeline chan *Mission, data []any) {
			// 在闭包内部复用解释器逻辑
			// 注意：这里是一个全新的执行上下文
			s.ExecuteDSL(w, scriptName, scriptContent, pipeline, data)
		})

		fmt.Printf("   -> [PreLoad] %s\n", scriptName)
		count++
	}
	return count
}

// ExecuteDSL 通用解释器核心逻辑
// 供 Materialize (冷启动) 和 LoadByTag (预加载闭包) 复用
func (s *DSLSpirit) ExecuteDSL(w *Wind, name, content string, pipeline chan *Mission, inputData []any) {
	// 1. 解析 JSON
	var script DSLScript
	if err := json.Unmarshal([]byte(content), &script); err != nil {
		log.Printf("[DSL Error] Syntax Error in '%s': %v\n", name, err)
		return
	}

	log.Printf(">>> DSL Start: %s", script.Name)

	// 2. 初始化上下文 (Stack Memory)
	ctx := &sync.Map{}

	// 2.1 加载 JSON 中定义的静态变量
	if script.Vars != nil {
		for k, v := range script.Vars {
			ctx.Store(k, v)
		}
	}
	// 2.2 加载输入数据 (如果有)
	if len(inputData) > 0 {
		ctx.Store("input", inputData)
	}

	// 3. 解释器循环 (Instruction Cycle)
	for i, step := range script.Steps {
		// 3.1 查找原子方法 (Atom) - 必须是已注册的 Native 方法或 Closure
		atom, ok := w.M.Load(step.Call)
		if !ok {
			log.Printf("[DSL Error] Atom '%s' not found at step %d", step.Call, i)
			return
		}

		// 3.2 解析参数 (Variable Resolution)
		realArgs := s.resolveArgs(step.Args, ctx)

		// 3.3 调用原子方法 (Invocation)
		results := s.invoke(atom, realArgs, pipeline)

		// 3.4 结果回写 (Write Back)
		if step.Save != "" && len(results) > 0 {
			// 默认只取第一个返回值存入上下文
			// 例如 http.Get 返回 (resp, error)，这里只存 resp
			val := results[0].Interface()
			ctx.Store(step.Save, val)
		}
	}

	log.Printf("<<< DSL End: %s", script.Name)
}

// resolveArgs 解析参数列表，处理 "$var" 和 "$var.Field" 引用
func (s *DSLSpirit) resolveArgs(args []any, ctx *sync.Map) []any {
	resolved := make([]any, len(args))
	for i, arg := range args {
		str, ok := arg.(string)
		if ok && strings.HasPrefix(str, "$") {
			// 去掉前缀 $
			fullKey := str[1:]
			parts := strings.Split(fullKey, ".")
			rootKey := parts[0]

			// 从 Context 获取根对象
			val, exists := ctx.Load(rootKey)
			if exists {
				if len(parts) > 1 {
					// 递归反射获取字段 (e.g. resp.Body)
					fieldVal := s.getField(val, parts[1:])
					resolved[i] = fieldVal
				} else {
					// 直接引用变量
					resolved[i] = val
				}
			} else {
				// 变量不存在，传 nil (避免 panic)
				resolved[i] = nil
			}
		} else {
			// 普通字面量参数
			resolved[i] = arg
		}
	}
	return resolved
}

// getField 递归反射获取结构体字段 (支持指针自动解引用)
func (s *DSLSpirit) getField(obj any, path []string) any {
	if obj == nil {
		return nil
	}

	v := reflect.ValueOf(obj)

	for _, fieldName := range path {
		// 1. 处理指针 (如 *http.Response)，直到获取到具体的值
		for v.Kind() == reflect.Ptr {
			if v.IsNil() {
				return nil // 空指针保护，防止 Panic
			}
			v = v.Elem()
		}

		// 2. 只有结构体才有字段
		if v.Kind() != reflect.Struct {
			return nil
		}

		// 3. 查找字段
		f := v.FieldByName(fieldName)
		if !f.IsValid() {
			return nil // 字段不存在
		}
		v = f
	}

	// 4. 返回最终值
	if v.IsValid() && v.CanInterface() {
		return v.Interface()
	}
	return nil
}

// invoke 反射调用任意函数
func (s *DSLSpirit) invoke(handler any, args []any, pipe chan *Mission) []reflect.Value {
	// ==========================================================
	// 1. 特殊处理：Wind 标准闭包 (Wind Closure)
	// ==========================================================
	// 如果 handler 是 func(chan *Mission, []any)，说明这是 Wind 内部注册的任务
	// 我们需要手动注入 pipeline，并将 args 作为 data 传入
	if closure, ok := handler.(func(chan *Mission, []any)); ok {
		closure(pipe, args)
		return nil // 闭包通常没有返回值
	}

	// ==========================================================
	// 2. 通用处理：标准库/普通函数 (Reflection)
	// ==========================================================
	v := reflect.ValueOf(handler)
	t := v.Type()

	in := make([]reflect.Value, 0, len(args))

	for i, arg := range args {
		// 处理变长参数的越界检查
		var targetType reflect.Type
		if t.IsVariadic() && i >= t.NumIn()-1 {
			targetType = t.In(t.NumIn() - 1).Elem()
		} else if i < t.NumIn() {
			targetType = t.In(i)
		} else {
			// 参数过多，忽略
			break
		}

		if arg == nil {
			// 如果参数是 nil，生成对应类型的零值
			in = append(in, reflect.Zero(targetType))
		} else {
			argVal := reflect.ValueOf(arg)
			// 尝试类型转换
			if argVal.Type().ConvertibleTo(targetType) {
				in = append(in, argVal.Convert(targetType))
			} else {
				// 硬传
				in = append(in, argVal)
			}
		}
	}

	// 安全检查：如果参数还是不够 (比如 args 为空，但函数需要参数)，补零值
	// 这主要针对 http.Get(url) 这种，如果 args 为空，这里需要补一个空字符串防止 panic
	for len(in) < t.NumIn() {
		if t.IsVariadic() && len(in) >= t.NumIn()-1 {
			break
		}
		targetType := t.In(len(in))
		in = append(in, reflect.Zero(targetType))
	}

	// 执行调用
	return v.Call(in)
}
