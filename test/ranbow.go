package test

import (
	"github.com/iEvan-lhr/nihility-dust/anything"
)

type Ran struct {
}

func (r Ran) Empty() {
	//TODO implement me
}

func (r *Ran) CountXY(mission chan *anything.Mission, a []any) {
	//log.Println(a)
	x, y := a[0].(int), a[1].(int)
	mission <- &anything.Mission{Name: "AllNumber", Pursuit: []any{x + y}}
}

func (r *Ran) CountXYA(mission chan *anything.Mission, a []any) {
	//log.Println(a)
	x, y := a[0].(int), a[1].(int)
	mission <- &anything.Mission{Name: "AllNumber", Pursuit: []any{x + y}}
}
