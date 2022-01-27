package dust

import (
	"nihility-dust/anything"
	"reflect"
)

type Wind struct {
	D *Dust
	M map[string]any
	C chan any
	//任务
	Mission *anything.Mission
}

func (w *Wind) Schedule(startName string, inData any) any {
	w.Mission = &anything.Mission{
		Name:   startName,
		Pursue: make(chan any),
	}
	w.Mission.Pursue <- inData
	for w.Mission.Name != "Exit" {
		order := <-w.Mission.Pursue
		w.Mission = w.M[w.Mission.Name].(func(any) *anything.Mission)(order)
	}
	return <-w.Mission.Pursue
}

func (w *Wind) Init() {
	dust := reflect.ValueOf(w.D).Type()
	w.M = make(map[string]any)
	for i := 0; i < dust.NumMethod(); i++ {
		method := dust.Method(i)
		if method.Name != "" && method.Name != " " {
			w.M[method.Name] = method.Func
		}
	}
}
