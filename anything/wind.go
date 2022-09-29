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
	M     sync.Map
	R     sync.Map
	C     sync.Map
	A     sync.Map
	E     map[int64]chan struct{}
	f     FOX
	IWork *Worker
}

var wind *Wind
var allMission sync.Map

// SchedulePipeline  方法调度器
func SchedulePipeline(Name string, mis chan *Mission, inData []any) {
	if wind.f != nil {
		wind.f.DoMaps()
	}
	load, ok := allMission.Load(Name)
	if ok {
		load.(reflect.Value).Call([]reflect.Value{reflect.ValueOf(mis), reflect.ValueOf(inData)})
	}
}

// Schedule 方法调度器
func (w *Wind) Schedule(startName string, inData []any) int64 {
	if w.f != nil {
		w.f.DoMaps()
	}
	key := GetId()
	w.E[key] = make(chan struct{}, 10)
	var doFunc func(i int64, name string, data []any)
	w.C.Store(key, make(chan *Mission, 10))
	doFunc = func(I int64, name string, data []any) {
		defer func() {
			if err := recover(); err != nil {
				v, o := w.C.Load(I)
				if o {
					fmt.Println("Schedule Error!------ Exit Mission", "Error:", err, "MissionName:", v)
				}
				w.E[I] <- struct{}{}
			}
			w.C.Delete(I)
		}()
		load, ok := w.C.Load(I)
		if ok {
			load.(chan *Mission) <- &Mission{
				Name:    name,
				Pursuit: data,
			}
		}
		for {
			mis, _ := w.C.Load(I)
			mission := <-mis.(chan *Mission)
			switch mission.Name {
			case DC:
				w.A.Store(I, mission.Pursuit)
			case ExitFunction:
				w.A.Store(I, mission.Pursuit)
				w.E[I] <- struct{}{}
				return
			case NM:
				k := GetId()
				mission.T = make(chan *Mission, 2)
				//w.C[k] = mission.T
				w.C.Store(k, mission.T)
				go doFunc(k, mission.Pursuit[0].(string), mission.Pursuit[1:])
			case IM:
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
				go func() {
					if err := recover(); err != nil {
						v, o := w.C.Load(I)
						if o {
							fmt.Println("Schedule Error!------ Exit Mission", "Error:", err, "MissionName:", v)
						}
						w.E[I] <- struct{}{}
						delete(w.E, I)
					}
					lo, ok1 := w.M.Load(mission.Name)
					if ok1 {
						v, o := w.C.Load(I)
						if o {
							lo.(reflect.Value).Call([]reflect.Value{reflect.ValueOf(v), reflect.ValueOf(mission.Pursuit)})
						}
					}
				}()
			}

		}
	}
	go doFunc(key, startName, inData)
	return key
}

// Init 初始化Wind tags:"来无影去无踪"
func (w *Wind) Init() {
	wind = w
	node, err := NewWorker(1)
	if err != nil {
		fmt.Println(err)
		return
	}
	w.IWork = node
	w.M = sync.Map{}
	allMission = sync.Map{}
	w.C = sync.Map{}
	w.E = make(map[int64]chan struct{})
	w.A = sync.Map{}
	for i := range w.D {
		client := reflect.ValueOf(w.D[i])
		dus := client.Type()
		for j := 0; j < dus.NumMethod(); j++ {
			method := dus.Method(j)
			if method.Name != "" && method.Name != " " {
				if _, ok := w.M.Load(method.Name); ok && method.Name != "Empty" {
					log.Println("panic:", method.Name, "已存在 请检查", client)
				}
				w.M.Store(method.Name, client.MethodByName(method.Name))
				allMission.Store(method.Name, client.MethodByName(method.Name))
			}
		}
	}
}

func (w *Wind) RegisterRouters(values []any) {
	w.R = sync.Map{}
	for i := range values {
		client := reflect.ValueOf(values[i])
		dus := client.Type()
		for j := 0; j < dus.NumMethod(); j++ {
			method := dus.Method(j)
			if method.Name != "" && method.Name != " " {
				w.R.Store(method.Name, client.MethodByName(method.Name))
			}
		}
	}
}

// Register 注册方法 根据结构体
// 注：若为指针则会注册所有公开方法 非指针只会注册非指针传递方法
func (w *Wind) Register(a ...any) {
	w.D = append(w.D, a...)
}

func (w *Wind) SetController(f FOX) {
	if f != nil {
		w.f = f
	} else {
		panic("not implemented")
	}
}
