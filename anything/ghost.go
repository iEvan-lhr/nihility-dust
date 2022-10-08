package anything

import (
	"log"
	"reflect"
)

func ErrorExit(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func ErrorDontExit(err error) {
	if err != nil {
		log.Println(err.Error())
	}
}

func DoChanTemp(mission chan *Mission, pursuit []any) chan *Mission {
	mis := Mission{Name: IM, Pursuit: pursuit, T: make(chan *Mission, 2)}
	mission <- &mis
	return mis.T
}

func DoChanN(Name string, pursuit []any) chan *Mission {
	mis := Mission{Name: Name, Pursuit: pursuit, T: make(chan *Mission, 2)}
	SchedulePipeline(Name, mis.T, pursuit)
	return mis.T
}

func GetReflectValues(data []any) []reflect.Value {
	var rf []reflect.Value
	for _, datum := range data {
		rf = append(rf, reflect.ValueOf(datum))
	}
	return rf
}
