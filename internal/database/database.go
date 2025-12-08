package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"myproxy.com/p/internal/config"
)

// DB 数据库连接
var DB *sql.DB

// InitDB 初始化数据库
func InitDB(dbPath string) error {
	// 创建目录（如果不存在）
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return fmt.Errorf("创建数据库目录失败: %w", err)
	}

	// 打开数据库连接
	var err error
	DB, err = sql.Open("sqlite3", dbPath+"?_foreign_keys=1")
	if err != nil {
		return fmt.Errorf("打开数据库失败: %w", err)
	}

	// 测试连接
	if err := DB.Ping(); err != nil {
		return fmt.Errorf("数据库连接测试失败: %w", err)
	}

	// 创建表
	if err := createTables(); err != nil {
		return fmt.Errorf("创建表失败: %w", err)
	}

	return nil
}

// createTables 创建数据库表
func createTables() error {
	// 创建订阅表
	createSubscriptionsTable := `
	CREATE TABLE IF NOT EXISTS subscriptions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		url TEXT NOT NULL UNIQUE,
		label TEXT NOT NULL DEFAULT '',
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);`

	// 创建服务器表
	createServersTable := `
	CREATE TABLE IF NOT EXISTS servers (
		id TEXT PRIMARY KEY,
		subscription_id INTEGER,
		name TEXT NOT NULL,
		addr TEXT NOT NULL,
		port INTEGER NOT NULL,
		username TEXT NOT NULL DEFAULT '',
		password TEXT NOT NULL DEFAULT '',
		delay INTEGER NOT NULL DEFAULT 0,
		selected INTEGER NOT NULL DEFAULT 0,
		enabled INTEGER NOT NULL DEFAULT 1,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (subscription_id) REFERENCES subscriptions(id) ON DELETE SET NULL
	);`

	// 创建布局配置表
	createLayoutConfigTable := `
	CREATE TABLE IF NOT EXISTS layout_config (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		key TEXT NOT NULL UNIQUE,
		value TEXT NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);`

	// 创建索引
	createIndexes := `
	CREATE INDEX IF NOT EXISTS idx_servers_subscription_id ON servers(subscription_id);
	CREATE INDEX IF NOT EXISTS idx_servers_enabled ON servers(enabled);
	CREATE INDEX IF NOT EXISTS idx_subscriptions_url ON subscriptions(url);
	CREATE INDEX IF NOT EXISTS idx_layout_config_key ON layout_config(key);
	`

	if _, err := DB.Exec(createSubscriptionsTable); err != nil {
		return fmt.Errorf("创建订阅表失败: %w", err)
	}

	if _, err := DB.Exec(createServersTable); err != nil {
		return fmt.Errorf("创建服务器表失败: %w", err)
	}

	if _, err := DB.Exec(createLayoutConfigTable); err != nil {
		return fmt.Errorf("创建布局配置表失败: %w", err)
	}

	if _, err := DB.Exec(createIndexes); err != nil {
		return fmt.Errorf("创建索引失败: %w", err)
	}

	return nil
}

// CloseDB 关闭数据库连接
func CloseDB() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

// Subscription 订阅信息
type Subscription struct {
	ID        int64     `json:"id"`
	URL       string    `json:"url"`
	Label     string    `json:"label"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AddOrUpdateSubscription 添加或更新订阅
func AddOrUpdateSubscription(url, label string) (*Subscription, error) {
	now := time.Now()

	// 先尝试查询是否存在
	var sub Subscription
	err := DB.QueryRow("SELECT id, url, label, created_at, updated_at FROM subscriptions WHERE url = ?", url).
		Scan(&sub.ID, &sub.URL, &sub.Label, &sub.CreatedAt, &sub.UpdatedAt)

	if err == sql.ErrNoRows {
		// 不存在，插入新记录
		result, err := DB.Exec(
			"INSERT INTO subscriptions (url, label, created_at, updated_at) VALUES (?, ?, ?, ?)",
			url, label, now, now,
		)
		if err != nil {
			return nil, fmt.Errorf("插入订阅失败: %w", err)
		}

		id, err := result.LastInsertId()
		if err != nil {
			return nil, fmt.Errorf("获取插入ID失败: %w", err)
		}

		sub.ID = id
		sub.URL = url
		sub.Label = label
		sub.CreatedAt = now
		sub.UpdatedAt = now
	} else if err != nil {
		return nil, fmt.Errorf("查询订阅失败: %w", err)
	} else {
		// 存在，更新记录
		if label != sub.Label {
			_, err = DB.Exec(
				"UPDATE subscriptions SET label = ?, updated_at = ? WHERE id = ?",
				label, now, sub.ID,
			)
			if err != nil {
				return nil, fmt.Errorf("更新订阅失败: %w", err)
			}
			sub.Label = label
			sub.UpdatedAt = now
		}
	}

	return &sub, nil
}

// GetSubscriptionByURL 根据URL获取订阅
func GetSubscriptionByURL(url string) (*Subscription, error) {
	var sub Subscription
	err := DB.QueryRow(
		"SELECT id, url, label, created_at, updated_at FROM subscriptions WHERE url = ?",
		url,
	).Scan(&sub.ID, &sub.URL, &sub.Label, &sub.CreatedAt, &sub.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询订阅失败: %w", err)
	}

	return &sub, nil
}

// GetAllSubscriptions 获取所有订阅
func GetAllSubscriptions() ([]*Subscription, error) {
	rows, err := DB.Query("SELECT id, url, label, created_at, updated_at FROM subscriptions ORDER BY created_at DESC")
	if err != nil {
		return nil, fmt.Errorf("查询订阅列表失败: %w", err)
	}
	defer rows.Close()

	var subscriptions []*Subscription
	for rows.Next() {
		var sub Subscription
		if err := rows.Scan(&sub.ID, &sub.URL, &sub.Label, &sub.CreatedAt, &sub.UpdatedAt); err != nil {
			return nil, fmt.Errorf("扫描订阅数据失败: %w", err)
		}
		subscriptions = append(subscriptions, &sub)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历订阅数据失败: %w", err)
	}

	return subscriptions, nil
}

// DeleteSubscription 删除订阅
func DeleteSubscription(url string) error {
	_, err := DB.Exec("DELETE FROM subscriptions WHERE url = ?", url)
	if err != nil {
		return fmt.Errorf("删除订阅失败: %w", err)
	}
	return nil
}

// AddOrUpdateServer 添加或更新服务器
// 如果 subscriptionID 为 nil 且服务器已存在，则保持原有的 subscription_id
func AddOrUpdateServer(server config.Server, subscriptionID *int64) error {
	now := time.Now()

	// 检查服务器是否存在
	var existingID string
	var existingSubscriptionID sql.NullInt64
	err := DB.QueryRow("SELECT id, subscription_id FROM servers WHERE id = ?", server.ID).
		Scan(&existingID, &existingSubscriptionID)

	if err == sql.ErrNoRows {
		// 不存在，插入新记录
		_, err = DB.Exec(
			`INSERT INTO servers (id, subscription_id, name, addr, port, username, password, delay, selected, enabled, created_at, updated_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			server.ID, subscriptionID, server.Name, server.Addr, server.Port,
			server.Username, server.Password, server.Delay,
			boolToInt(server.Selected), boolToInt(server.Enabled),
			now, now,
		)
		if err != nil {
			return fmt.Errorf("插入服务器失败: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("查询服务器失败: %w", err)
	} else {
		// 存在，更新记录
		// 如果 subscriptionID 为 nil，保持原有的 subscription_id
		updateSubscriptionID := subscriptionID
		if updateSubscriptionID == nil && existingSubscriptionID.Valid {
			updateSubscriptionID = &existingSubscriptionID.Int64
		}

		_, err = DB.Exec(
			`UPDATE servers SET 
				subscription_id = ?, name = ?, addr = ?, port = ?, username = ?, password = ?,
				delay = ?, selected = ?, enabled = ?, updated_at = ?
			 WHERE id = ?`,
			updateSubscriptionID, server.Name, server.Addr, server.Port,
			server.Username, server.Password, server.Delay,
			boolToInt(server.Selected), boolToInt(server.Enabled),
			now, server.ID,
		)
		if err != nil {
			return fmt.Errorf("更新服务器失败: %w", err)
		}
	}

	return nil
}

// GetServerSubscriptionID 获取服务器的订阅ID
func GetServerSubscriptionID(serverID string) (*int64, error) {
	var subscriptionID sql.NullInt64
	err := DB.QueryRow("SELECT subscription_id FROM servers WHERE id = ?", serverID).
		Scan(&subscriptionID)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询服务器订阅ID失败: %w", err)
	}

	if subscriptionID.Valid {
		return &subscriptionID.Int64, nil
	}
	return nil, nil
}

// GetServer 获取服务器
func GetServer(id string) (*config.Server, error) {
	var server config.Server
	var selected, enabled int

	err := DB.QueryRow(
		`SELECT id, name, addr, port, username, password, delay, selected, enabled
		 FROM servers WHERE id = ?`,
		id,
	).Scan(&server.ID, &server.Name, &server.Addr, &server.Port,
		&server.Username, &server.Password, &server.Delay,
		&selected, &enabled)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("服务器不存在: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("查询服务器失败: %w", err)
	}

	server.Selected = intToBool(selected)
	server.Enabled = intToBool(enabled)

	return &server, nil
}

// GetAllServers 获取所有服务器
func GetAllServers() ([]config.Server, error) {
	rows, err := DB.Query(
		`SELECT id, name, addr, port, username, password, delay, selected, enabled
		 FROM servers ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("查询服务器列表失败: %w", err)
	}
	defer rows.Close()

	var servers []config.Server
	for rows.Next() {
		var server config.Server
		var selected, enabled int

		if err := rows.Scan(&server.ID, &server.Name, &server.Addr, &server.Port,
			&server.Username, &server.Password, &server.Delay,
			&selected, &enabled); err != nil {
			return nil, fmt.Errorf("扫描服务器数据失败: %w", err)
		}

		server.Selected = intToBool(selected)
		server.Enabled = intToBool(enabled)
		servers = append(servers, server)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历服务器数据失败: %w", err)
	}

	return servers, nil
}

// GetServersBySubscriptionID 根据订阅ID获取服务器列表
func GetServersBySubscriptionID(subscriptionID int64) ([]config.Server, error) {
	rows, err := DB.Query(
		`SELECT id, name, addr, port, username, password, delay, selected, enabled
		 FROM servers WHERE subscription_id = ? ORDER BY created_at DESC`,
		subscriptionID,
	)
	if err != nil {
		return nil, fmt.Errorf("查询服务器列表失败: %w", err)
	}
	defer rows.Close()

	var servers []config.Server
	for rows.Next() {
		var server config.Server
		var selected, enabled int

		if err := rows.Scan(&server.ID, &server.Name, &server.Addr, &server.Port,
			&server.Username, &server.Password, &server.Delay,
			&selected, &enabled); err != nil {
			return nil, fmt.Errorf("扫描服务器数据失败: %w", err)
		}

		server.Selected = intToBool(selected)
		server.Enabled = intToBool(enabled)
		servers = append(servers, server)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历服务器数据失败: %w", err)
	}

	return servers, nil
}

// UpdateServerDelay 更新服务器延迟
func UpdateServerDelay(id string, delay int) error {
	_, err := DB.Exec(
		"UPDATE servers SET delay = ?, updated_at = ? WHERE id = ?",
		delay, time.Now(), id,
	)
	if err != nil {
		return fmt.Errorf("更新服务器延迟失败: %w", err)
	}
	return nil
}

// DeleteServer 删除服务器
func DeleteServer(id string) error {
	_, err := DB.Exec("DELETE FROM servers WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("删除服务器失败: %w", err)
	}
	return nil
}

// DeleteServersBySubscriptionID 删除指定订阅的所有服务器
func DeleteServersBySubscriptionID(subscriptionID int64) error {
	_, err := DB.Exec("DELETE FROM servers WHERE subscription_id = ?", subscriptionID)
	if err != nil {
		return fmt.Errorf("删除订阅服务器失败: %w", err)
	}
	return nil
}

// SetLayoutConfig 设置布局配置
func SetLayoutConfig(key, value string) error {
	now := time.Now()
	_, err := DB.Exec(
		`INSERT INTO layout_config (key, value, created_at, updated_at)
		 VALUES (?, ?, ?, ?)
		 ON CONFLICT(key) DO UPDATE SET value = ?, updated_at = ?`,
		key, value, now, now, value, now,
	)
	if err != nil {
		return fmt.Errorf("设置布局配置失败: %w", err)
	}
	return nil
}

// GetLayoutConfig 获取布局配置
func GetLayoutConfig(key string) (string, error) {
	var value string
	err := DB.QueryRow("SELECT value FROM layout_config WHERE key = ?", key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("获取布局配置失败: %w", err)
	}
	return value, nil
}

// GetLayoutConfigWithDefault 获取布局配置，如果不存在则返回默认值
func GetLayoutConfigWithDefault(key, defaultValue string) (string, error) {
	value, err := GetLayoutConfig(key)
	if err != nil {
		return "", err
	}
	if value == "" {
		// 如果不存在，写入默认值
		if err := SetLayoutConfig(key, defaultValue); err != nil {
			return "", err
		}
		return defaultValue, nil
	}
	return value, nil
}

// boolToInt 将布尔值转换为整数
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// intToBool 将整数转换为布尔值
func intToBool(i int) bool {
	return i != 0
}
