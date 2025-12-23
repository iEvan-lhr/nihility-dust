package anything

import (
	"fmt"
	"log"
	"reflect"
)

// Schedule 核心方法调度器
func (w *Wind) Schedule(startName string, inData []any) int64 {
	// log.Println(startName)
	var do chan struct{}
	if w.f != nil {
		// 注册协程到协程数量控制器中
		do = w.f.DoMaps()
	}
	// 单任务流程唯一键ID
	key := GetId()

	w.E[key] = make(chan struct{}, 10)
	var doFunc func(i int64, name string, data []any, doChan chan struct{})
	// 根据KEY 初始化协程
	w.C.Store(key, make(chan *Mission, 10))

	// 方法执行器 执行多种状态
	doFunc = func(I int64, name string, data []any, doChan chan struct{}) {
		// defer 释放操作
		defer func() {
			if err := recover(); err != nil {
				v, o := w.C.Load(I)
				if o {
					fmt.Println("Schedule Error!------ Exit Mission", "Error:", err, "MissionName:", v)
				}
			}
			// 删除 初始化完成的协程
			w.C.Delete(I)
			if w.f != nil && doChan != nil {
				// 回收协程控制器
				doChan <- struct{}{}
			}
			// 给出返回值 外部可获取执行完成状态
			w.E[I] <- struct{}{}
		}()

		// 读取协程 I=KEY
		load, ok := w.C.Load(I)
		if ok {
			// 初始化首个任务
			load.(chan *Mission) <- &Mission{
				Name:    name,
				Pursuit: data,
			}
		}

		for {
			// 读取协程
			mis, on := w.C.Load(I)
			if !on {
				return
			}
			// 收取任务
			mission := <-mis.(chan *Mission)

			// 执行任务
			switch mission.Name {
			case DC:
				// 任务数据中转
				w.A.Store(I, mission.Pursuit)
			case ExitFunction:
				// 退出任务
				if mission.Pursuit != nil {
					w.A.Store(I, mission.Pursuit)
				}
				return
			case IM:
				// 中断逻辑 (保留原样)
				go func() {
					if err := recover(); err != nil {
						v, o := w.C.Load(I)
						if o {
							fmt.Println("Schedule Error!------ Exit Mission", "Error:", err, "MissionName:", v)
						}
						w.E[I] <- struct{}{}
					}
					lo, ok1 := w.M.Load(mission.Pursuit[0].(string))
					if ok1 {
						lo.(reflect.Value).Call([]reflect.Value{reflect.ValueOf(mission.T), reflect.ValueOf(mission.Pursuit[1:])})
					}
				}()
			case RM:
				log.Println("RM MissionName:")
			default:
				// 调用 Spirit 接口 (这是之前重构的核心点)
				w.S.Materialize(I, w, mission, mis.(chan *Mission))
			}
		}
	}

	go doFunc(key, startName, inData, do)
	return key
}
