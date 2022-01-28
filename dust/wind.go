package dust

import (
	"github.com/iEvan-lhr/nihility-dust/anything"
	"reflect"
)

// Wind 实现自Home Nothing
type Wind struct {
	D []any
	M map[string]func(*Dust, chan *anything.Mission, []any)
	C chan *anything.Mission
	A map[string]*anything.Mission
}

// Schedule 方法调度器
func (w *Wind) Schedule(startName string, inData ...any) {
	go func() {
		w.C <- &anything.Mission{
			Name:    startName,
			Pursuit: inData,
		}
		i := 0
		for {
			mission := <-w.C
			if mission.Name != anything.ExitFunction {
				i++
				go w.M[mission.Name](&Dust{}, w.C, mission.Pursuit)
			} else {
				i--
				if i == 0 {
					w.A[startName] = mission
					return
				}
			}
		}
	}()
}

// Init 初始化Wind tags:"来无影去无踪"
func (w *Wind) Init() {
	w.M = make(map[string]func(*Dust, chan *anything.Mission, []any))
	w.C = make(chan *anything.Mission, 10)
	w.A = make(map[string]*anything.Mission)
	for i := range w.D {
		dus := reflect.ValueOf(w.D[i]).Type()
		for j := 0; j < dus.NumMethod(); j++ {
			method := dus.Method(j)
			if method.Name != "" && method.Name != " " {
				w.M[method.Name] = method.Func.Interface().(func(*Dust, chan *anything.Mission, []any))
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
