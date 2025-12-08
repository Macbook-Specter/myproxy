package database

import (
	"os"
	"testing"

	"myproxy.com/p/internal/config"
)

func TestDatabase(t *testing.T) {
	// 使用临时数据库文件进行测试
	dbPath := "./test_myproxy.db"
	defer os.Remove(dbPath)

	// 初始化数据库
	if err := InitDB(dbPath); err != nil {
		t.Fatalf("初始化数据库失败: %v", err)
	}
	defer CloseDB()

	// 测试添加订阅
	sub, err := AddOrUpdateSubscription("https://example.com/sub1", "测试订阅1")
	if err != nil {
		t.Fatalf("添加订阅失败: %v", err)
	}
	if sub == nil {
		t.Fatal("订阅应该被创建")
	}
	if sub.Label != "测试订阅1" {
		t.Errorf("订阅标签不正确，期望: 测试订阅1, 实际: %s", sub.Label)
	}

	// 测试更新订阅标签
	sub, err = AddOrUpdateSubscription("https://example.com/sub1", "更新后的标签")
	if err != nil {
		t.Fatalf("更新订阅失败: %v", err)
	}
	if sub.Label != "更新后的标签" {
		t.Errorf("订阅标签更新失败，期望: 更新后的标签, 实际: %s", sub.Label)
	}

	// 测试添加服务器
	server := config.Server{
		ID:       "test-server-1",
		Name:     "测试服务器1",
		Addr:     "127.0.0.1",
		Port:     1080,
		Username: "user",
		Password: "pass",
		Delay:    100,
		Selected: false,
		Enabled:  true,
	}

	subscriptionID := sub.ID
	if err := AddOrUpdateServer(server, &subscriptionID); err != nil {
		t.Fatalf("添加服务器失败: %v", err)
	}

	// 测试获取服务器
	retrievedServer, err := GetServer(server.ID)
	if err != nil {
		t.Fatalf("获取服务器失败: %v", err)
	}
	if retrievedServer.Name != server.Name {
		t.Errorf("服务器名称不正确，期望: %s, 实际: %s", server.Name, retrievedServer.Name)
	}

	// 测试更新服务器延迟
	if err := UpdateServerDelay(server.ID, 200); err != nil {
		t.Fatalf("更新服务器延迟失败: %v", err)
	}

	updatedServer, err := GetServer(server.ID)
	if err != nil {
		t.Fatalf("获取更新后的服务器失败: %v", err)
	}
	if updatedServer.Delay != 200 {
		t.Errorf("服务器延迟更新失败，期望: 200, 实际: %d", updatedServer.Delay)
	}

	// 测试根据订阅ID获取服务器
	servers, err := GetServersBySubscriptionID(subscriptionID)
	if err != nil {
		t.Fatalf("根据订阅ID获取服务器失败: %v", err)
	}
	if len(servers) != 1 {
		t.Errorf("服务器数量不正确，期望: 1, 实际: %d", len(servers))
	}

	// 测试获取所有订阅
	subscriptions, err := GetAllSubscriptions()
	if err != nil {
		t.Fatalf("获取所有订阅失败: %v", err)
	}
	if len(subscriptions) != 1 {
		t.Errorf("订阅数量不正确，期望: 1, 实际: %d", len(subscriptions))
	}

	// 测试获取所有服务器
	allServers, err := GetAllServers()
	if err != nil {
		t.Fatalf("获取所有服务器失败: %v", err)
	}
	if len(allServers) != 1 {
		t.Errorf("服务器数量不正确，期望: 1, 实际: %d", len(allServers))
	}
}
