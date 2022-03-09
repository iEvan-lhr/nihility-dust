package wind

import (
	"github.com/iEvan-lhr/nihility-dust/anything"
	"reflect"
	"sync"
)

// Wind 实现自Home Nothing
type Wind struct {
	D []any
	M map[string]reflect.Value
	C chan *anything.Mission
	A sync.Map
	//R []reflect.Value
}

// Schedule 方法调度器
func (w *Wind) Schedule(startName string, inData ...any) {
	go func() {
		w.C <- &anything.Mission{
			Name:    startName,
			Pursuit: inData,
		}
		for {
			mission := <-w.C
			//log.Println(mission.Name)
			switch mission.Name {
			case anything.DC:
				w.A.Store(startName+anything.DC, mission)
			case anything.ExitFunction:
				w.A.Store(startName+anything.ExitFunction, mission)
				return
			default:
				go func() {
					defer func() {
						recover()
					}()
					w.M[mission.Name].Call([]reflect.Value{reflect.ValueOf(w.C), reflect.ValueOf(mission.Pursuit)})
				}()
			}
		}
	}()
}

// Init 初始化Wind tags:"来无影去无踪"
func (w *Wind) Init() {
	w.M = make(map[string]reflect.Value)
	w.C = make(chan *anything.Mission, 10)
	w.A = sync.Map{}
	for i := range w.D {
		client := reflect.ValueOf(w.D[i])
		dus := client.Type()
		for j := 0; j < dus.NumMethod(); j++ {
			method := dus.Method(j)
			if method.Name != "" && method.Name != " " {
				w.M[method.Name] = client.MethodByName(method.Name)
			}
		}
	}
}

func (w *Wind) RegisterRouters(values []reflect.Value) {
	for i := range values {
		dus := values[i].Type()
		for j := 0; j < dus.NumMethod(); j++ {
			method := dus.Method(j)
			if method.Name != "" && method.Name != " " {
				w.M[method.Name] = values[i].MethodByName(method.Name)
			}
		}
	}
}

// Register 注册方法 根据结构体
// 注：若为指针则会注册所有公开方法 非指针只会注册非指针传递方法
func (w *Wind) Register(a ...any) {
	w.D = append(w.D, a...)
}

//panic: interface conversion:
//	interface {} is func(*dust.Dust, chan *anything.Mission, []interface {}),
//				not func(chan *anything.Mission, []interface {})
