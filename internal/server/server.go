package server

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"time"

	"myproxy.com/p/internal/config"
	"myproxy.com/p/internal/database"
)

// ServerManager 服务器管理器
type ServerManager struct {
	config *config.Config
}

// NewServerManager 创建新的服务器管理器
func NewServerManager(config *config.Config) *ServerManager {
	return &ServerManager{
		config: config,
	}
}

// AddServer 添加服务器
func (sm *ServerManager) AddServer(server config.Server) error {
	// 先添加到内存配置
	if err := sm.config.AddServer(server); err != nil {
		return err
	}

	// 添加到数据库（subscription_id 为 nil，表示手动添加的服务器）
	if err := database.AddOrUpdateServer(server, nil); err != nil {
		return fmt.Errorf("添加服务器到数据库失败: %w", err)
	}

	return nil
}

// RemoveServer 删除服务器
func (sm *ServerManager) RemoveServer(id string) error {
	// 先从内存配置删除
	if err := sm.config.RemoveServer(id); err != nil {
		return err
	}

	// 从数据库删除
	if err := database.DeleteServer(id); err != nil {
		return fmt.Errorf("从数据库删除服务器失败: %w", err)
	}

	return nil
}

// GetServer 获取服务器
func (sm *ServerManager) GetServer(id string) (*config.Server, error) {
	return sm.config.GetServer(id)
}

// ListServers 获取所有服务器
func (sm *ServerManager) ListServers() []config.Server {
	return sm.config.Servers
}

// SelectServer 选择服务器
func (sm *ServerManager) SelectServer(id string) error {
	return sm.config.SelectServer(id)
}

// GetSelectedServer 获取当前选中的服务器
func (sm *ServerManager) GetSelectedServer() (*config.Server, error) {
	return sm.config.GetSelectedServer()
}

// UpdateServer 更新服务器信息
func (sm *ServerManager) UpdateServer(server config.Server) error {
	// 先更新内存配置
	for i, s := range sm.config.Servers {
		if s.ID == server.ID {
			sm.config.Servers[i] = server
			// 如果更新的是选中的服务器，确保选中状态正确
			if server.ID == sm.config.SelectedServerID {
				sm.config.Servers[i].Selected = true
			}

			// 更新数据库（保留原有的 subscription_id）
			// 使用 nil 作为 subscriptionID，AddOrUpdateServer 会自动保持原有的 subscription_id
			if err := database.AddOrUpdateServer(server, nil); err != nil {
				return fmt.Errorf("更新服务器到数据库失败: %w", err)
			}

			return nil
		}
	}

	return fmt.Errorf("服务器不存在: %s", server.ID)
}

// UpdateServerDelay 更新服务器延迟
func (sm *ServerManager) UpdateServerDelay(id string, delay int) error {
	// 先更新内存配置
	for i, s := range sm.config.Servers {
		if s.ID == id {
			sm.config.Servers[i].Delay = delay

			// 更新数据库
			if err := database.UpdateServerDelay(id, delay); err != nil {
				return fmt.Errorf("更新服务器延迟到数据库失败: %w", err)
			}

			return nil
		}
	}

	return fmt.Errorf("服务器不存在: %s", id)
}

// GenerateServerID 生成服务器唯一ID
func GenerateServerID(addr string, port int, username string) string {
	// 使用地址、端口和用户名生成唯一ID
	data := fmt.Sprintf("%s:%d:%s:%d", addr, port, username, time.Now().UnixNano())
	hash := md5.Sum([]byte(data))
	return hex.EncodeToString(hash[:])
}
