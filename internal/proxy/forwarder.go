package proxy

import (
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"myproxy.com/p/internal/socks5"
	"myproxy.com/p/internal/server"
)

// LogCallback 定义日志回调函数类型
// 参数：level (日志级别，如 "INFO", "ERROR"), logType (日志类型，如 "proxy"), message (日志消息)
type LogCallback func(level, logType, message string)

// Forwarder 定义端口转发器结构体
type Forwarder struct {
	SOCKS5Client   *socks5.SOCKS5Client
	LocalAddr      string // 本地监听地址 (e.g., "127.0.0.1:8080")
	RemoteAddr     string // 远程目标地址 (e.g., "example.com:80")
	Protocol       string // 协议类型 ("tcp" 或 "udp")
	IsRunning      bool
	IsAutoProxy    bool   // 是否为自动代理模式
	ServerManager  *server.ServerManager // 服务器管理器，用于自动代理模式
	listener       net.Listener
	udpConn        *net.UDPConn
	logCallback    LogCallback // 日志回调函数
}

// NewForwarder 创建一个新的端口转发器
func NewForwarder(proxyAddr, localAddr, remoteAddr, protocol string, username, password string) *Forwarder {
	return &Forwarder{
		SOCKS5Client: &socks5.SOCKS5Client{
			ProxyAddr: proxyAddr,
			Username:  username,
			Password:  password,
		},
		LocalAddr:  localAddr,
		RemoteAddr: remoteAddr,
		Protocol:   protocol,
		IsRunning:  false,
		IsAutoProxy: false,
	}
}

// NewAutoProxyForwarder 创建一个新的自动代理转发器
func NewAutoProxyForwarder(localAddr, protocol string, serverManager *server.ServerManager) *Forwarder {
	return &Forwarder{
		LocalAddr:     localAddr,
		Protocol:      protocol,
		IsRunning:     false,
		IsAutoProxy:   true,
		ServerManager: serverManager,
	}
}

// SetLogCallback 设置日志回调函数
func (f *Forwarder) SetLogCallback(callback LogCallback) {
	f.logCallback = callback
}

// log 内部日志方法，优先使用回调函数，否则使用标准log
func (f *Forwarder) log(level, logType, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	if f.logCallback != nil {
		f.logCallback(level, logType, message)
	} else {
		// 如果没有设置回调，使用标准log输出
		log.Printf("[%s][%s] %s", level, logType, message)
	}
}

// Start 启动端口转发
func (f *Forwarder) Start() error {
	if f.IsRunning {
		return fmt.Errorf("转发器已在运行")
	}

	if f.IsAutoProxy {
		// 自动代理模式
		f.log("INFO", "proxy", "启动自动代理 %s 转发: %s (自动选择 SOCKS5 代理)", 
			f.Protocol, f.LocalAddr)
	} else {
		// 普通代理模式
		f.log("INFO", "proxy", "启动 %s 转发: %s -> %s (通过 SOCKS5 代理 %s)", 
			f.Protocol, f.LocalAddr, f.RemoteAddr, f.SOCKS5Client.ProxyAddr)
	}

	if f.Protocol == "tcp" {
		return f.startTCP()
	} else if f.Protocol == "udp" {
		return f.startUDP()
	}

	return fmt.Errorf("不支持的协议类型: %s", f.Protocol)
}

// Stop 停止端口转发
func (f *Forwarder) Stop() error {
	if !f.IsRunning {
		return fmt.Errorf("转发器未运行")
	}

	f.log("INFO", "proxy", "停止 %s 转发: %s -> %s", f.Protocol, f.LocalAddr, f.RemoteAddr)

	if f.Protocol == "tcp" {
		return f.stopTCP()
	} else if f.Protocol == "udp" {
		return f.stopUDP()
	}

	return fmt.Errorf("不支持的协议类型: %s", f.Protocol)
}

// startTCP 启动TCP端口转发
func (f *Forwarder) startTCP() error {
	listener, err := net.Listen("tcp", f.LocalAddr)
	if err != nil {
		return fmt.Errorf("监听本地端口失败: %w", err)
	}

	f.listener = listener
	f.IsRunning = true

	go f.tcpForwardLoop()

	return nil
}

// tcpForwardLoop 处理TCP连接转发
func (f *Forwarder) tcpForwardLoop() {
	defer func() {
		f.IsRunning = false
		f.listener.Close()
	}()

	for {
		// 接受本地连接
		localConn, err := f.listener.Accept()
		if err != nil {
			if !f.IsRunning {
				return
			}
			f.log("ERROR", "proxy", "接受本地连接失败: %v", err)
			continue
		}

		// 处理连接转发
		go f.handleTCPConnection(localConn)
	}
}

// handleTCPConnection 处理单个TCP连接的转发
func (f *Forwarder) handleTCPConnection(localConn net.Conn) {
	defer localConn.Close()

	f.log("INFO", "proxy", "收到本地TCP连接: %s -> %s", localConn.RemoteAddr(), f.LocalAddr)

	// 如果是自动代理模式，作为SOCKS5服务器处理请求
	if f.IsAutoProxy {
		f.handleSOCKS5Request(localConn)
		return
	}

	// 普通转发模式
	// 通过SOCKS5代理连接到远程服务器
	proxyConn, err := f.SOCKS5Client.Dial("tcp", f.RemoteAddr)
	if err != nil {
		f.log("ERROR", "proxy", "通过SOCKS5代理连接到远程服务器失败: %v", err)
		return
	}
	defer proxyConn.Close()

	f.log("INFO", "proxy", "成功通过SOCKS5代理连接到远程服务器: %s", f.RemoteAddr)

	// 设置超时
	localConn.SetDeadline(time.Now().Add(30 * time.Second))
	proxyConn.SetDeadline(time.Now().Add(30 * time.Second))

	// 双向数据转发
	go func() {
		if _, err := io.Copy(proxyConn, localConn); err != nil {
			if !f.IsRunning {
				return
			}
			f.log("ERROR", "proxy", "本地到远程数据转发失败: %v", err)
		}
	}()

	if _, err := io.Copy(localConn, proxyConn); err != nil {
		if !f.IsRunning {
			return
		}
		f.log("ERROR", "proxy", "远程到本地数据转发失败: %v", err)
	}

	f.log("INFO", "proxy", "TCP连接关闭: %s -> %s", localConn.RemoteAddr(), f.RemoteAddr)
}

// handleSOCKS5Request 处理SOCKS5请求（自动代理模式）
func (f *Forwarder) handleSOCKS5Request(conn net.Conn) {
	// 1. 读取客户端的SOCKS5握手请求
	handshake := make([]byte, 256)
	n, err := conn.Read(handshake)
	if err != nil {
		f.log("ERROR", "proxy", "读取SOCKS5握手请求失败: %v", err)
		return
	}

	// 检查版本号
	if handshake[0] != 0x05 {
		f.log("ERROR", "proxy", "不支持的SOCKS版本: %d", handshake[0])
		return
	}

	// 2. 回复客户端支持的认证方法
	authReply := []byte{0x05, 0x00} // 只支持无认证
	if _, err := conn.Write(authReply); err != nil {
		f.log("ERROR", "proxy", "发送SOCKS5认证回复失败: %v", err)
		return
	}

	// 3. 读取客户端的SOCKS5请求
	request := make([]byte, 256)
	n, err = conn.Read(request)
	if err != nil {
		f.log("ERROR", "proxy", "读取SOCKS5请求失败: %v", err)
		return
	}

	// 检查请求格式
	if n < 10 || request[0] != 0x05 {
		f.log("ERROR", "proxy", "无效的SOCKS5请求")
		return
	}

	// 解析目标地址
	targetAddr, err := parseSOCKS5TargetAddr(request)
	if err != nil {
		f.log("ERROR", "proxy", "解析SOCKS5目标地址失败: %v", err)
		return
	}

	f.log("INFO", "proxy", "收到SOCKS5请求: %s -> %s", conn.RemoteAddr(), targetAddr)

	// 4. 获取选中的服务器
	selectedServer, err := f.ServerManager.GetSelectedServer()
	if err != nil {
		f.log("ERROR", "proxy", "获取选中的服务器失败: %v", err)
		// 发送SOCKS5错误响应
		errReply := []byte{0x05, 0x04, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
		conn.Write(errReply)
		return
	}

	// 5. 创建SOCKS5客户端连接到选中的服务器
	socks5Client := &socks5.SOCKS5Client{
		ProxyAddr: fmt.Sprintf("%s:%d", selectedServer.Addr, selectedServer.Port),
		Username:  selectedServer.Username,
		Password:  selectedServer.Password,
	}

	// 6. 通过选中的SOCKS5服务器连接到目标地址
	proxyConn, err := socks5Client.Dial("tcp", targetAddr)
	if err != nil {
		f.log("ERROR", "proxy", "通过选中的SOCKS5服务器连接到目标地址失败: %v", err)
		// 发送SOCKS5错误响应
		errReply := []byte{0x05, 0x05, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
		conn.Write(errReply)
		return
	}
	defer proxyConn.Close()

	// 7. 发送SOCKS5成功响应
	successReply := []byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	if _, err := conn.Write(successReply); err != nil {
		f.log("ERROR", "proxy", "发送SOCKS5成功响应失败: %v", err)
		return
	}

	f.log("INFO", "proxy", "成功通过SOCKS5代理连接到目标地址: %s (服务器: %s)", targetAddr, selectedServer.Name)

	// 8. 设置超时
	conn.SetDeadline(time.Now().Add(30 * time.Second))
	proxyConn.SetDeadline(time.Now().Add(30 * time.Second))

	// 9. 双向数据转发
	go func() {
		if _, err := io.Copy(proxyConn, conn); err != nil {
			if !f.IsRunning {
				return
			}
			f.log("ERROR", "proxy", "客户端到代理数据转发失败: %v", err)
		}
	}()

	if _, err := io.Copy(conn, proxyConn); err != nil {
		if !f.IsRunning {
			return
		}
		f.log("ERROR", "proxy", "代理到客户端数据转发失败: %v", err)
	}

	f.log("INFO", "proxy", "SOCKS5连接关闭: %s -> %s", conn.RemoteAddr(), targetAddr)
}

// parseSOCKS5TargetAddr 解析SOCKS5请求中的目标地址
func parseSOCKS5TargetAddr(request []byte) (string, error) {
	// 请求格式: VER (1) + CMD (1) + RSV (1) + ATYP (1) + DST.ADDR + DST.PORT
	if len(request) < 10 {
		return "", fmt.Errorf("无效的SOCKS5请求")
	}

	atyp := request[3]
	var targetAddr string
	var port int

	switch atyp {
	case 0x01: // IPv4
		if len(request) < 10 {
			return "", fmt.Errorf("IPv4地址太短")
		}
		ip := net.IP(request[4:8]).String()
		port = int(request[8])<<8 | int(request[9])
		targetAddr = fmt.Sprintf("%s:%d", ip, port)
	case 0x03: // 域名
		addrLen := int(request[4])
		if len(request) < 5+addrLen+2 {
			return "", fmt.Errorf("域名地址太短")
		}
		domain := string(request[5 : 5+addrLen])
		port = int(request[5+addrLen])<<8 | int(request[5+addrLen+1])
		targetAddr = fmt.Sprintf("%s:%d", domain, port)
	case 0x04: // IPv6
		if len(request) < 22 {
			return "", fmt.Errorf("IPv6地址太短")
		}
		ip := net.IP(request[4:20]).String()
		port = int(request[20])<<8 | int(request[21])
		targetAddr = fmt.Sprintf("%s:%d", ip, port)
	default:
		return "", fmt.Errorf("不支持的地址类型: %d", atyp)
	}

	return targetAddr, nil
}

// startUDP 启动UDP端口转发
func (f *Forwarder) startUDP() error {
	// 解析本地地址
	localAddr, err := net.ResolveUDPAddr("udp", f.LocalAddr)
	if err != nil {
		return fmt.Errorf("解析本地UDP地址失败: %w", err)
	}

	// 监听本地UDP端口
	conn, err := net.ListenUDP("udp", localAddr)
	if err != nil {
		return fmt.Errorf("监听本地UDP端口失败: %w", err)
	}

	f.udpConn = conn
	f.IsRunning = true

	go f.udpForwardLoop()

	return nil
}

// udpForwardLoop 处理UDP数据转发
func (f *Forwarder) udpForwardLoop() {
	defer func() {
		f.IsRunning = false
		f.udpConn.Close()
	}()

	// 启动UDP关联
	tcpConn, proxyUDPAddr, err := f.SOCKS5Client.UDPAssociate()
	if err != nil {
		f.log("ERROR", "proxy", "UDP关联失败: %v", err)
		return
	}
	defer tcpConn.Close()

	f.log("INFO", "proxy", "成功获取代理服务器UDP地址: %s:%d", proxyUDPAddr.Host, proxyUDPAddr.Port)

	// 连接到代理服务器的UDP端口
	proxyNetAddr := fmt.Sprintf("%s:%d", proxyUDPAddr.Host, proxyUDPAddr.Port)
	remoteAddr, err := net.ResolveUDPAddr("udp", proxyNetAddr)
	if err != nil {
		f.log("ERROR", "proxy", "解析代理UDP地址失败: %v", err)
		return
	}

	// 用于保存客户端地址映射
	clientMap := make(map[string]*net.UDPAddr)

	buf := make([]byte, 65535)
	for {
		// 从本地UDP端口读取数据
		n, clientAddr, err := f.udpConn.ReadFromUDP(buf)
		if err != nil {
			if !f.IsRunning {
				return
			}
			f.log("ERROR", "proxy", "读取本地UDP数据失败: %v", err)
			continue
		}

		// 保存客户端地址映射
		clientKey := clientAddr.String()
		clientMap[clientKey] = clientAddr

		f.log("INFO", "proxy", "收到本地UDP数据: %d 字节，来自 %s", n, clientKey)

		// 封装SOCKS5 UDP数据包并发送到代理服务器
		socks5Packet, err := buildUDPPacket(f.RemoteAddr, buf[:n])
		if err != nil {
			f.log("ERROR", "proxy", "封装SOCKS5 UDP数据包失败: %v", err)
			continue
		}

		// 发送数据到代理服务器
		if _, err := f.udpConn.WriteToUDP(socks5Packet, remoteAddr); err != nil {
			f.log("ERROR", "proxy", "发送UDP数据到代理服务器失败: %v", err)
			continue
		}

		// 启动一个goroutine处理响应
		go f.handleUDPResponse(clientAddr, remoteAddr)
	}
}

// handleUDPResponse 处理UDP响应
func (f *Forwarder) handleUDPResponse(clientAddr, proxyAddr *net.UDPAddr) {
	buf := make([]byte, 65535)

	// 设置读取超时
	f.udpConn.SetReadDeadline(time.Now().Add(5 * time.Second))
	
	n, addr, err := f.udpConn.ReadFromUDP(buf)
	if err != nil {
		if !f.IsRunning {
			return
		}
		// 忽略超时错误
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return
		}
		f.log("ERROR", "proxy", "读取UDP响应失败: %v", err)
		return
	}

	// 只处理来自代理服务器的响应
	if addr.String() != proxyAddr.String() {
		return
	}

	f.log("INFO", "proxy", "收到来自代理服务器的UDP响应: %d 字节", n)

	// 解析SOCKS5 UDP响应
	// 响应格式: RSV (2) + FRAG (1) + ATYP (1) + DST.ADDR + DST.PORT + DATA
	if n < 10 {
		f.log("ERROR", "proxy", "UDP响应数据太短: %d 字节", n)
		return
	}

	// 跳过SOCKS5 UDP头部，提取实际数据
	// 简单处理，实际应根据ATYP动态计算头部长度
	data := buf[10:n]

	// 发送响应数据到客户端
	if _, err := f.udpConn.WriteToUDP(data, clientAddr); err != nil {
		f.log("ERROR", "proxy", "发送UDP响应到客户端失败: %v", err)
		return
	}

	f.log("INFO", "proxy", "成功转发UDP响应到客户端: %d 字节", len(data))
}

// stopTCP 停止TCP转发
func (f *Forwarder) stopTCP() error {
	f.IsRunning = false
	if f.listener != nil {
		return f.listener.Close()
	}
	return nil
}

// stopUDP 停止UDP转发
func (f *Forwarder) stopUDP() error {
	f.IsRunning = false
	if f.udpConn != nil {
		return f.udpConn.Close()
	}
	return nil
}

// buildUDPPacket 封装SOCKS5 UDP数据报头部
func buildUDPPacket(targetAddr string, data []byte) ([]byte, error) {
	// 实现与cmd/myproxy/main.go中相同的逻辑
	host, portStr, err := net.SplitHostPort(targetAddr)
	if err != nil {
		return nil, fmt.Errorf("目标地址格式错误: %w", err)
	}
	port, _ := net.LookupPort("udp", portStr)

	// 头部: RSV (2) + FRAG (1) + ATYP (1)
	// FRAG=0 表示这是一个完整的包
	buf := make([]byte, 0, 10+len(data))
	buf = append(buf, 0x00, 0x00, 0x00, socks5.ATypDomain)

	// 目标地址类型和值
	buf = append(buf, byte(len(host)))
	buf = append(buf, []byte(host)...)

	// 端口 (2 bytes 大端序)
	buf = append(buf, byte(port>>8), byte(port&0xff))

	// 负载数据
	buf = append(buf, data...)

	return buf, nil
}
