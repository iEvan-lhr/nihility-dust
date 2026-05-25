package astral

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// AstralStore 统一抽象的存储接口，解耦具体数据库实现
type AstralStore interface {
	SaveFlow(f *Flow) error
	SaveAnchor(a *NodeAnchor) error
	GetAnchor(id int64) (*NodeAnchor, error)
	GetAllAnchors() ([]NodeAnchor, error)
	GetActiveFlowsForAnchor(anchorID int64, sinceTime int64) ([]Flow, error)
	SearchByRelativity(targetEmbedding []float64, limit int) ([]struct {
		Anchor     *NodeAnchor
		Similarity float64
	}, error)
	SearchFlowsByRelativity(targetEmbedding []float64, limit int) ([]struct {
		Flow       Flow
		Similarity float64
	}, error)
	Close() error
}

// StoreConfig 数据库配置结构体，支持完全独立的自定义注入与初始化
type StoreConfig struct {
	DriverName string  // 驱动名称: "sqlite3" 或 "postgres"
	DSN        string  // 数据库连接字符串 (如 PostgreSQL 的 "postgres://user:pwd@localhost:5432/db?sslmode=disable")
	DB         *sql.DB // (★极其重要) 可选：直接注入外部已初始化完毕的 *sql.DB 连接池，实现企业级微服务无缝整合
}

// GenericSQLStore 通用 SQL 存储实现，完美兼容 SQLite 与 PostgreSQL
type GenericSQLStore struct {
	db         *sql.DB
	driverName string
}

// NewGenericSQLStore 基于结构体配置来初始化存储
func NewGenericSQLStore(cfg StoreConfig) (*GenericSQLStore, error) {
	var db *sql.DB
	var err error

	// 1. 如果外部注入了 DB，直接复用连接池
	if cfg.DB != nil {
		db = cfg.DB
		if cfg.DriverName == "" {
			cfg.DriverName = "sqlite3" // 默认退回 sqlite3
		}
	} else {
		// 2. 否则，根据配置动态打开数据库连接
		if cfg.DriverName == "" {
			cfg.DriverName = "sqlite3"
		}
		if cfg.DSN == "" {
			cfg.DSN = "dust.db"
		}
		db, err = sql.Open(cfg.DriverName, cfg.DSN)
		if err != nil {
			return nil, fmt.Errorf("failed to open database pool: %v", err)
		}
	}

	store := &GenericSQLStore{
		db:         db,
		driverName: strings.ToLower(cfg.DriverName),
	}

	// 3. 执行符合特定数据库方言（Dialect）的初始化建表
	if err := store.bootstrap(); err != nil {
		if cfg.DB == nil {
			db.Close()
		}
		return nil, err
	}

	return store, nil
}

// Close 关闭连接池
func (s *GenericSQLStore) Close() error {
	return s.db.Close()
}

// bootstrap 自动建表与升级
func (s *GenericSQLStore) bootstrap() error {
	var flowTableDDL, anchorTableDDL string

	if s.driverName == "postgres" {
		// PostgreSQL 数据类型方言
		flowTableDDL = `CREATE TABLE IF NOT EXISTS astral_flows (
			id BIGSERIAL PRIMARY KEY,
			anchors TEXT NOT NULL,
			payload TEXT,
			timestamp BIGINT NOT NULL,
			decay_rate DOUBLE PRECISION NOT NULL,
			origin_energy TEXT NOT NULL,
			base_embedding TEXT,
			asymmetric_energies TEXT
		);`
		anchorTableDDL = `CREATE TABLE IF NOT EXISTS astral_anchors (
			id BIGINT PRIMARY KEY,
			description TEXT NOT NULL DEFAULT '',
			base_embedding TEXT NOT NULL,
			last_collapse_time BIGINT NOT NULL,
			last_state TEXT NOT NULL
		);`
	} else {
		// SQLite 数据类型方言
		flowTableDDL = `CREATE TABLE IF NOT EXISTS astral_flows (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			anchors TEXT NOT NULL,
			payload TEXT,
			timestamp INTEGER NOT NULL,
			decay_rate REAL NOT NULL,
			origin_energy TEXT NOT NULL,
			base_embedding TEXT,
			asymmetric_energies TEXT
		);`
		anchorTableDDL = `CREATE TABLE IF NOT EXISTS astral_anchors (
			id INTEGER PRIMARY KEY,
			description TEXT NOT NULL DEFAULT '',
			base_embedding TEXT NOT NULL,
			last_collapse_time INTEGER NOT NULL,
			last_state TEXT NOT NULL
		);`
	}

	queries := []string{
		flowTableDDL,
		anchorTableDDL,
		`CREATE INDEX IF NOT EXISTS idx_flows_timestamp ON astral_flows(timestamp);`,
	}

	for _, q := range queries {
		if _, err := s.db.Exec(q); err != nil {
			return fmt.Errorf("failed to bootstrap table: %v", err)
		}
	}

	// 增量升级：如果 astral_anchors 表已存在但缺失 description，执行增量添加并忽略已存在错误
	if s.driverName != "postgres" {
		_, _ = s.db.Exec("ALTER TABLE astral_anchors ADD COLUMN description TEXT NOT NULL DEFAULT '';")
	} else {
		// PostgreSQL 安全增量添加字段
		alterQuery := `
		DO $$ 
		BEGIN 
			BEGIN
				ALTER TABLE astral_anchors ADD COLUMN description TEXT NOT NULL DEFAULT '';
			EXCEPTION
				WHEN duplicate_column THEN NULL;
			END;
		END $$;`
		_, _ = s.db.Exec(alterQuery)
	}

	return nil
}

// translate 占位符运行时翻译器，将统一的 '?' 语法糖动态转换为 PostgreSQL 专属的 '$1, $2' 占位符
func (s *GenericSQLStore) translate(query string) string {
	if s.driverName == "postgres" {
		n := 1
		for strings.Contains(query, "?") {
			query = strings.Replace(query, "?", fmt.Sprintf("$%d", n), 1)
			n++
		}
	}
	return query
}

// SaveFlow 记录流动事件
func (s *GenericSQLStore) SaveFlow(f *Flow) error {
	var sb strings.Builder
	sb.WriteByte(',')
	for _, id := range f.Anchors {
		sb.WriteString(strconv.FormatInt(id, 10))
		sb.WriteByte(',')
	}
	anchorsStr := sb.String()

	energyJSON, err := json.Marshal(f.OriginEnergy)
	if err != nil {
		return fmt.Errorf("failed to marshal flow energy: %v", err)
	}

	embeddingJSON, err := json.Marshal(f.BaseEmbedding)
	if err != nil {
		return fmt.Errorf("failed to marshal flow embedding: %v", err)
	}

	var asymmetricJSON []byte
	if f.AsymmetricEnergies != nil {
		asymmetricJSON, err = json.Marshal(f.AsymmetricEnergies)
		if err != nil {
			return fmt.Errorf("failed to marshal asymmetric energies: %v", err)
		}
	} else {
		asymmetricJSON = []byte("{}")
	}

	query := s.translate(`INSERT INTO astral_flows (anchors, payload, timestamp, decay_rate, origin_energy, base_embedding, asymmetric_energies)
	          VALUES (?, ?, ?, ?, ?, ?, ?)`)

	res, err := s.db.Exec(query, anchorsStr, f.Payload, f.Timestamp, f.DecayRate, string(energyJSON), string(embeddingJSON), string(asymmetricJSON))
	if err != nil {
		return fmt.Errorf("failed to insert flow: %v", err)
	}

	// SQLite 支持 LastInsertId，PostgreSQL 在 BIGSERIAL 下通常无需回写 ID 或可用 RETURNING，这里做安全兼容
	if s.driverName != "postgres" {
		id, err := res.LastInsertId()
		if err == nil {
			f.ID = id
		}
	}
	return nil
}

// SaveAnchor 初始注册或完全覆盖节点坐标 (ON CONFLICT 语法完美兼容 SQLite 与 PostgreSQL)
func (s *GenericSQLStore) SaveAnchor(a *NodeAnchor) error {
	embeddingJSON, err := json.Marshal(a.BaseEmbedding)
	if err != nil {
		return fmt.Errorf("failed to marshal anchor embedding: %v", err)
	}

	stateJSON, err := json.Marshal(a.LastState)
	if err != nil {
		return fmt.Errorf("failed to marshal anchor state: %v", err)
	}

	// 使用标准 ANSI SQL 兼容的 ON CONFLICT 语法，完美实现防重复与覆盖
	query := s.translate(`INSERT INTO astral_anchors (id, description, base_embedding, last_collapse_time, last_state)
	          VALUES (?, ?, ?, ?, ?)
	          ON CONFLICT (id) DO UPDATE SET 
	             description = excluded.description, 
	             base_embedding = excluded.base_embedding, 
	             last_collapse_time = excluded.last_collapse_time, 
	             last_state = excluded.last_state`)

	_, err = s.db.Exec(query, a.ID, a.Description, string(embeddingJSON), a.LastCollapseTime, string(stateJSON))
	return err
}

// GetAnchor 获取坐标快照
func (s *GenericSQLStore) GetAnchor(id int64) (*NodeAnchor, error) {
	query := s.translate(`SELECT description, base_embedding, last_collapse_time, last_state FROM astral_anchors WHERE id = ?`)
	row := s.db.QueryRow(query, id)

	var description, embeddingStr, stateStr string
	var lastCollapseTime int64
	err := row.Scan(&description, &embeddingStr, &lastCollapseTime, &stateStr)
	if err == sql.ErrNoRows {
		return &NodeAnchor{
			ID:               id,
			Description:      "Virtual Unknown Anchor",
			BaseEmbedding:    make([]float64, 0),
			LastCollapseTime: 0,
			LastState:        Vector6D{Time: 0, Space: 0, PosNeg: 0, Influence: 0, Danger: 0, Base: 0},
		}, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to get anchor: %v", err)
	}

	var baseEmbedding []float64
	if err := json.Unmarshal([]byte(embeddingStr), &baseEmbedding); err != nil {
		return nil, fmt.Errorf("failed to unmarshal anchor embedding: %v", err)
	}

	var lastState Vector6D
	if err := json.Unmarshal([]byte(stateStr), &lastState); err != nil {
		return nil, fmt.Errorf("failed to unmarshal anchor state: %v", err)
	}

	return &NodeAnchor{
		ID:               id,
		Description:      description,
		BaseEmbedding:    baseEmbedding,
		LastCollapseTime: lastCollapseTime,
		LastState:        lastState,
	}, nil
}

// GetAllAnchors 获取全量锚点
func (s *GenericSQLStore) GetAllAnchors() ([]NodeAnchor, error) {
	query := s.translate(`SELECT id, description, base_embedding, last_collapse_time, last_state FROM astral_anchors`)
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var anchors []NodeAnchor
	for rows.Next() {
		var id, lastCollapseTime int64
		var description, embeddingStr, stateStr string

		if err := rows.Scan(&id, &description, &embeddingStr, &lastCollapseTime, &stateStr); err != nil {
			continue
		}

		var baseEmbedding []float64
		_ = json.Unmarshal([]byte(embeddingStr), &baseEmbedding)

		var lastState Vector6D
		_ = json.Unmarshal([]byte(stateStr), &lastState)

		anchors = append(anchors, NodeAnchor{
			ID:               id,
			Description:      description,
			BaseEmbedding:    baseEmbedding,
			LastCollapseTime: lastCollapseTime,
			LastState:        lastState,
		})
	}
	return anchors, nil
}

// GetActiveFlowsForAnchor 增量拉取活跃 Flow
func (s *GenericSQLStore) GetActiveFlowsForAnchor(anchorID int64, sinceTime int64) ([]Flow, error) {
	matchPattern := "%," + strconv.FormatInt(anchorID, 10) + ",%"

	query := s.translate(`SELECT id, anchors, payload, timestamp, decay_rate, origin_energy, base_embedding, asymmetric_energies
	          FROM astral_flows
	          WHERE timestamp >= ? AND anchors LIKE ?`)
	rows, err := s.db.Query(query, sinceTime, matchPattern)
	if err != nil {
		return nil, fmt.Errorf("failed to query active flows: %v", err)
	}
	defer rows.Close()

	var flows []Flow
	for rows.Next() {
		var id, timestamp int64
		var decayRate float64
		var anchorsStr, payload, energyStr, embeddingStr, asymmetricStr string

		err := rows.Scan(&id, &anchorsStr, &payload, &timestamp, &decayRate, &energyStr, &embeddingStr, &asymmetricStr)
		if err != nil {
			return nil, fmt.Errorf("failed to scan flow: %v", err)
		}

		parts := strings.Split(anchorsStr, ",")
		var anchors []int64
		for _, p := range parts {
			if p == "" {
				continue
			}
			nodeID, err := strconv.ParseInt(p, 10, 64)
			if err == nil {
				anchors = append(anchors, nodeID)
			}
		}

		var originEnergy Vector6D
		if err := json.Unmarshal([]byte(energyStr), &originEnergy); err != nil {
			return nil, fmt.Errorf("failed to unmarshal energy: %v", err)
		}

		var baseEmbedding []float64
		if embeddingStr != "" {
			_ = json.Unmarshal([]byte(embeddingStr), &baseEmbedding)
		}

		var asymmetricEnergies map[int64]Vector6D
		if asymmetricStr != "" {
			_ = json.Unmarshal([]byte(asymmetricStr), &asymmetricEnergies)
		}

		flows = append(flows, Flow{
			ID:                 id,
			Anchors:            anchors,
			Payload:            payload,
			Timestamp:          timestamp,
			OriginEnergy:       originEnergy,
			DecayRate:          decayRate,
			BaseEmbedding:      baseEmbedding,
			AsymmetricEnergies: asymmetricEnergies,
		})
	}

	return flows, nil
}

// SearchByRelativity 语义引力检索
func (s *GenericSQLStore) SearchByRelativity(targetEmbedding []float64, limit int) ([]struct {
	Anchor     *NodeAnchor
	Similarity float64
}, error) {
	query := s.translate(`SELECT id, description, base_embedding, last_collapse_time, last_state FROM astral_anchors`)
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []struct {
		Anchor     *NodeAnchor
		Similarity float64
	}

	for rows.Next() {
		var id, lastCollapseTime int64
		var description, embeddingStr, stateStr string

		if err := rows.Scan(&id, &description, &embeddingStr, &lastCollapseTime, &stateStr); err != nil {
			continue
		}

		var baseEmbedding []float64
		_ = json.Unmarshal([]byte(embeddingStr), &baseEmbedding)

		if len(baseEmbedding) == 0 {
			continue
		}

		var lastState Vector6D
		_ = json.Unmarshal([]byte(stateStr), &lastState)

		anchor := &NodeAnchor{
			ID:               id,
			Description:      description,
			BaseEmbedding:    baseEmbedding,
			LastCollapseTime: lastCollapseTime,
			LastState:        lastState,
		}

		sim := CosineSimilarity(targetEmbedding, baseEmbedding)
		results = append(results, struct {
			Anchor     *NodeAnchor
			Similarity float64
		}{Anchor: anchor, Similarity: sim})
	}

	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[i].Similarity < results[j].Similarity {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	if len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// SearchFlowsByRelativity 根据引力波进行 RAG 瞬时向量检索流动事件 (检索具体的文档内容/Payload，作为大模型的上下文知识)
func (s *GenericSQLStore) SearchFlowsByRelativity(targetEmbedding []float64, limit int) ([]struct {
	Flow       Flow
	Similarity float64
}, error) {
	query := s.translate(`SELECT id, anchors, payload, timestamp, decay_rate, origin_energy, base_embedding, asymmetric_energies FROM astral_flows`)
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []struct {
		Flow       Flow
		Similarity float64
	}

	for rows.Next() {
		var id, timestamp int64
		var decayRate float64
		var anchorsStr, payload, energyStr, embeddingStr, asymmetricStr string

		if err := rows.Scan(&id, &anchorsStr, &payload, &timestamp, &decayRate, &energyStr, &embeddingStr, &asymmetricStr); err != nil {
			continue
		}

		var baseEmbedding []float64
		_ = json.Unmarshal([]byte(embeddingStr), &baseEmbedding)
		if len(baseEmbedding) == 0 {
			continue
		}

		var originEnergy Vector6D
		_ = json.Unmarshal([]byte(energyStr), &originEnergy)

		parts := strings.Split(anchorsStr, ",")
		var anchors []int64
		for _, p := range parts {
			if p == "" {
				continue
			}
			nodeID, err := strconv.ParseInt(p, 10, 64)
			if err == nil {
				anchors = append(anchors, nodeID)
			}
		}

		var asymmetricEnergies map[int64]Vector6D
		if asymmetricStr != "" {
			_ = json.Unmarshal([]byte(asymmetricStr), &asymmetricEnergies)
		}

		f := Flow{
			ID:                 id,
			Anchors:            anchors,
			Payload:            payload,
			Timestamp:          timestamp,
			OriginEnergy:       originEnergy,
			DecayRate:          decayRate,
			BaseEmbedding:      baseEmbedding,
			AsymmetricEnergies: asymmetricEnergies,
		}

		sim := CosineSimilarity(targetEmbedding, baseEmbedding)
		results = append(results, struct {
			Flow       Flow
			Similarity float64
		}{Flow: f, Similarity: sim})
	}

	// 降序对相似度进行排序
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[i].Similarity < results[j].Similarity {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	if len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// NewSQLiteAstralStore 向上兼容旧方法，底层转发给 GenericSQLStore
func NewSQLiteAstralStore(filepath string) (AstralStore, error) {
	return NewGenericSQLStore(StoreConfig{
		DriverName: "sqlite3",
		DSN:        filepath,
	})
}
