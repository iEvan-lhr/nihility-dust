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

func DoChanTemp(starName string, pursuit []any, model int) (chan struct{}, chan *Mission) {
	mis := make(chan *Mission)
	schedule := wind.Schedule(starName, pursuit)
	if model == 0 {
		return wind.E[schedule], nil
	} else {
		do := wind.f.DoMaps()
		go func() {
			<-wind.E[schedule]
			mission, _ := wind.A.Load(schedule)
			mis <- &Mission{Pursuit: mission.([]any)}
			delete(wind.E, schedule)
			do <- struct{}{}
		}()
		return nil, mis
	}
}

func DoOnceMission(starName string, pursuit []any) {
	OnceSchedule(starName, pursuit)
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
