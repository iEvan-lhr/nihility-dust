// Package anything
// 程序执行的道路或许以前我们需要按部就班，但也出现了很多问题。
// 像下面这样
// 任务A    任务B    任务C    任务D
// 100%    0%      0%       0%
// 100%    100%    0%       0%
// 100%    100%    100%     0%
// 100%    100%    100%     100%  预期总用时A+B+C+D
// 因为程序间有需求与依赖关系，或许我们可以使用人为的方法优化这些方法
// 但是在优化时会变相增加代码的复杂度和代码的冗余，并且对于程序员而言，思考了太多执行顺序相关的代码。
// 在Go1.18beta1的版本中，我首次使用channel与泛型实现了无序执行，即程序无需考虑是否有依赖
// 由调度器来决定是否调用方法，而程序只需要从channel中获取数据，执行操作就可以了
// 使用这种方法，可以实现代码的最小化，优化程序的执行效率，充分的利用多核的性能
// 并且对于Golang编程人员来说无需再考虑循环依赖的问题。
// 预期的执行效果如下
// 任务A    任务B    任务C    任务D
// 10%     0%      0%       0%
// 50%     30%     10%      0%
// 80%     60%     40%      20%
// 100%    80%     60%      40%
// 100%    100%    80%      60%
// 100%    100%    100%     80%
// 100%    100%    100%     100%  预期总用时(A+B+C+D)/2.7
package anything

type TInterface interface {
	run(params ...any)
	goNext() any
	getStation() string
}

// Subway 创造性的使用方法 必须实现调度器方法 Schedule()
// 配合Station使用
// 像下面这样
//for {
//	mission := <-s.mission
//	if v, ok := s.stations[mission.(TInterface).getStation()]; ok {
//		switch v.mode {
//		case 1:
//			go v.area(mission)
//		case 0:
//			go v.area(mission)
//		}
//	} else {
//		log.Println("Error router:", mission.(TInterface).getStation(), "does not exist ")
//	}
//}
//type Subway interface {
//	chan any
//	map[string]Station
//	Schedule()
//	InitChanAny()
//	Register(f func(tInterface ...any), any ...any)
//	InSubway()
//}

// Station 站台 使用站台来区分多任务的模式，以期望达到最优化的效果
// 站台只需要注册一次 并且支持在程序执行中进行注册和删除。
//type Station interface {
//	//当前站台名称
//	string
//	//当前站台方法
//	func(...any)
//	int
//}

//Mission 即使是灰尘 也有他的使命
//Even the dust has his pursuit
type Mission struct {
	Name   string
	Pursue chan any
}
