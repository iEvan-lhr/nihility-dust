package anything

import (
	"log"
	"reflect"
)

// OnceSchedule 单步任务调度器
func OnceSchedule(Name string, inData []any) {
	go func() {
		if wind != nil && wind.f != nil {
			do := wind.f.DoMaps()
			defer func() {
				do <- struct{}{}
			}()
		}
		if err := recover(); err != nil {
			log.Println(err)
		}
		load, ok := allMission.Load(Name)
		if ok {
			load.(reflect.Value).Call([]reflect.Value{reflect.ValueOf(inData)})
		} else {
			// 简单模式调度
			load, ok = easyModel.Load(Name)
			if ok {
				// 不支持有返回值的单次调度模式
				load.(reflect.Value).Call(GetReflectValues(inData))
			} else {
				panic("Func is not find:" + Name)
			}
		}
	}()
}

// DoSchedule 同步任务调度器
func DoSchedule(name string, inData ...any) (res []any) {
	if len(inData) > 1 && inData[1] == 1 {
		go func() {
			if wind != nil && wind.f != nil {
				do := wind.f.DoMaps()
				defer func() {
					do <- struct{}{}
				}()
			}
			res = exec(name, inData[0].([]any))
		}()
	} else {
		res = exec(name, inData[0].([]any))
	}
	return
}

// exec 内部执行辅助函数
func exec(name string, inData []any) (res []any) {
	if err := recover(); err != nil {
		log.Println(err)
	}
	var call []reflect.Value
	load, ok := allMission.Load(name)
	if ok {
		call = load.(reflect.Value).Call(GetReflectValues(inData))
	} else {
		// 简单模式调度
		load, ok = easyModel.Load(name)
		if ok {
			call = load.(reflect.Value).Call(GetReflectValues(inData))
		} else {
			panic("Func is not find:" + name)
		}
	}
	for _, value := range call {
		res = append(res, value.Interface())
	}
	return
}

// SchedulePipeline 方法调度器
func SchedulePipeline(Name string, mis chan *Mission, inData []any) {
	go func() {
		if wind != nil && wind.f != nil {
			do := wind.f.DoMaps()
			defer func() {
				do <- struct{}{}
			}()
		}
		load, ok := allMission.Load(Name)
		if ok {
			// 流程控制器调度
			load.(reflect.Value).Call([]reflect.Value{reflect.ValueOf(mis), reflect.ValueOf(inData)})
		} else {
			// 简单模式调度
			load, ok = easyModel.Load(Name)
			if ok {
				call := load.(reflect.Value).Call(append([]reflect.Value{reflect.ValueOf(mis)}, GetReflectValues(inData)...))
				var res []any
				for _, value := range call {
					res = append(res, value.Interface())
				}
				mis <- &Mission{Name: RM, Pursuit: res}
			} else {
				panic("Func is not find:" + Name)
			}
		}
	}()
}
