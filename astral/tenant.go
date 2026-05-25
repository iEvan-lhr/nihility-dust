package astral

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// TenantStoreManager 负责动态生命周期与缓存管理各个独立 Session 的 SQLite 连接 (方案一: Connection 级物理文件隔离)
type TenantStoreManager struct {
	mu      sync.RWMutex
	stores  map[string]AstralStore // 缓存 key: fmt.Sprintf("%s:%s", userID, sessionID)
	baseDir string                 // 数据库文件存放目录 (如 "./db_stores/")
}

// NewTenantStoreManager 初始化多租户物理库管理器
func NewTenantStoreManager(baseDir string) *TenantStoreManager {
	return &TenantStoreManager{
		stores:  make(map[string]AstralStore),
		baseDir: baseDir,
	}
}

// GetStore 动态获取或初始化指定用户与会话专属的 SQLite Store 物理连接句柄
func (m *TenantStoreManager) GetStore(userID, sessionID string) (AstralStore, error) {
	if userID == "" || sessionID == "" {
		return nil, fmt.Errorf("userID and sessionID cannot be empty")
	}

	key := fmt.Sprintf("%s:%s", userID, sessionID)

	m.mu.RLock()
	store, exists := m.stores[key]
	m.mu.RUnlock()
	if exists {
		return store, nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	// 双重校验，防止高并发下重复开辟物理文件连接
	if store, exists = m.stores[key]; exists {
		return store, nil
	}

	// 确保基础存储目录存在
	if err := os.MkdirAll(m.baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory %s: %v", m.baseDir, err)
	}

	// 安全清洗会话名称做文件名，避免特殊字符导致安全注入或目录穿透
	safeUserID := sanitizeFilename(userID)
	safeSessionID := sanitizeFilename(sessionID)
	dbPath := filepath.Join(m.baseDir, fmt.Sprintf("tenant_%s_%s.db", safeUserID, safeSessionID))

	newStore, err := NewSQLiteAstralStore(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to init store for tenant [%s]: %v", key, err)
	}

	m.stores[key] = newStore
	return newStore, nil
}

// DeleteStore 安全关闭租户 Store 句柄并直接从磁盘物理删除该对话对应的数据库文件 (一键物理清空对话)
func (m *TenantStoreManager) DeleteStore(userID, sessionID string) error {
	if userID == "" || sessionID == "" {
		return fmt.Errorf("userID and sessionID cannot be empty")
	}

	key := fmt.Sprintf("%s:%s", userID, sessionID)

	m.mu.Lock()
	store, exists := m.stores[key]
	if exists {
		_ = store.Close()
		delete(m.stores, key)
	}
	m.mu.Unlock()

	safeUserID := sanitizeFilename(userID)
	safeSessionID := sanitizeFilename(sessionID)
	dbPath := filepath.Join(m.baseDir, fmt.Sprintf("tenant_%s_%s.db", safeUserID, safeSessionID))

	// 如果文件存在，执行物理删除
	if _, err := os.Stat(dbPath); err == nil {
		if err := os.Remove(dbPath); err != nil {
			return fmt.Errorf("failed to remove physical database file: %v", err)
		}
	}
	return nil
}

// CloseAll 在系统关闭时，安全优雅地关闭所有已打开的物理库连接句柄并清空连接池
func (m *TenantStoreManager) CloseAll() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var lastErr error
	for key, store := range m.stores {
		if err := store.Close(); err != nil {
			lastErr = err
		}
		delete(m.stores, key)
	}
	return lastErr
}

// sanitizeFilename 简易的安全字符清洗器，过滤掉潜在的路径穿透符号
func sanitizeFilename(s string) string {
	res := make([]rune, 0, len(s))
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			res = append(res, r)
		}
	}
	if len(res) == 0 {
		return "default"
	}
	return string(res)
}
