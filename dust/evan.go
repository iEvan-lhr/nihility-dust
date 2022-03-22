package dust

import (
	"fmt"
	"github.com/iEvan-lhr/nihility-dust/anything"
	"math/rand"
)

/*
如果这个世界上有什么东西是最小的，
	那他就不能使用任何东西
If there is the smallest thing in this world,
	he can’t use anything.
*/

func (d Dust) Empty() {
	//TODO implement me
}

//Dust 灰尘 即使是最小的分子 团结起来也能创造一切
type Dust struct {
	imgMap map[int]string
}

func (d *Dust) SliceString(mission chan *anything.Mission, data []any) {
	fmt.Println(data)
	mission <- &anything.Mission{Name: anything.ExitFunction}
}

func (d *Dust) CheckString(mission chan *anything.Mission, data []any) {
	fmt.Println(data)
	mission <- &anything.Mission{Name: anything.ExitFunction}
}

func (d *Dust) CheckIsBig(mission chan *anything.Mission, a []any) {
	x, y := rand.Intn(20), rand.Intn(20)
	mission <- &anything.Mission{Name: "CountXY", Pursuit: []any{x, y}}
}

func (d *Dust) AllNumber(mission chan *anything.Mission, a []any) {
	if a[0].(int) == 23 {
		mission <- &anything.Mission{Name: anything.ExitFunction, Pursuit: []any{23}}
	} else {
		mission <- &anything.Mission{Name: anything.DC, Pursuit: []any{a[0].(int)}}
		mission <- &anything.Mission{Name: "CheckIsBig", Pursuit: []any{}}
	}

}
