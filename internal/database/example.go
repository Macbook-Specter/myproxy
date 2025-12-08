package database

// 使用示例：
//
// 1. 在应用启动时初始化数据库：
//    import "myproxy.com/p/internal/database"
//
//    func main() {
//        // 初始化数据库（数据库文件路径，例如："./data/myproxy.db"）
//        if err := database.InitDB("./data/myproxy.db"); err != nil {
//            log.Fatalf("初始化数据库失败: %v", err)
//        }
//        defer database.CloseDB()
//
//        // 其他初始化代码...
//    }
//
// 2. 使用订阅管理器时，会自动保存到数据库：
//    import "myproxy.com/p/internal/subscription"
//
//    sm := subscription.NewSubscriptionManager(serverManager)
//
//    // 获取订阅（会自动保存到数据库）
//    servers, err := sm.FetchSubscription("https://example.com/subscription", "我的订阅标签")
//
//    // 更新订阅（会自动更新数据库）
//    err = sm.UpdateSubscription("https://example.com/subscription", "更新后的标签")
//
// 3. 直接使用数据库 API：
//    // 添加或更新订阅
//    sub, err := database.AddOrUpdateSubscription("https://example.com/sub", "标签")
//
//    // 获取所有订阅
//    subscriptions, err := database.GetAllSubscriptions()
//
//    // 添加或更新服务器
//    var subscriptionID *int64 = &sub.ID
//    err = database.AddOrUpdateServer(server, subscriptionID)
//
//    // 获取所有服务器
//    servers, err := database.GetAllServers()
//
//    // 根据订阅ID获取服务器
//    servers, err := database.GetServersBySubscriptionID(sub.ID)
