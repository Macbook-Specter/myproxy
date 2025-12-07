package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"myproxy.com/p/internal/config"
	"myproxy.com/p/internal/proxy"
)

func main() {
	// 命令行参数解析
	configFile := flag.String("config", "./config.json", "配置文件路径")
	proxyAddr := flag.String("proxy", "", "SOCKS5代理服务器地址 (覆盖配置文件)")
	localAddr := flag.String("local", "", "本地监听地址 (覆盖配置文件)")
	remoteAddr := flag.String("remote", "", "远程目标地址 (覆盖配置文件)")
	protocol := flag.String("proto", "tcp", "协议类型 (tcp/udp) (覆盖配置文件)")
	username := flag.String("username", "", "认证用户名 (覆盖配置文件)")
	password := flag.String("password", "", "认证密码 (覆盖配置文件)")
	flag.Parse()

	// 加载配置文件
	cfg, err := config.LoadConfig(*configFile)
	if err != nil {
		log.Fatalf("加载配置文件失败: %v", err)
	}

	// 使用命令行参数覆盖配置文件
	if *proxyAddr != "" {
		cfg.ProxyAddr = *proxyAddr
	}
	if *username != "" {
		cfg.Username = *username
	}
	if *password != "" {
		cfg.Password = *password
	}

	// 创建转发器列表
	var forwarders []*proxy.Forwarder

	// 如果指定了远程地址，则只启动单个转发规则
	if *remoteAddr != "" {
		local := *localAddr
		if local == "" {
			local = "127.0.0.1:8080"
		}

		forwarder := proxy.NewForwarder(cfg.ProxyAddr, local, *remoteAddr, *protocol, cfg.Username, cfg.Password)
		forwarders = append(forwarders, forwarder)
	} else {
		// 否则从配置文件加载所有启用的转发规则
		for _, rule := range cfg.ForwardRules {
			if rule.Enabled {
				forwarder := proxy.NewForwarder(cfg.ProxyAddr, rule.LocalAddr, rule.RemoteAddr, rule.Protocol, cfg.Username, cfg.Password)
				forwarders = append(forwarders, forwarder)
			}
		}
	}

	// 检查是否有转发规则
	if len(forwarders) == 0 {
		log.Fatalf("没有启用的转发规则")
	}

	// 启动所有转发器
	for _, forwarder := range forwarders {
		if err := forwarder.Start(); err != nil {
			log.Fatalf("启动转发器失败: %v", err)
		}
	}

	// 设置信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Printf("按 Ctrl+C 停止所有转发...")

	// 等待终止信号
	<-sigChan

	// 停止所有转发器
	for _, forwarder := range forwarders {
		if err := forwarder.Stop(); err != nil {
			log.Printf("停止转发器失败: %v", err)
		} else {
			log.Println("转发器已成功停止")
		}
	}
}
