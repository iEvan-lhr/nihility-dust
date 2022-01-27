package dust

import "nihility-dust/anything"

/*
如果这个世界上有什么东西是最小的，
	那他就不能使用任何东西
If there is the smallest thing in this world,
	he can’t use anything.
*/

//Dust 灰尘 即使是最小的分子 团结起来也能创造一切
type Dust struct {
	//执行结果
	Answer chan any
}

func (d *Dust) findType() string {
	return ""
}

func (d *Dust) SliceString(data any) *anything.Mission {
	return nil
}
