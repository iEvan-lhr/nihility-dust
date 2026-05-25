package astral

import "math"

// Vector6D 六维核心状态向量
type Vector6D struct {
	Time      float64 `json:"time"`      // 时间有效期能 / 时间压力 (如：工期交付节点迫近造成的紧迫度)
	Space     float64 `json:"space"`     // 空间/虚拟结界阻尼 (如：跨地域上海到北京的空间距离能量衰竭度)
	PosNeg    float64 `json:"pos_neg"`   // 积极与消极色彩能级，情绪色彩与偏置系数 [-1.0, 1.0]
	Influence float64 `json:"influence"` // 支配与波及能量大小 (如：技术主导权)
	Danger    float64 `json:"danger"`    // 不稳定与破坏性量化 [0.0, 1.0] (如：线上Bug、创业失败风险)
	Base      float64 `json:"base"`      // 基础本体语义的物理表征值
}

// NodeAnchor 孤立的星空坐标锚点 (业务本体的物理代表坐标，包含原始文本描述，完美支持通用化检索)
type NodeAnchor struct {
	ID               int64     `json:"id"`
	Description      string    `json:"description"`        // ★ 新增：人类可读的本体语义描述/名称，用于引力波和全量坐标检索
	BaseEmbedding    []float64 `json:"base_embedding"`     // 节点的基态引力向量 (LLM 检索与定位的锚点，存储语义表征)
	LastCollapseTime int64     `json:"last_collapse_time"` // 上次观测相干坍缩发生的时间戳 (毫秒)
	LastState        Vector6D  `json:"last_state"`         // 上次相干状态的快照 (增量坍缩计算的基准点)
}

// Flow 跨越时空的流动事件载体 (系统存储的唯一物理实体，代表一束光照亮并波及多个 Anchor)
type Flow struct {
	ID                 int64              `json:"id"`
	Anchors            []int64            `json:"anchors"`             // 被该流动事件瞬间照亮并卷入的所有节点坐标 ID (天然支持多元超边)
	Payload            string             `json:"payload"`             // 流动携带的高维载荷内容 (如：文档片段、告警日志)
	Timestamp          int64              `json:"timestamp"`           // 发射时间戳 (毫秒)
	OriginEnergy       Vector6D           `json:"origin_energy"`       // 事件发生时的默认初始六维能量张量
	DecayRate          float64            `json:"decay_rate"`          // 该事件的半衰衰减系数常数
	BaseEmbedding      []float64          `json:"base_embedding"`      // 事件载荷的高维文本语义向量，用于引力波检索
	AsymmetricEnergies map[int64]Vector6D `json:"asymmetric_energies"` // 节点专属的非对称能级映射 (Key: AnchorID)，完美支持双向独立非对称关系
}

// CosineSimilarity 高效计算两个向量的余弦相似度 (纯 Go 实现，用于引力波检索)
func CosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0.0
	}
	var dotProduct, normA, normB float64
	for i := 0; i < len(a); i++ {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	if normA == 0.0 || normB == 0.0 {
		return 0.0
	}
	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}
