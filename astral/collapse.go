package astral

import (
	"math"
	"sync"
)

// CollapseEngine 无锁半衰期延迟衰减与增量相干坍缩引擎
type CollapseEngine struct {
	Store AstralStore
	mu    sync.RWMutex
	// BaseDecayRate 节点自身基态状态的背景半衰衰减系数 (默认 0.02)
	BaseDecayRate float64
	// EvolutionRate 自演化学习率 (默认 0.01, 即每次流动冲刷将 1% 的流动语义融进节点基态)
	EvolutionRate float64
}

// NewCollapseEngine 初始化坍缩引擎
func NewCollapseEngine(store AstralStore) *CollapseEngine {
	return &CollapseEngine{
		Store:         store,
		BaseDecayRate: 0.02,
		EvolutionRate: 0.01,
	}
}

// Collapse 在特定时间戳 t (毫秒) 瞬间，对 anchorID 节点执行观测坍缩，求出此时绝对的六维状态
func (e *CollapseEngine) Collapse(anchorID int64, t int64) (Vector6D, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// 1. 获取基态节点坐标与最近快照
	anchor, err := e.Store.GetAnchor(anchorID)
	if err != nil {
		return Vector6D{}, err
	}

	// 2. 计算节点最近快照本身，由于时间推移产生的自衰减
	dt := float64(t-anchor.LastCollapseTime) / 1000.0 // 换算为秒级
	if dt < 0 {
		dt = 0
	}
	baseDecayFactor := math.Exp(-e.BaseDecayRate * dt)

	// 计算背景自衰减后的状态
	collapsedState := Vector6D{
		Time:      anchor.LastState.Time * baseDecayFactor,
		Space:     anchor.LastState.Space * baseDecayFactor,
		PosNeg:    anchor.LastState.PosNeg * baseDecayFactor,
		Influence: anchor.LastState.Influence * baseDecayFactor,
		Danger:    anchor.LastState.Danger * baseDecayFactor,
		Base:      anchor.LastState.Base * baseDecayFactor,
	}

	// 3. 增量拉取自上次坍缩时间点至今，所有经过该节点的活跃流动事件 (Flows)
	flows, err := e.Store.GetActiveFlowsForAnchor(anchorID, anchor.LastCollapseTime)
	if err != nil {
		return collapsedState, err
	}

	// 4. 将所有增量 Flows 的残余势能，通过解析式衰减叠加至坍缩状态中
	for _, f := range flows {
		flowDt := float64(t-f.Timestamp) / 1000.0
		if flowDt < 0 {
			flowDt = 0
		}
		// 计算流的半衰期残余能量因子
		flowDecayFactor := math.Exp(-f.DecayRate * flowDt)

		// ★ 关键重构点：非对称超边能级映射处理
		// 如果 Flow 中针对当前节点有专属的非对称能级修正，优先使用；否则使用默认初始能量 OriginEnergy。
		flowEnergy := f.OriginEnergy
		if f.AsymmetricEnergies != nil {
			if customEnergy, exists := f.AsymmetricEnergies[anchorID]; exists {
				flowEnergy = customEnergy
			}
		}

		// 进行六维物理势能矩阵的张量求和
		collapsedState.Time += flowEnergy.Time * flowDecayFactor
		collapsedState.Space += flowEnergy.Space * flowDecayFactor
		collapsedState.PosNeg += flowEnergy.PosNeg * flowDecayFactor
		collapsedState.Influence += flowEnergy.Influence * flowDecayFactor
		collapsedState.Danger += flowEnergy.Danger * flowDecayFactor
		collapsedState.Base += flowEnergy.Base * flowDecayFactor

		// 自演化雕刻 (Self-evolution Loop)：
		if len(f.BaseEmbedding) > 0 {
			if len(anchor.BaseEmbedding) == 0 {
				anchor.BaseEmbedding = make([]float64, len(f.BaseEmbedding))
				copy(anchor.BaseEmbedding, f.BaseEmbedding)
			} else if len(anchor.BaseEmbedding) == len(f.BaseEmbedding) {
				for k := 0; k < len(anchor.BaseEmbedding); k++ {
					anchor.BaseEmbedding[k] = (1.0-e.EvolutionRate)*anchor.BaseEmbedding[k] + e.EvolutionRate*f.BaseEmbedding[k]
				}
			}
		}
	}

	// 限制能界边界约束 (积极消极在 [-1, 1], 危险度在 [0, 1])
	collapsedState.PosNeg = math.Max(-1.0, math.Min(1.0, collapsedState.PosNeg))
	collapsedState.Danger = math.Max(0.0, math.Min(1.0, collapsedState.Danger))

	// 6. 将全新的相干快照更新回 SQLite
	anchor.LastCollapseTime = t
	anchor.LastState = collapsedState
	err = e.Store.SaveAnchor(anchor)
	if err != nil {
		return collapsedState, err
	}

	return collapsedState, nil
}
