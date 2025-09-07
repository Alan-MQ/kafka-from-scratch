package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/kafka-from-scratch/internal/broker"
	"github.com/kafka-from-scratch/internal/server"
)

func main() {
	// 创建内存版Broker
	memoryBroker := broker.NewMemoryBroker()

	// 创建TCP服务器
	address := ":9092" // 使用Kafka默认端口
	tcpServer := server.NewTCPServer(address, memoryBroker)

	// 启动服务器
	fmt.Printf("🚀 Mini Kafka Broker 启动中...\n")
	fmt.Printf("📡 监听地址: %s\n", address)
	fmt.Printf("💡 使用 Ctrl+C 停止服务器\n\n")

	// 在goroutine中启动服务器
	go func() {
		if err := tcpServer.Start(); err != nil {
			log.Printf("TCP服务器错误: %v\n", err)
		}
	}()

	// 等待停止信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Printf("\n🛑 正在停止服务器...\n")
	if err := tcpServer.Stop(); err != nil {
		log.Printf("停止服务器时出错: %v\n", err)
	} else {
		fmt.Printf("✅ 服务器已安全停止\n")
	}
}
