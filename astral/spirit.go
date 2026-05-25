package astral

import (
	"sync"
	"time"

	"github.com/iEvan-lhr/nihility-dust/anything"
)

// AstralSpirit 结合六维时空坍缩的灵魂路由器
type AstralSpirit struct {
	anything.DefaultSpirit                 // 继承默认之灵以处理原生/反射调用
	Engine                 *CollapseEngine // 延迟计算与相干坍缩引擎
}

// NewAstralSpirit 初始化 Astral 专属魂魄
func NewAstralSpirit(engine *CollapseEngine) *AstralSpirit {
	return &AstralSpirit{
		Engine: engine,
	}
}

// Materialize 显形：接收 Wind 传来的任务，拦截并执行量子相干坍缩观测，最后唤醒原子 Dust
func (s *AstralSpirit) Materialize(id int64, w *anything.Wind, mission *anything.Mission, pipeline chan *anything.Mission) {
	// 1. 判断该任务流中是否携带坐标锚点 ID
	var anchorID int64
	var hasAnchor bool

	if len(mission.Pursuit) > 0 {
		switch v := mission.Pursuit[0].(type) {
		case int64:
			anchorID = v
			hasAnchor = true
		case int:
			anchorID = int64(v)
			hasAnchor = true
		}
	}

	// 2. 如果携带有效坐标，进行量子观测与状态延迟坍缩
	if hasAnchor {
		// 瞬时观测：以当前毫秒时间戳发生坍缩
		nowMs := time.Now().UnixNano() / int64(time.Millisecond)
		collapsedState, err := s.Engine.Collapse(anchorID, nowMs)
		if err == nil {
			// 初始化记忆共享 Store，防止为空
			if mission.Store == nil {
				mission.Store = &sync.Map{}
			}
			// 将这一微秒的绝对实存状态注入隐形上下文中，供原子 Dust 执行体直接读取
			mission.Store.Store("astral_state", collapsedState)
		}
	}

	// 3. 完美向下委托，完成反射/原生闭包的业务调用
	s.DefaultSpirit.Materialize(id, w, mission, pipeline)
}
