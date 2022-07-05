package anything

import "log"

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

func DoChanN(mission chan *Mission, pursuit []any) chan *Mission {
	mis := Mission{Name: NM, Pursuit: pursuit, T: make(chan *Mission, 2)}
	mission <- &mis
	return mis.T
}
