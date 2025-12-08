# SOCKS5 代理客户端工具

一个功能完整的 SOCKS5 代理客户端桌面应用，支持订阅管理、服务器管理、自动代理和图形界面。

> **重要提示**：本项目仅提供 **GUI 桌面应用**作为正式入口。`cmd/` 目录下的其他程序（`client`、`myproxy`、`server`）仅用于开发和测试，不作为项目正式入口。

## 功能特性

### 核心功能
- ✅ **TCP 端口转发**：将本地 TCP 端口流量通过 SOCKS5 代理转发到远程服务器
- ✅ **UDP 端口转发**：支持 UDP 数据的转发，实现 SOCKS5 UDP ASSOCIATE 功能
- ✅ **多规则支持**：可以同时配置多个转发规则
- ✅ **认证支持**：支持 SOCKS5 用户名/密码认证
- ✅ **自动代理模式**：作为本地 SOCKS5 服务器，自动选择配置的代理服务器进行转发

### 服务器管理
- ✅ **多服务器支持**：管理多个 SOCKS5 代理服务器
- ✅ **订阅管理**：支持从订阅 URL 获取服务器列表
  - 支持 VMess 格式订阅
  - 支持 SOCKS5 格式订阅
  - 支持 JSON 格式订阅
  - 支持 Base64 编码的订阅
  - 支持为每个订阅添加标签
- ✅ **延迟测试**：支持单个或批量测试服务器延迟
- ✅ **服务器选择**：支持设置默认服务器
- ✅ **数据库存储**：使用 SQLite 数据库持久化存储订阅和服务器信息
  - 自动保存从 URL 解析的服务器数据
  - 更新操作同步更新数据库
  - 订阅地址与标签的对应关系存储在数据库中

### 配置管理
- ✅ **配置文件**：从 JSON 配置文件加载代理信息和转发规则
- ✅ **命令行参数**：支持通过命令行参数覆盖配置文件设置
- ✅ **动态规则**：支持在配置文件中定义多个启用/禁用状态的规则

### 界面支持
- ✅ **桌面应用（正式入口）**：基于 Fyne 框架的 GUI 界面，可打包为桌面应用
  - 服务器列表管理（查看、添加、编辑、删除、启用/禁用）
  - 订阅管理（支持多个订阅，每个订阅可设置标签）
  - 延迟测试（单个或批量测试服务器延迟）
  - 自动代理控制（启动/停止本地 SOCKS5 代理服务器）
  - 实时日志查看（支持按级别和类型过滤）
  - 状态信息显示（代理状态、端口、选中服务器）
  - 布局配置持久化（窗口布局设置自动保存到数据库）

> **注意**：`cmd/client`、`cmd/myproxy`、`cmd/server` 仅用于开发和测试，不作为项目正式入口。

### 开发工具（仅用于测试）
- ⚠️ **SOCKS5 服务器** (`cmd/server`)：内置 SOCKS5 服务器实现，仅用于开发和测试
- ⚠️ **命令行客户端** (`cmd/client`)：轻量级命令行客户端，仅用于开发和测试
- ⚠️ **测试程序** (`cmd/myproxy`)：提供 TCP 和 UDP 功能测试示例，仅用于开发和测试

> **重要**：以上工具仅供开发和测试使用，不作为项目正式入口。正式使用请使用 GUI 桌面应用。

## 架构设计

### 代码结构
```
├── cmd/
│   ├── gui/             # 🎯 桌面应用入口（唯一正式入口）
│   ├── client/          # ⚠️ 命令行客户端（仅用于测试）
│   ├── myproxy/         # ⚠️ 测试程序（仅用于测试）
│   └── server/          # ⚠️ SOCKS5 服务器（仅用于测试）
├── doc/                 # 文档目录
├── internal/            # 内部包
│   ├── config/          # 配置管理
│   ├── database/        # 数据库管理（SQLite）
│   ├── logging/         # 日志管理（支持日志归档）
│   ├── ping/            # 延迟测试
│   ├── proxy/           # 端口转发核心逻辑（自动代理）
│   ├── server/          # 服务器管理
│   ├── socks5/          # SOCKS5 协议实现（客户端）
│   ├── subscription/   # 订阅管理（支持多种格式）
│   ├── ui/              # GUI 界面组件
│   │   ├── app.go       # 应用状态管理
│   │   ├── mainwindow.go # 主窗口
│   │   ├── subscription.go # 订阅管理面板
│   │   ├── serverlist.go   # 服务器列表面板
│   │   ├── status.go       # 状态信息面板
│   │   ├── logs.go         # 日志显示面板
│   │   └── customborder.go # 自定义布局
│   └── utils/           # 工具函数
├── ui/                  # 桌面应用 UI 资源
├── config.json          # 默认配置文件
└── README.md            # 项目说明文档
```

### 协议流程

```mermaid
sequenceDiagram
    participant User as 用户
    participant Client as SOCKS5 客户端
    participant Proxy as SOCKS5 代理服务器
    participant Target as 目标服务器

    User->>Client: 访问本地转发端口
    Client->>Proxy: 建立 TCP 连接
    
    rect rgb(200, 255, 200)
        Client->>Proxy: 协商 SOCKS5 版本及认证方法
        Proxy->>Client: 回复选择的认证方法
    end
    
    opt 认证阶段
        Client->>Proxy: 发送用户名/密码
        Proxy->>Client: 回复认证状态
    end
    
    rect rgb(255, 255, 200)
        Client->>Proxy: 发送 CONNECT 请求
        Proxy->>Target: 连接目标服务器
        Proxy->>Client: 回复连接成功
    end
    
    loop 数据转发
        User->>Client: 发送数据
        Client->>Proxy: 转发数据
        Proxy->>Target: 转发数据
        
        Target->>Proxy: 回复数据
        Proxy->>Client: 转发数据
        Client->>User: 回复数据
    end
```

## 快速开始

### 前提条件
- Go 1.25.4 或更高版本
- SOCKS5 代理服务器（用于代理转发）

### 编译和运行

#### 🎯 编译桌面应用（正式入口）

```bash
# 编译桌面应用
go build -o gui ./cmd/gui

# 或直接运行
go run ./cmd/gui/main.go
```

#### ⚠️ 开发工具（仅用于测试）

以下工具仅供开发和测试使用，不作为项目正式入口：

```bash
# 编译命令行客户端（测试用）
go build -o client ./cmd/client

# 编译 SOCKS5 服务器（测试用）
go build -o server ./cmd/server

# 编译测试程序（测试用）
go build -o myproxy ./cmd/myproxy
```

### 启动桌面应用

```bash
# 方式1：直接运行源码
go run ./cmd/gui/main.go

# 方式2：使用编译后的二进制文件
./gui

# 方式3：指定配置文件路径
go run ./cmd/gui/main.go /path/to/config.json
```

应用启动后会自动：
- 初始化 SQLite 数据库（`data/myproxy.db`）
- 加载配置文件（默认 `./config.json`）
- 归档旧日志文件（如果存在）
- 加载保存的窗口布局配置

## 配置文件格式

### 基本配置

```json
{
  "proxyAddr": "127.0.0.1:1080",
  "username": "",
  "password": "",
  "forwardRules": [
    {
      "id": "default-tcp",
      "enabled": true,
      "protocol": "tcp",
      "localAddr": "127.0.0.1:8080",
      "remoteAddr": "example.com:80"
    },
    {
      "id": "dns-udp",
      "enabled": false,
      "protocol": "udp",
      "localAddr": "127.0.0.1:5353",
      "remoteAddr": "8.8.8.8:53"
    }
  ],
  "subscriptionURL": "https://example.com/subscription",
  "servers": [
    {
      "id": "server-1",
      "name": "服务器1",
      "addr": "proxy.example.com",
      "port": 1080,
      "username": "user",
      "password": "pass",
      "delay": 0,
      "selected": false,
      "enabled": true
    }
  ],
  "selectedServerID": "",
  "autoProxyEnabled": false,
  "autoProxyPort": 12800,
  "autoSetSystemProxy": true,
  "logLevel": "info",
  "logFile": "myproxy.log"
}
```

### 配置字段说明

#### 基本配置
| 字段名 | 类型 | 描述 |
|-------|------|------|
| `proxyAddr` | string | SOCKS5 代理服务器地址 |
| `username` | string | 认证用户名（可选） |
| `password` | string | 认证密码（可选） |
| `subscriptionURL` | string | 订阅 URL（可选） |
| `logLevel` | string | 日志级别（debug/info/warn/error） |
| `logFile` | string | 日志文件路径 |

#### 转发规则字段
| 字段名 | 类型 | 描述 |
|-------|------|------|
| `id` | string | 规则唯一标识 |
| `enabled` | boolean | 是否启用该规则 |
| `protocol` | string | 协议类型（"tcp" 或 "udp"） |
| `localAddr` | string | 本地监听地址 |
| `remoteAddr` | string | 远程目标地址 |

#### 服务器字段
| 字段名 | 类型 | 描述 |
|-------|------|------|
| `id` | string | 服务器唯一标识 |
| `name` | string | 服务器名称 |
| `addr` | string | 服务器地址 |
| `port` | int | 服务器端口 |
| `username` | string | 认证用户名 |
| `password` | string | 认证密码 |
| `delay` | int | 延迟（毫秒） |
| `selected` | boolean | 是否被选中 |
| `enabled` | boolean | 是否启用 |

#### 自动代理配置
| 字段名 | 类型 | 描述 |
|-------|------|------|
| `autoProxyEnabled` | boolean | 是否启用自动代理 |
| `autoProxyPort` | int | 自动代理监听端口 |
| `autoSetSystemProxy` | boolean | 是否自动设置系统代理 |
| `selectedServerID` | string | 当前选中的服务器 ID |

## 命令行参数

```
Usage of client:
  -config string
        配置文件路径 (默认 "./config.json")
  -local string
        本地监听地址 (覆盖配置文件)
  -password string
        认证密码 (覆盖配置文件)
  -proto string
        协议类型 (tcp/udp) (覆盖配置文件) (默认 "tcp")
  -proxy string
        SOCKS5代理服务器地址 (覆盖配置文件)
  -remote string
        远程目标地址 (覆盖配置文件)
  -username string
        认证用户名 (覆盖配置文件)
```

## 订阅格式支持

项目支持多种订阅格式：

1. **VMess 格式**：`vmess://base64(json)`
2. **SOCKS5 格式**：`socks5://username:password@addr:port`
3. **JSON 格式**：标准 JSON 数组格式
4. **Base64 编码**：支持 Base64 编码的订阅内容

订阅管理器会自动识别并解析这些格式。

## 示例用法

### 示例 1: HTTP 代理转发

```bash
# 启动本地 8080 端口转发到 example.com:80
./client -proxy 127.0.0.1:1080 -local 127.0.0.1:8080 -remote example.com:80

# 测试访问
curl http://localhost:8080
```

### 示例 2: DNS 代理转发

```bash
# 启动本地 5353 端口 UDP 转发到 8.8.8.8:53
./client -proxy 127.0.0.1:1080 -local 127.0.0.1:5353 -remote 8.8.8.8:53 -proto udp

# 测试 DNS 查询
dig @127.0.0.1 -p 5353 example.com
```

### 示例 3: 从配置文件启动多个规则

1. 创建配置文件 `config.json`
2. 定义多个转发规则
3. 启动客户端

```bash
./client -config ./config.json
```

### 示例 4: 使用自动代理模式

1. 在 GUI 中配置多个服务器
2. 选择一个服务器作为默认服务器
3. 启动自动代理
4. 配置系统代理指向 `127.0.0.1:12800`（或配置的端口）

## 桌面应用（正式入口）

### 功能特性

#### 1. 订阅管理（顶部 20% 区域）
- ✅ 支持多个订阅地址
- ✅ 为每个订阅设置标签
- ✅ 点击标签编辑订阅 URL 和标签
- ✅ 自动从订阅 URL 获取服务器列表
- ✅ 订阅数据持久化存储到数据库
- ✅ 使用双向数据绑定动态显示标签

#### 2. 服务器列表（中间 50% 区域）
- ✅ 显示所有服务器（来自订阅或手动添加）
- ✅ 显示服务器名称、地址、端口、延迟
- ✅ 右键菜单操作：
  - 启动代理（使用选中服务器）
  - 停止代理
  - 测试延迟
  - 编辑服务器
  - 删除服务器
- ✅ 服务器数据持久化存储到数据库

#### 3. 日志显示（中间 25% 区域）
- ✅ 实时显示应用日志和代理日志
- ✅ 支持按日志级别过滤（全部/DEBUG/INFO/WARN/ERROR/FATAL）
- ✅ 支持按日志类型过滤（全部/app/proxy）
- ✅ 等宽字体显示，提高可读性
- ✅ 自动滚动到最新日志
- ✅ 日志文件自动归档（启动时归档旧日志）

#### 4. 状态信息（底部 5% 区域）
- ✅ 显示代理状态（运行中/代理未启动）
- ✅ 显示代理端口
- ✅ 显示当前选中的服务器
- ✅ 使用双向数据绑定自动更新
- ✅ 水平布局，避免内容过长

#### 5. 布局管理
- ✅ 窗口布局配置持久化（保存到数据库）
- ✅ 启动时自动加载布局配置
- ✅ 支持自定义分割比例
- ✅ 默认布局：订阅 20%，服务器列表 50%，日志 25%，状态 5%

### 运行桌面应用

```bash
# 直接运行
go run ./cmd/gui/main.go

# 或使用编译后的二进制文件
./gui

# 指定配置文件路径
go run ./cmd/gui/main.go /path/to/config.json
```

### 界面特点

- 🎨 简洁的现代化设计
- 📱 基于 Fyne 框架，跨平台支持（Windows/macOS/Linux）
- 📐 响应式布局，适配不同屏幕尺寸
- 💾 布局配置自动保存和恢复
- 🔄 双向数据绑定，UI 自动更新

## 开发指南

### 项目入口

**唯一正式入口**：`cmd/gui/main.go`

其他 `cmd/` 目录下的程序仅供开发和测试使用：
- `cmd/client/` - 命令行客户端（测试用）
- `cmd/myproxy/` - SOCKS5 功能测试程序（测试用）
- `cmd/server/` - SOCKS5 服务器实现（测试用）

### 目录说明

#### 核心模块 (`internal/`)
- **`internal/socks5/`**: SOCKS5 协议实现，包含客户端逻辑
- **`internal/proxy/`**: 端口转发核心逻辑，处理 TCP 和 UDP 转发（自动代理）
- **`internal/config/`**: 配置管理，处理配置文件的加载和保存
- **`internal/database/`**: 数据库管理，使用 SQLite 存储订阅和服务器信息
- **`internal/server/`**: 服务器管理，处理多服务器逻辑
- **`internal/subscription/`**: 订阅管理，支持多种订阅格式解析
- **`internal/ping/`**: 延迟测试功能
- **`internal/logging/`**: 日志管理（支持日志归档）

#### GUI 模块 (`internal/ui/`)
- **`internal/ui/app.go`**: 应用状态管理，数据绑定
- **`internal/ui/mainwindow.go`**: 主窗口，布局管理
- **`internal/ui/subscription.go`**: 订阅管理面板
- **`internal/ui/serverlist.go`**: 服务器列表面板
- **`internal/ui/status.go`**: 状态信息面板
- **`internal/ui/logs.go`**: 日志显示面板
- **`internal/ui/customborder.go`**: 自定义布局实现

#### 入口程序 (`cmd/`)
- **`cmd/gui/`**: 🎯 桌面应用入口（唯一正式入口）
- **`cmd/client/`**: ⚠️ 命令行客户端（仅用于测试）
- **`cmd/myproxy/`**: ⚠️ 测试程序（仅用于测试）
- **`cmd/server/`**: ⚠️ SOCKS5 服务器（仅用于测试）

## 后续升级功能建议

### 🔥 高优先级功能

#### 1. 代理规则和路由
- **规则路由**：根据目标地址/域名选择不同的代理服务器
  - 支持域名匹配（如 `*.google.com`）
  - 支持 IP 段匹配（CIDR）
  - 支持正则表达式匹配
  - 规则优先级管理
- **分流规则**：直连/代理/阻止规则
- **规则组**：将多个规则组织成组，便于管理

#### 2. 流量统计和监控
- **实时流量统计**：显示当前连接数、上传/下载速度
- **历史流量统计**：按日/周/月统计流量使用情况
- **服务器流量统计**：统计每个服务器的流量使用
- **流量图表**：可视化显示流量趋势

#### 3. 系统代理自动配置
- **自动设置系统代理**：启动时自动配置系统代理设置
- **PAC 脚本生成**：自动生成 PAC 脚本用于浏览器代理
- **代理模式切换**：全局代理/规则代理/直连模式

#### 4. 服务器健康检查
- **自动健康检查**：定期测试服务器可用性
- **故障自动切换**：服务器不可用时自动切换到备用服务器
- **服务器评分**：根据延迟、成功率等指标评分

### 🚀 中优先级功能

#### 5. 多协议支持
- **HTTP/HTTPS 代理**：支持 HTTP CONNECT 方法
- **SOCKS4/SOCKS4a**：支持旧版 SOCKS 协议
- **Shadowsocks**：支持 Shadowsocks 协议
- **VMess**：支持 VMess 协议（已有订阅解析，可扩展）

#### 6. 代理链（Chain Proxy）
- **多级代理**：支持代理链配置
- **负载均衡**：多个代理服务器负载均衡
- **故障转移**：代理链中服务器故障时自动切换

#### 7. 安全和加密
- **TLS 加密**：支持 SOCKS5 over TLS
- **本地认证**：为本地 SOCKS5 服务器添加认证
- **访问控制**：IP 白名单/黑名单

#### 8. 用户体验优化
- **系统托盘**：最小化到系统托盘
- **快捷键支持**：全局快捷键快速切换代理
- **通知提醒**：代理状态变化时系统通知
- **主题切换**：支持深色/浅色主题
- **多语言支持**：国际化（i18n）

### 💡 低优先级功能

#### 9. 高级功能
- **SOCKS5 BIND 命令**：支持反向连接
- **UDP 转发优化**：改进 UDP 转发性能
- **IPv6 支持**：完整的 IPv6 支持
- **DNS 解析优化**：自定义 DNS 服务器

#### 10. 开发和运维
- **跨平台打包脚本**：自动打包为各平台安装包
- **自动更新**：应用自动更新功能
- **配置导入/导出**：配置文件导入导出
- **日志分析**：日志文件分析和可视化

#### 11. 高级订阅功能
- **订阅自动更新**：定时自动更新订阅
- **订阅合并**：多个订阅合并显示
- **订阅过滤**：根据规则过滤订阅中的服务器
- **订阅加密**：支持加密的订阅链接

### 📦 技术改进建议

1. **性能优化**
   - 连接池管理
   - 内存使用优化
   - 并发处理优化

2. **代码质量**
   - 增加单元测试覆盖率
   - 添加集成测试
   - 代码文档完善

3. **可维护性**
   - 模块化重构
   - 插件系统设计
   - 配置管理优化

## 测试

### 单元测试

```bash
go test ./...
```

### GUI 应用测试

1. 启动 GUI 应用
   ```bash
   go run ./cmd/gui/main.go
   ```

2. 测试功能：
   - 添加订阅并获取服务器列表
   - 测试服务器延迟
   - 启动/停止自动代理
   - 查看日志显示
   - 调整窗口布局并重启应用验证布局保存

### 开发工具测试（仅用于开发）

以下工具仅用于开发和测试 SOCKS5 协议功能：

```bash
# 1. 启动测试 SOCKS5 服务器
./server

# 2. 在另一个终端运行测试程序
./myproxy

# 3. 或使用命令行客户端测试
./client -proxy 127.0.0.1:1080 -local 127.0.0.1:8080 -remote example.com:80
```

## 数据库功能

项目使用 SQLite 数据库持久化存储订阅和服务器信息。

### 数据库初始化

在应用启动时，需要初始化数据库：

```go
import "myproxy.com/p/internal/database"

func main() {
    // 初始化数据库（数据库文件路径，例如："./data/myproxy.db"）
    if err := database.InitDB("./data/myproxy.db"); err != nil {
        log.Fatalf("初始化数据库失败: %v", err)
    }
    defer database.CloseDB()
    
    // 其他初始化代码...
}
```

### 数据库表结构

- **subscriptions 表**：存储订阅 URL 和标签
  - `id`: 主键
  - `url`: 订阅 URL（唯一）
  - `label`: 订阅标签
  - `created_at`: 创建时间
  - `updated_at`: 更新时间

- **servers 表**：存储服务器信息
  - `id`: 服务器唯一标识（主键）
  - `subscription_id`: 关联的订阅 ID（外键，可为空）
  - `name`: 服务器名称
  - `addr`: 服务器地址
  - `port`: 服务器端口
  - `username`: 认证用户名
  - `password`: 认证密码
  - `delay`: 延迟（毫秒）
  - `selected`: 是否被选中
  - `enabled`: 是否启用
  - `created_at`: 创建时间
  - `updated_at`: 更新时间

### 使用订阅管理器

订阅管理器会自动将数据保存到数据库：

```go
import "myproxy.com/p/internal/subscription"

sm := subscription.NewSubscriptionManager(serverManager)

// 获取订阅（会自动保存到数据库，可指定标签）
servers, err := sm.FetchSubscription("https://example.com/subscription", "我的订阅标签")

// 更新订阅（会自动更新数据库，可更新标签）
err = sm.UpdateSubscription("https://example.com/subscription", "更新后的标签")
```

### 直接使用数据库 API

```go
import "myproxy.com/p/internal/database"

// 添加或更新订阅
sub, err := database.AddOrUpdateSubscription("https://example.com/sub", "标签")

// 获取所有订阅
subscriptions, err := database.GetAllSubscriptions()

// 添加或更新服务器
var subscriptionID *int64 = &sub.ID
err = database.AddOrUpdateServer(server, subscriptionID)

// 获取所有服务器
servers, err := database.GetAllServers()

// 根据订阅ID获取服务器
servers, err := database.GetServersBySubscriptionID(sub.ID)
```

## 依赖项

主要依赖：
- `fyne.io/fyne/v2`: GUI 框架（桌面应用必需）
- `github.com/mattn/go-sqlite3`: SQLite 数据库驱动（数据持久化）

## 项目状态

### ✅ 已完成功能
- [x] GUI 桌面应用（唯一正式入口）
- [x] 订阅管理（多订阅、标签、多种格式支持）
- [x] 服务器管理（增删改查、启用/禁用）
- [x] 延迟测试（单个/批量）
- [x] 自动代理（启动/停止本地 SOCKS5 服务器）
- [x] 日志查看（过滤、归档）
- [x] 状态显示（双向数据绑定）
- [x] 布局配置持久化
- [x] 数据库存储（SQLite）

### 🚧 计划中功能
- [ ] 代理规则和路由
- [ ] 流量统计和监控
- [ ] 系统代理自动配置
- [ ] 服务器健康检查
- [ ] 多协议支持（HTTP/HTTPS/Shadowsocks）
- [ ] 代理链支持
- [ ] 系统托盘
- [ ] 主题切换

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！

### 贡献指南

1. Fork 本项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

## 联系方式

如有问题或建议，欢迎通过以下方式联系：

- 提交 GitHub Issue
- 发送邮件至项目维护者

---

**重要提示**：
- 🎯 **本项目仅提供 GUI 桌面应用作为正式入口**
- ⚠️ `cmd/` 目录下的其他程序（`client`、`myproxy`、`server`）仅用于开发和测试
- 📝 本项目仅供学习和研究使用，请勿用于非法用途
