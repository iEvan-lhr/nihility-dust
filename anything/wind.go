package anything

import (
	"fmt"

	"log"
	"reflect"
	"sync"
)

// Wind 实现自Home Nothing
type Wind struct {
	D     []any
	M     map[string]reflect.Value
	R     map[string]reflect.Value
	C     map[int64]chan *Mission
	A     sync.Map
	E     map[int64]chan struct{}
	IWork *Worker
}

var allMission map[string]reflect.Value

//var chanM map[int64]chan *Mission

// Schedule 方法调度器
func Schedule(Name string, mis chan *Mission, inData []any) {
	//go func(k int64, m chan *Mission) {
	//	chanM[k] = m
	//	<-m
	//}(key, mis)
	allMission[Name].Call([]reflect.Value{reflect.ValueOf(mis), reflect.ValueOf(inData)})
}

//// Schedule 方法调度器
//func (w *Wind) Schedule(startName string, inData ...any) int64 {
//	key := w.IWork.GetId()
//	w.E[key] = make(chan struct{}, 10)
//	var doFunc func(i int64, name string, data ...any)
//	w.C[key] = make(chan *anything.Mission, 10)
//	doFunc = func(I int64, name string, data ...any) {
//		defer func() {
//			if err := recover(); err != nil {
//				fmt.Println("Schedule Error!------ Exit Mission", "Error:", err, "MissionName:", w.C[key])
//				w.E[key] <- struct{}{}
//			}
//			delete(w.C, I)
//		}()
//		w.C[I] <- &anything.Mission{
//			Name:    name,
//			Pursuit: data,
//		}
//		for {
//			mission := <-w.C[I]
//			switch mission.Name {
//			//case anything.DC:
//			//	if val, ok := w.A.Load(I); ok {
//			//		w.A.Store(I, anything.SetValReturn(mission, val.([]any)))
//			//	} else {
//			//		w.A.Store(I, mission.Pursuit)
//			//	}
//			case anything.ExitFunction:
//				//if val, ok := w.A.Load(I); ok {
//				//	w.A.Store(I, anything.SetValReturn(mission, val.([]any)))
//				//} else {
//				//	w.A.Store(I, mission.Pursuit)
//				//}
//				w.E[I] <- struct{}{}
//				return
//			//case anything.NM:
//			//	k := w.IWork.GetId()
//			//	w.C[k] = mission.T
//			//	doFunc(k, mission.Pursuit[0].(string), mission.Pursuit[1:])
//			//case anything.IM:
//			//	go func() {
//			//		if err := recover(); err != nil {
//			//			fmt.Println("Schedule Error!------ Exit Mission", "Error:", err, "MissionName:", w.C[I])
//			//			w.E[I] <- struct{}{}
//			//		}
//			//		w.M[mission.Pursuit[0].(string)].Call([]reflect.Value{reflect.ValueOf(mission.T), reflect.ValueOf(mission.Pursuit[1:])})
//			//	}()
//			//case anything.RM:
//			//	log.Println("RM MissionName:")
//			default:
//				go func() {
//					if err := recover(); err != nil {
//						fmt.Println("Schedule Error!------ Exit Mission", "Error:", err, "MissionName:", w.C[I])
//						w.E[I] <- struct{}{}
//					}
//					w.M[mission.Name].Call([]reflect.Value{reflect.ValueOf(w.C[I]), reflect.ValueOf(mission.Pursuit)})
//				}()
//			}
//		}
//	}
//	go doFunc(key, startName, inData)
//	return key
//}

// Init 初始化Wind tags:"来无影去无踪"
func (w *Wind) Init() {
	node, err := NewWorker(1)
	if err != nil {
		fmt.Println(err)
		return
	}
	w.IWork = node
	w.M = make(map[string]reflect.Value)
	w.C = make(map[int64]chan *Mission)
	w.E = make(map[int64]chan struct{})
	w.A = sync.Map{}
	for i := range w.D {
		client := reflect.ValueOf(w.D[i])
		dus := client.Type()
		for j := 0; j < dus.NumMethod(); j++ {
			method := dus.Method(j)
			if method.Name != "" && method.Name != " " {
				if _, ok := w.M[method.Name]; ok && method.Name != "Empty" {
					log.Println("panic:", method.Name, "已存在 请检查", client)
				}
				w.M[method.Name] = client.MethodByName(method.Name)
			}
		}
	}
	allMission = w.M
}

func (w *Wind) RegisterRouters(values []any) {
	w.R = make(map[string]reflect.Value)
	for i := range values {
		client := reflect.ValueOf(values[i])
		dus := client.Type()
		for j := 0; j < dus.NumMethod(); j++ {
			method := dus.Method(j)
			if method.Name != "" && method.Name != " " {
				w.R[method.Name] = client.MethodByName(method.Name)
			}
		}
	}
}

// Register 注册方法 根据结构体
// 注：若为指针则会注册所有公开方法 非指针只会注册非指针传递方法
func (w *Wind) Register(a ...any) {
	w.D = append(w.D, a...)
}
