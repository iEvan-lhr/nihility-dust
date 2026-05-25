package astral

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestTenantStoreManager_IsolationAndLifecycle(t *testing.T) {
	ctx := context.Background()

	// 1. 设置专属的测试用临时多租户目录
	tempDir := filepath.Join(".", "test_tenant_db_stores")
	defer os.RemoveAll(tempDir) // 运行完后自动销毁测试文件目录

	manager := NewTenantStoreManager(tempDir)
	defer manager.CloseAll()

	// 2. 动态创建 User_A 的 Session_1
	storeA1, err := manager.GetStore("User_A", "Session_1")
	if err != nil {
		t.Fatalf("failed to get store for User_A Session_1: %v", err)
	}

	// 3. 动态创建 User_B 的 Session_1 (两个完全独立的用户)
	storeB1, err := manager.GetStore("User_B", "Session_1")
	if err != nil {
		t.Fatalf("failed to get store for User_B Session_1: %v", err)
	}

	// 4. 验证缓存池复用机制 (再次获取 A1，应该得到相同的实例)
	storeA1Second, err := manager.GetStore("User_A", "Session_1")
	if err != nil {
		t.Fatalf("failed to get store A1 second time: %v", err)
	}
	if storeA1 != storeA1Second {
		t.Errorf("expected same store instance for identical tenant key, but got different references")
	}

	// -------------------------------------------------------------------------
	// 5. 验证绝对数据隔离性
	// -------------------------------------------------------------------------
	// 在 User_A 的 Session_1 中注册一个基态锚点
	anchorA := &NodeAnchor{
		ID:          999,
		Description: "User A's Exclusive Node",
	}
	if err := storeA1.SaveAnchor(anchorA); err != nil {
		t.Fatalf("failed to save anchor in User A's store: %v", err)
	}

	// 校验 A1 库中存在此锚点
	gotAnchorA, err := storeA1.GetAnchor(999)
	if err != nil || gotAnchorA.Description != "User A's Exclusive Node" {
		t.Fatalf("failed to retrieve anchor from User A's store or description mismatch: %v", err)
	}

	// 🚨 强行到 User_B 的 Session_1 库中读取 999 号锚点，理应拿到空白基态 (隔离成功!)
	gotAnchorB, err := storeB1.GetAnchor(999)
	if err != nil {
		t.Fatalf("reading non-existent anchor in B1 should return empty base anchor instead of database error: %v", err)
	}
	if gotAnchorB.Description == "User A's Exclusive Node" {
		t.Errorf("🚨 严重安全漏洞: User_B 读取到了 User_A 的专属物理库数据!")
	}
	t.Log("✔ 验证隔离性成功：不同用户的物理库数据完全独立隔离。")

	// -------------------------------------------------------------------------
	// 6. 验证物理文件一键清空与删除机制
	// -------------------------------------------------------------------------
	dbPathA1 := filepath.Join(tempDir, "tenant_User_A_Session_1.db")
	if _, err := os.Stat(dbPathA1); os.IsNotExist(err) {
		t.Fatalf("expected SQLite file to exist at %s, but it does not", dbPathA1)
	}

	// 物理清空 User_A 的 Session_1
	t.Log("一键物理删除 User_A Session_1...")
	if err := manager.DeleteStore("User_A", "Session_1"); err != nil {
		t.Fatalf("failed to physically delete store: %v", err)
	}

	// 校验物理文件确实在磁盘上被擦除 (零残留!)
	if _, err := os.Stat(dbPathA1); !os.IsNotExist(err) {
		t.Errorf("expected SQLite file %s to be physically deleted, but it still exists on disk", dbPathA1)
	}
	t.Log("✔ 验证物理销毁成功：SQLite 库文件已物理清除，无任何隐私残留。")

	_ = ctx // 防止 unused warning
}
