package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Server 定义单个服务器信息
type Server struct {
	ID       string `json:"id"`       // 服务器唯一标识
	Name     string `json:"name"`     // 服务器名称
	Addr     string `json:"addr"`     // 服务器地址
	Port     int    `json:"port"`     // 服务器端口
	Username string `json:"username"` // 认证用户名
	Password string `json:"password"` // 认证密码
	Delay    int    `json:"delay"`    // 延迟（毫秒）
	Selected bool   `json:"selected"` // 是否被选中
	Enabled  bool   `json:"enabled"`  // 是否启用
}

// ForwardRule 定义单个转发规则
type ForwardRule struct {
	ID          string `json:"id"`          // 规则唯一标识
	Enabled     bool   `json:"enabled"`     // 是否启用
	Protocol    string `json:"protocol"`    // 协议类型 ("tcp" 或 "udp")
	LocalAddr   string `json:"localAddr"`   // 本地监听地址
	RemoteAddr  string `json:"remoteAddr"`  // 远程目标地址
}

// Config 定义配置文件结构
type Config struct {
	ProxyAddr         string        `json:"proxyAddr"`         // SOCKS5代理服务器地址
	Username          string        `json:"username"`          // 认证用户名
	Password          string        `json:"password"`          // 认证密码
	ForwardRules      []ForwardRule `json:"forwardRules"`      // 转发规则列表
	SubscriptionURL   string        `json:"subscriptionURL"`   // 订阅URL
	Servers           []Server      `json:"servers"`           // 服务器列表
	SelectedServerID  string        `json:"selectedServerID"`  // 当前选中的服务器ID
	AutoProxyEnabled  bool          `json:"autoProxyEnabled"`  // 自动代理是否启用
	AutoProxyPort     int           `json:"autoProxyPort"`     // 自动代理监听端口
	AutoSetSystemProxy bool         `json:"autoSetSystemProxy"` // 是否自动设置系统代理
	LogLevel          string        `json:"logLevel"`          // 日志级别
	LogFile           string        `json:"logFile"`           // 日志文件路径
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		ProxyAddr:         "127.0.0.1:1080",
		AutoProxyEnabled:  false,
		AutoProxyPort:     1080,
		AutoSetSystemProxy: false,
		LogLevel:          "info",
		LogFile:           "myproxy.log",
		ForwardRules: []ForwardRule{
			{
				ID:          "default-tcp",
				Enabled:     true,
				Protocol:    "tcp",
				LocalAddr:   "127.0.0.1:8080",
				RemoteAddr:  "example.com:80",
			},
		},
		Servers: []Server{
			{
				ID:       "default-server",
				Name:     "Default Server",
				Addr:     "127.0.0.1",
				Port:     1080,
				Username: "",
				Password: "",
				Delay:    0,
				Selected: true,
				Enabled:  true,
			},
		},
		SelectedServerID: "default-server",
	}
}

// LoadConfig 从文件加载配置
func LoadConfig(filePath string) (*Config, error) {
	// 如果文件不存在，返回默认配置
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		defaultConfig := DefaultConfig()
		// 保存默认配置到文件
		if err := SaveConfig(defaultConfig, filePath); err != nil {
			return nil, fmt.Errorf("保存默认配置失败: %w", err)
		}
		return defaultConfig, nil
	}

	// 读取配置文件
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 解析配置
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 验证配置
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}

	return &config, nil
}

// SaveConfig 保存配置到文件
func SaveConfig(config *Config, filePath string) error {
	// 验证配置
	if err := config.Validate(); err != nil {
		return fmt.Errorf("配置验证失败: %w", err)
	}

	// 创建目录（如果不存在）
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}

	// 序列化配置
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	// 保存到文件
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	return nil
}

// Validate 验证配置有效性
func (c *Config) Validate() error {
	// 检查代理地址
	if c.ProxyAddr == "" {
		return fmt.Errorf("代理地址不能为空")
	}

	// 检查转发规则
	for i, rule := range c.ForwardRules {
		if rule.ID == "" {
			return fmt.Errorf("转发规则 %d 的ID不能为空", i)
		}

		if rule.Protocol != "tcp" && rule.Protocol != "udp" {
			return fmt.Errorf("转发规则 %s 的协议类型无效: %s", rule.ID, rule.Protocol)
		}

		if rule.LocalAddr == "" {
			return fmt.Errorf("转发规则 %s 的本地地址不能为空", rule.ID)
		}

		if rule.RemoteAddr == "" {
			return fmt.Errorf("转发规则 %s 的远程地址不能为空", rule.ID)
		}
	}

	// 检查服务器列表
	for i, server := range c.Servers {
		if server.ID == "" {
			return fmt.Errorf("服务器 %d 的ID不能为空", i)
		}

		if server.Addr == "" {
			return fmt.Errorf("服务器 %s 的地址不能为空", server.ID)
		}

		if server.Port <= 0 || server.Port > 65535 {
			return fmt.Errorf("服务器 %s 的端口无效: %d", server.ID, server.Port)
		}
	}

	return nil
}

// AddForwardRule 添加转发规则
func (c *Config) AddForwardRule(rule ForwardRule) error {
	// 检查ID是否已存在
	for _, r := range c.ForwardRules {
		if r.ID == rule.ID {
			return fmt.Errorf("转发规则ID已存在: %s", rule.ID)
		}
	}

	c.ForwardRules = append(c.ForwardRules, rule)
	return nil
}

// RemoveForwardRule 删除转发规则
func (c *Config) RemoveForwardRule(id string) error {
	for i, r := range c.ForwardRules {
		if r.ID == id {
			c.ForwardRules = append(c.ForwardRules[:i], c.ForwardRules[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("转发规则不存在: %s", id)
}

// GetForwardRule 获取转发规则
func (c *Config) GetForwardRule(id string) (*ForwardRule, error) {
	for i, r := range c.ForwardRules {
		if r.ID == id {
			return &c.ForwardRules[i], nil
		}
	}

	return nil, fmt.Errorf("转发规则不存在: %s", id)
}

// AddServer 添加服务器
func (c *Config) AddServer(server Server) error {
	// 检查ID是否已存在
	for _, s := range c.Servers {
		if s.ID == server.ID {
			return fmt.Errorf("服务器ID已存在: %s", server.ID)
		}
	}

	c.Servers = append(c.Servers, server)
	return nil
}

// RemoveServer 删除服务器
func (c *Config) RemoveServer(id string) error {
	for i, s := range c.Servers {
		if s.ID == id {
			c.Servers = append(c.Servers[:i], c.Servers[i+1:]...)
			// 如果删除的是选中的服务器，重置选中服务器
			if c.SelectedServerID == id {
				c.SelectedServerID = ""
			}
			return nil
		}
	}

	return fmt.Errorf("服务器不存在: %s", id)
}

// GetServer 获取服务器
func (c *Config) GetServer(id string) (*Server, error) {
	for i, s := range c.Servers {
		if s.ID == id {
			return &c.Servers[i], nil
		}
	}

	return nil, fmt.Errorf("服务器不存在: %s", id)
}

// SelectServer 选择服务器
func (c *Config) SelectServer(id string) error {
	// 检查服务器是否存在
	found := false
	for i, s := range c.Servers {
		if s.ID == id {
			found = true
			// 取消所有服务器的选中状态
			for j := range c.Servers {
				c.Servers[j].Selected = false
			}
			// 选中当前服务器
			c.Servers[i].Selected = true
			c.SelectedServerID = id
			break
		}
	}

	if !found {
		return fmt.Errorf("服务器不存在: %s", id)
	}

	return nil
}

// GetSelectedServer 获取当前选中的服务器
func (c *Config) GetSelectedServer() (*Server, error) {
	for _, s := range c.Servers {
		if s.ID == c.SelectedServerID {
			return &s, nil
		}
	}

	return nil, fmt.Errorf("没有选中的服务器")
}
