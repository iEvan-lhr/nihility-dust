package astral

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/iEvan-lhr/nihility-dust/anything"
)

// VectorEncoder 向量编码器接口，用于将自然语言文本转化为高维语义向量 (Embedding)
type VectorEncoder interface {
	Encode(ctx context.Context, text string) ([]float64, error)
}

// SimpleMockEncoder 纯 Go 实现的通用特征哈希算法 (Feature Hashing)
type SimpleMockEncoder struct {
	Dimension int
}

// NewSimpleMockEncoder 初始化指定维度的编码器
func NewSimpleMockEncoder(dim int) *SimpleMockEncoder {
	return &SimpleMockEncoder{Dimension: dim}
}

// Encode 基于 DJB2 散列的特征哈希与单字混合算法，将任意输入文本转化为标准化高维语义向量，完美支持中英文
func (e *SimpleMockEncoder) Encode(ctx context.Context, text string) ([]float64, error) {
	vec := make([]float64, e.Dimension)
	cleanText := strings.Map(func(r rune) rune {
		if strings.ContainsRune("!@#$%^&*()_+-=[]{}|;':\",./<>?`~，。？！（）—…", r) {
			return ' '
		}
		return r
	}, strings.ToLower(text))

	words := strings.Fields(cleanText)
	if len(words) == 0 {
		return vec, nil
	}

	for _, w := range words {
		// 1. 单词/词组整体哈希
		var hash uint32 = 5381
		for i := 0; i < len(w); i++ {
			hash = ((hash << 5) + hash) + uint32(w[i])
		}
		idx := int(hash % uint32(e.Dimension))
		vec[idx] += 1.0

		// 2. 针对中文等非 ASCII 字符，单独进行单字哈希，避免中文无空格分词导致的检索失真
		runes := []rune(w)
		for _, r := range runes {
			if r > 127 {
				var charHash uint32 = 5381
				charHash = ((charHash << 5) + charHash) + uint32(r)
				cIdx := int(charHash % uint32(e.Dimension))
				vec[cIdx] += 1.0
			}
		}
	}

	var sum float64
	for _, val := range vec {
		sum += val * val
	}
	if sum > 0 {
		norm := math.Sqrt(sum)
		for i := range vec {
			vec[i] /= norm
		}
	}
	return vec, nil
}

// ActiveEvent 描述穿透节点的未衰减活跃流的结构，用于向外部大模型提供时空记忆上下文
type ActiveEvent struct {
	ID                  int64    `json:"id"`
	Payload             string   `json:"payload"`               // 事件文本载荷
	Timestamp           int64    `json:"timestamp"`             // 事件发生时间戳 (毫秒)
	DtSeconds           float64  `json:"dt_seconds"`            // 距离当前观测时间的时间差 (秒)
	ResidualEnergyRatio float64  `json:"residual_energy_ratio"` // 衰减后剩余的残余温能百分比 [0.0 - 1.0]
	OriginEnergy        Vector6D `json:"origin_energy"`         // 初始六维能谱
}

// CollapsedContext 瞬间坍缩状态与活跃事件上下文
type CollapsedContext struct {
	AnchorID       int64         `json:"anchor_id"`       // 被观测节点坐标 ID
	Description    string        `json:"description"`     // 节点本体描述
	CollapsedState Vector6D      `json:"collapsed_state"` // 这一微秒坍缩出的六维状态能谱
	ActiveEvents   []ActiveEvent `json:"active_events"`   // 当前仍然起作用的活跃流载荷记忆
}

// AstralPipeline 星空拓扑六维空间物理管线 (专注于时空场的存储、映射、向量检索与量子延迟相干坍缩)
type AstralPipeline struct {
	Store          AstralStore
	CollapseEngine *CollapseEngine
	Encoder        VectorEncoder
}

// NewAstralPipeline 初始化通用的物理管线
func NewAstralPipeline(store AstralStore, engine *CollapseEngine, encoder VectorEncoder) *AstralPipeline {
	return &AstralPipeline{
		Store:          store,
		CollapseEngine: engine,
		Encoder:        encoder,
	}
}

// RegisterEntity 注册一个新的基态真空节点，若 id 为 0，则自动使用高并发 Snowflake 唯一 ID 防止冲突
func (p *AstralPipeline) RegisterEntity(ctx context.Context, id int64, description string) (*NodeAnchor, error) {
	var realID int64
	if id == 0 {
		realID = anything.GetId() // 自动分发唯一雪花 ID
	} else {
		realID = id
	}

	embedding, err := p.Encoder.Encode(ctx, description)
	if err != nil {
		return nil, fmt.Errorf("failed to encode description: %v", err)
	}

	anchor := &NodeAnchor{
		ID:               realID,
		Description:      description,
		BaseEmbedding:    embedding,
		LastCollapseTime: time.Now().UnixNano() / 1e6,
		LastState:        Vector6D{Time: 0, Space: 0, PosNeg: 0, Influence: 0, Danger: 0, Base: 0},
	}

	if err := p.Store.SaveAnchor(anchor); err != nil {
		return nil, fmt.Errorf("failed to save anchor: %v", err)
	}

	return anchor, nil
}

// EmitEvent 发射一束事件流动 (Flow)，关联被照亮的节点超边
func (p *AstralPipeline) EmitEvent(ctx context.Context, anchors []int64, payload string, decayRate float64, energy Vector6D, asymmetric map[int64]Vector6D) (*Flow, error) {
	embedding, err := p.Encoder.Encode(ctx, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to encode payload: %v", err)
	}

	flow := &Flow{
		Anchors:            anchors,
		Payload:            payload,
		Timestamp:          time.Now().UnixNano() / 1e6,
		OriginEnergy:       energy,
		DecayRate:          decayRate,
		BaseEmbedding:      embedding,
		AsymmetricEnergies: asymmetric,
	}

	if err := p.Store.SaveFlow(flow); err != nil {
		return nil, fmt.Errorf("failed to save flow: %v", err)
	}

	return flow, nil
}

// GenerateTempIDMapping 实用工具：当大模型输出临时节点 ID 时，为它们自动映射分配唯一的真实雪花 ID
func (p *AstralPipeline) GenerateTempIDMapping(tempIDs []string) map[string]int64 {
	mapping := make(map[string]int64)
	for _, tempID := range tempIDs {
		if strings.HasPrefix(tempID, "$") {
			mapping[tempID] = anything.GetId() // 分发唯一 ID
		}
	}
	return mapping
}

// CollapseToContext 核心量子观测接口：计算六维坍缩能量，并过滤提取当前未衰减的活跃流文本记忆上下文。
// 该结构体支持一键 JSON 序列化，供您外部的任何大模型 Prompt 模板调用。
func (p *AstralPipeline) CollapseToContext(ctx context.Context, anchorID int64) (*CollapsedContext, error) {
	nowMs := time.Now().UnixNano() / 1e6

	// 1. 量子延迟坍缩计算六维绝对势能
	collapsedState, err := p.CollapseEngine.Collapse(anchorID, nowMs)
	if err != nil {
		return nil, fmt.Errorf("failed to collapse: %v", err)
	}

	// 2. 读取节点元数据
	anchor, err := p.Store.GetAnchor(anchorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get anchor metadata: %v", err)
	}

	// 3. 拉取并过滤穿透节点的活跃 flows 记忆 (排除能量 < 2% 的死流)
	flows, err := p.Store.GetActiveFlowsForAnchor(anchorID, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get active flows: %v", err)
	}

	var activeEvents []ActiveEvent
	for _, f := range flows {
		dtSec := float64(nowMs-f.Timestamp) / 1000.0
		if dtSec < 0 {
			dtSec = 0
		}
		residual := math.Exp(-f.DecayRate * dtSec)

		if residual >= 0.02 { // 仅保留当前仍有能量波及影响的流
			activeEvents = append(activeEvents, ActiveEvent{
				ID:                  f.ID,
				Payload:             f.Payload,
				Timestamp:           f.Timestamp,
				DtSeconds:           dtSec,
				ResidualEnergyRatio: residual,
				OriginEnergy:        f.OriginEnergy,
			})
		}
	}

	return &CollapsedContext{
		AnchorID:       anchorID,
		Description:    anchor.Description,
		CollapsedState: collapsedState,
		ActiveEvents:   activeEvents,
	}, nil
}

// KnowledgeItem 供 AI 使用的单条知识载荷与关联元数据
type KnowledgeItem struct {
	FlowID     int64   `json:"flow_id"`
	Payload    string  `json:"payload"`    // 事件文本/文档载荷
	Similarity float64 `json:"similarity"` // 检索匹配的引力相似度
	Timestamp  int64   `json:"timestamp"`  // 发生时间戳
	Anchors    []int64 `json:"anchors"`    // 关联的物理节点 ID 列表
}

// SearchKnowledge 依据高维语义引力，在持久化 Flow 库中匹配出 Top-K 个相关的事件文档/知识 Payload，直接供 AI 拼接为上下文知识
func (p *AstralPipeline) SearchKnowledge(ctx context.Context, query string, limit int) ([]KnowledgeItem, error) {
	qVec, err := p.Encoder.Encode(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to encode search query: %v", err)
	}
	results, err := p.Store.SearchFlowsByRelativity(qVec, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search flows: %v", err)
	}
	var items []KnowledgeItem
	for _, r := range results {
		items = append(items, KnowledgeItem{
			FlowID:     r.Flow.ID,
			Payload:    r.Flow.Payload,
			Similarity: r.Similarity,
			Timestamp:  r.Flow.Timestamp,
			Anchors:    r.Flow.Anchors,
		})
	}
	return items, nil
}
