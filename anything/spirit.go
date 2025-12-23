package anything

import (
	"fmt"
	"log"
	"reflect"
)

// Spirit (灵)
// 它是 Wind 的灵魂，决定了任务如何被路由、执行和跳转。
type Spirit interface {
	// Materialize 显形：接收 Wind 传来的任务，执行并决定下一步
	Materialize(id int64, w *Wind, mission *Mission, pipeline chan *Mission)
}

// DefaultSpirit 默认之灵
// 包含了兼容模式：支持反射调用(旧) 和 原生闭包调用(新)
type DefaultSpirit struct{}

func (d *DefaultSpirit) Materialize(id int64, w *Wind, mission *Mission, pipeline chan *Mission) {
	// 保持原有的异步执行特性
	go func() {
		var newDo chan struct{}
		// 1. 协程控制器逻辑
		if w.f != nil {
			newDo = w.f.DoMaps()
		}

		// 2. 错误恢复与退出逻辑
		defer func() {
			if err := recover(); err != nil {
				// 获取当前管道名称用于日志
				taskName := "Unknown"
				if mission != nil {
					taskName = mission.Name
				}
				fmt.Println("Schedule Error!------ Exit Mission", "Error:", err, "MissionName:", taskName)

				// 尝试发送退出信号 (保留原有逻辑)
				if ch, ok := w.E[id]; ok {
					// 非阻塞发送，防止死锁
					select {
					case ch <- struct{}{}:
					default:
					}
				}
				// 只有在 panic 时才删除 E，正常流程由外层 Schedule 处理
				// (注：这里保留您原有代码的逻辑风格)
			}

			// 任务结束后的清理逻辑 (Defer)
			if mission != nil && len(mission.Name) > 0 && mission.Name[0] == '-' {
				// 如果是自动结束任务
				// 注意：这里需要重新获取 channel，防止闭包引用的 pipeline 已经失效?
				// 实际上 pipeline 是局部变量，没问题。

				// 检查 channel 是否还开启 (虽然 pipeline 只要没被 close 就可以写)
				// 但为了安全，我们尝试捕获 panic (虽然外层已有 recover)
				defer func() { recover() }()
				pipeline <- &Mission{Name: ExitFunction, Pursuit: nil}
			}

			if newDo != nil {
				newDo <- struct{}{}
			}
		}()

		// 3. 核心执行逻辑
		// 从 Wind 的 M 中加载方法
		lo, ok1 := w.M.Load(mission.Name)
		if ok1 {
			// 二次检查管道是否存在 (防止任务流已被取消)
			_, o := w.C.Load(id)
			if o {
				// === 核心修复点 ===
				// 判断 lo 的类型，分别处理
				switch v := lo.(type) {

				// Case A: 原生函数 (来自 Adapt 或 闭包适配器)
				// 这种方式性能最高，直接调用
				case func(chan *Mission, []any):
					v(pipeline, mission.Pursuit)

				// Case B: 反射值 (来自旧的 Register)
				// 保持兼容性
				case reflect.Value:
					v.Call([]reflect.Value{reflect.ValueOf(pipeline), reflect.ValueOf(mission.Pursuit)})

				default:
					log.Printf("[Spirit] Error: Unknown handler type for '%s': %T\n", mission.Name, lo)
				}
			}
		} else {
			// 404: 任务未找到
			log.Printf("[Spirit] Warning: Mission '%s' not found.\n", mission.Name)
		}
	}()
}
