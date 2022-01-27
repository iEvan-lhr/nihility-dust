package dust

/*
如果这个世界上有什么东西是最小的，
	那他就不能使用任何东西
If there is the smallest thing in this world,
	he can’t use anything.
*/

//Dust 灰尘 即使是最小的分子 团结起来也能创造一切
type Dust struct {
	//描述
	Describe string
	//任务
	Mission mission
	//counter
	Counter int
	//执行结果
	Answer any
}

//mission 即使是灰尘 也有他的使命
//Even the dust has his pursuit
type mission struct {
	name   string
	pursue chan any
}

func (d *Dust) findType() string {
	return ""
}

func (d *Dust) SliceString() {
	all := <-d.Mission.pursue

	if d.Mission.name != "SliceString" {
		d.Mission.pursue <- all
	} else {
		d.Answer = all
	}
}

func (d *Dust) SetMission(name string, pursue chan any) {
	d.Mission.pursue = pursue
	d.Mission.name = name
}
