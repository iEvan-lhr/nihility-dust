package anything

import (
	"database/sql"
	"fmt"
	"strings"

	// 引入 sqlite3 驱动
	_ "github.com/mattn/go-sqlite3"
)

// ScriptStore 定义 DSL 脚本的存储接口
// 它是 Wind 与底层数据库交互的唯一窗口
type ScriptStore interface {
	// Get 获取指定脚本的内容
	Get(name string) (string, error)

	// Save 保存脚本
	// name: 唯一标识 (如 "Skill.CrawlBaidu")
	// content: JSON DSL 内容
	// tags: 用于分类检索 (如 "network,crawl")
	// desc: 自然语言描述，用于 AI 理解 (如 "爬取百度首页")
	// sig: 函数签名，用于 AI 生成代码 (如 "Skill.CrawlBaidu()")
	Save(name, content, tags, desc, sig string) error

	// Search 通过标签模糊搜索 (用于人工/逻辑批量加载)
	// tagPattern: 如 "crawl"，将匹配 "%crawl%"
	Search(tagPattern string) (map[string]string, error)

	// SearchByKeywords 根据关键词列表进行 AI 关联检索 (RAG 核心)
	// keywords: AI 提取的关键词列表 (如 ["baidu", "network"])
	// 返回: 匹配到的工具元数据列表
	SearchByKeywords(keywords []string) ([]ToolMeta, error)
}

// SQLiteStore 基于 SQLite 的实现
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore 初始化数据库连接并自动迁移表结构
func NewSQLiteStore(filepath string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite3", filepath)
	if err != nil {
		return nil, err
	}

	// 自动建表：支持存储 DSL 内容、标签、描述和签名
	query := `
	CREATE TABLE IF NOT EXISTS scripts (
		name TEXT PRIMARY KEY,
		content TEXT,
		tags TEXT,
		description TEXT,
		signature TEXT
	);`
	if _, err := db.Exec(query); err != nil {
		return nil, fmt.Errorf("failed to create table: %v", err)
	}

	return &SQLiteStore{db: db}, nil
}

// Get 获取脚本内容
func (s *SQLiteStore) Get(name string) (string, error) {
	var content string
	err := s.db.QueryRow("SELECT content FROM scripts WHERE name = ?", name).Scan(&content)
	if err != nil {
		return "", err
	}
	return content, nil
}

// Save 保存或更新脚本信息
func (s *SQLiteStore) Save(name, content, tags, desc, sig string) error {
	// 使用 REPLACE INTO 确保 name 唯一，重复则覆盖
	query := `
		INSERT OR REPLACE INTO scripts (name, content, tags, description, signature) 
		VALUES (?, ?, ?, ?, ?)
	`
	_, err := s.db.Exec(query, name, content, tags, desc, sig)
	return err
}

// Search 通过标签进行模糊搜索 (返回 map[name]content)
// 主要用于 DSLSpirit.LoadByTag
func (s *SQLiteStore) Search(tagPattern string) (map[string]string, error) {
	queryTag := "%" + tagPattern + "%"

	rows, err := s.db.Query("SELECT name, content FROM scripts WHERE tags LIKE ?", queryTag)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make(map[string]string)
	for rows.Next() {
		var name, content string
		if err := rows.Scan(&name, &content); err != nil {
			continue
		}
		results[name] = content
	}
	return results, nil
}

// SearchByKeywords 核心 RAG 检索逻辑
// 根据 AI 提取的关键词，在 Name, Tags, Description 中进行宽泛匹配
func (s *SQLiteStore) SearchByKeywords(keywords []string) ([]ToolMeta, error) {
	if len(keywords) == 0 {
		return nil, nil
	}

	// 动态构建 SQL 查询语句
	// 目标 SQL: SELECT ... WHERE (tags LIKE ? OR desc LIKE ? ...) OR (tags LIKE ? ...)
	var queryParts []string
	var args []any

	for _, k := range keywords {
		kPattern := "%" + k + "%"
		// 对每个关键词，尝试匹配三个字段
		queryParts = append(queryParts, "(tags LIKE ? OR description LIKE ? OR name LIKE ?)")
		args = append(args, kPattern, kPattern, kPattern)
	}

	query := fmt.Sprintf(
		"SELECT name, description, signature FROM scripts WHERE %s LIMIT 10",
		strings.Join(queryParts, " OR "),
	)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 使用 Map 去重 (防止多个关键词匹配到同一个工具)
	uniqueTools := make(map[string]ToolMeta)

	for rows.Next() {
		var t ToolMeta
		// 注意：Signature 可能会为空，数据库里可能存 NULL，Scan 需要处理
		// 这里假设我们 Save 时存的是空字符串而非 NULL
		if err := rows.Scan(&t.Name, &t.Description, &t.Signature); err == nil {
			uniqueTools[t.Name] = t
		}
	}

	// 转换为 Slice 返回
	var results []ToolMeta
	for _, t := range uniqueTools {
		results = append(results, t)
	}

	return results, nil
}
