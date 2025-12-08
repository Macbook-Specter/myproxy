# 实现SOCKS5客户端订阅、延迟测试和自动代理功能

## 1. 配置结构扩展

* **修改文件**: `internal/config/config.go`

* **新增字段**:

  * `SubscriptionURL`：订阅URL

  * `Servers`：服务器列表

  * `SelectedServer`：当前选中的服务器

  * `AutoProxyPort`：自动代理端口

  * `LogLevel`：日志级别

## 2. 服务器模型和订阅管理

* **新增文件**: `internal/server/server.go`

  * 定义`Server`结构体，包含服务器信息和状态

  * 实现服务器列表管理功能

* **新增文件**: `internal/subscription/subscription.go`

  * 实现从URL获取订阅服务器列表

  * 支持不同格式的订阅解析

  * 实现订阅更新功能

## 3. 延迟测试功能

* **新增文件**: `internal/ping/ping.go`

  * 实现TCP ping功能测试服务器延迟

  * 支持单个服务器测试和批量测试

## 4. 自动代理功能

* **修改文件**: `internal/proxy/forwarder.go`

  * 扩展`Forwarder`结构体，支持自动代理模式

  * 实现监听固定端口，将数据转发到选中的服务器

## 5. 日志记录功能

* **新增文件**: `internal/logging/logging.go`

  * 实现日志记录功能

  * 支持不同级别日志

  * 支持文件日志和控制台日志

## 6. 桌面应用扩展

* **修改文件**: `cmd/gui/main.go`

  * 添加新的API端点：

    * `/api/subscription/fetch`：获取订阅

    * `/api/subscription/update`：更新订阅

    * `/api/server/test`：测试服务器延迟

    * `/api/server/testall`：测试所有服务器延迟

    * `/api/server/select`：选择服务器

    * `/api/autoproxy/start`：启动自动代理

    * `/api/autoproxy/stop`：停止自动代理

    * `/api/logs`：获取日志

* **更新HTML/JS/CSS界面**：

  * 添加订阅管理面板

  * 添加服务器列表和状态显示

  * 添加延迟测试按钮

  * 添加自动代理控制

  * 添加日志查看功能

## 7. 实现顺序

1. 扩展配置结构
2. 实现服务器模型和订阅管理
3. 实现延迟测试功能
4. 实现日志记录功能
5. 扩展转发器支持自动代理
6. 更新桌面应用API和UI
7. 测试所有功能

## 8. 核心功能实现

* **订阅管理**：支持从URL获取和更新服务器列表

* **延迟测试**：TCP ping测试，显示延迟结果

* **自动代理**：监听固定端口，自动转发到选中服务器

* **一键操作**：一键测试所有服务器，一键测试选中服务器

* **日志记录**：记录转发日志，支持查看

* **桌面界面**：所有功能都可以通过桌面应用完成

这个计划将实现用户需求的所有功能，同时保持代码的模块化和可维护性。
