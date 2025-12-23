package anything

import (
	"context"
	"sync"
)

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
// 100%    100%    100%     100%  预期总用时约等于(A+B+C+D)/2.7

const ExitFunction = "EXIT_FUNCTION"
const DC = "DISCHARGE_CARGO"
const NM = "NEW_MISSION"
const IM = "INTERRUPT_MISSION"
const RM = "RECOVERY_MISSION"

type Nothing interface {
	Register(...any)
	Schedule(string, []any) int64
	Init()
}

// Mission 即使是灰尘 也有他的使命
// Even the dust has his pursuit
type Mission struct {
	Name    string
	I       int64
	Pursuit []any
	A       []any
	C       byte
	// === 新增：隐形上下文 ===
	// 1. Context: 用于控制生命周期 (超时/取消)
	Context context.Context
	// 2. Store: 用于存储流程中的共享变量 (记忆)
	// 这样业务方法就不需要在入参里显式声明 map 了
	Store *sync.Map
	T     chan *Mission
}

// Everything 一切都源自于根
// Everything comes from the root
type Everything interface {
	Empty()
	map[int]string
}

type FOX interface {
	Init()
	DoMaps() chan struct{}
}

// 辅助方法：快速获取记忆
func (m *Mission) Get(key string) (any, bool) {
	if m.Store == nil {
		return nil, false
	}
	return m.Store.Load(key)
}

func (m *Mission) Set(key string, value any) {
	if m.Store == nil {
		m.Store = &sync.Map{}
	}
	m.Store.Store(key, value)
}

// DSLScript DSL 脚本结构
type DSLScript struct {
	Name  string         `json:"name"`
	Tags  string         `json:"tags"`  // 标签，如 "network,http,crawl"
	Vars  map[string]any `json:"vars"`  // 初始变量
	Steps []DSLStep      `json:"steps"` // 步骤
}

// DSLStep 单个步骤
type DSLStep struct {
	Call string `json:"call"` // 调用的原子方法名 (Native Method)
	Args []any  `json:"args"` // 参数，支持 "$var" 引用
	Save string `json:"save"` // 结果存入哪个变量
}
