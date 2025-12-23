package anything

import (
	"fmt"
	"log"
	"reflect"
	"sync"
)

// Wind 实现自Home Nothing
type Wind struct {
	D []any
	// M wind 中的func集合
	M sync.Map
	// R wind 中的Router集合
	R sync.Map
	// C wind 中的Channel集合
	C sync.Map
	// A wind 中的返回值集合 在捕获到A中的返回值后 任务流程会默认认为已完成任务流
	A sync.Map
	// B (Brain) AI 大脑
	// 用于处理自然语言意图，将其转化为可执行的任务
	B Brain
	// E wind 中的外部判段Map 注：不推荐使用异步操作此Map 可能会出现操作异常
	E map[int64]chan struct{}
	// S (Spirit)
	// 核心执行逻辑的接口。如果不设置，默认为 DefaultSpirit。
	S Spirit
	// f wind 中的FOX控制器 基于系统阈值来调度任务,均衡调度系统使用,避免出现任务长等待情况,基准测试效率偏差不超过15%
	// 通过 SetController() 方法来设置控制器 需要实现FOX接口
	f FOX
	// IWork 基于雪花闪电算法的唯一键ID生成器 避免任务冲突 无需手动初始化
	// 如需初始化可通过 Wind.IWork=&struct 来设置 必须实现 NewWorker(workerId int64) (*Worker, error) 和 GetId() int64 方法
	IWork *Worker
}

// wind  指针类型的wind 外部不可操作 用于easyModel的操作
var wind *Wind

// allMission wind中的任务map的Copy用于执行异步中的短同步任务
var allMission sync.Map

// easyModel 非路由模式下的注册方法,使用非路由的模式进行执行 在需要高效率的开发过程中不推荐使用这种模式来开发 对代码耦合度较高的环境可以快速解耦
// 可通过不同包中的 init() 方法初始化需要解耦的方法 通过反射模式执行
var easyModel sync.Map

func init() {
	easyModel = sync.Map{}
}

// Init 初始化Wind 注册所有方法 tags:"来无影去无踪"
// 增加可变参数 spirits，允许在初始化时注入自定义的 Spirit (如星际跳跃逻辑)
func (w *Wind) Init(components ...any) {
	wind = w
	node, err := NewWorker(1)
	if err != nil {
		fmt.Println(err)
		return
	}
	w.IWork = node
	w.M = sync.Map{}
	allMission = sync.Map{}
	w.C = sync.Map{}
	w.E = make(map[int64]chan struct{})
	w.A = sync.Map{}

	// 组件注入逻辑
	for _, c := range components {
		switch v := c.(type) {
		case Spirit:
			w.S = v
		case Brain:
			w.B = v
		}
	}

	for i := range w.D {
		client := reflect.ValueOf(w.D[i])
		dus := client.Type()
		for j := 0; j < dus.NumMethod(); j++ {
			method := dus.Method(j)
			if method.Name != "" && method.Name != " " {
				if _, ok := w.M.Load(method.Name); ok && method.Name != "Empty" {
					log.Println("panic:", method.Name, "已存在 请检查", client)
				}
				w.M.Store(method.Name, client.MethodByName(method.Name))
				allMission.Store(method.Name, client.MethodByName(method.Name))
			}
		}
	}
}

// SetController 添加协程数量控制器
func SetController(f FOX) {
	if f != nil {
		if wind == nil {
			wind = &Wind{
				f: f,
			}
		} else {
			wind.f = f
		}
	} else {
		panic("not implemented")
	}
}
