package main

import (
	"fmt"
	"log"
	"time"

	"github.com/kafka-from-scratch/pkg/producer"
)

func main() {
	// 创建网络版Producer
	networkProducer := producer.NewNetworkProducer("localhost:9092")
	
	fmt.Println("🔗 正在连接到Broker...")
	if err := networkProducer.Connect(); err != nil {
		log.Fatalf("连接失败: %v", err)
	}
	defer networkProducer.Close()
	
	fmt.Println("✅ 连接成功！")
	
	// 1. 创建Topic
	fmt.Println("\n📝 创建Topic 'test-topic'...")
	if err := networkProducer.CreateTopic("test-topic", 3); err != nil {
		log.Fatalf("创建Topic失败: %v", err)
	}
	fmt.Println("✅ Topic创建成功！")
	
	// 2. 发送消息
	fmt.Println("\n📤 发送消息...")
	messages := []struct{
		key   string
		value string
	}{
		{"user-1", "Hello from Network Producer!"},
		{"user-2", "This is message 2"},
		{"user-1", "Another message from user-1"},
	}
	
	for i, msg := range messages {
		fmt.Printf("发送消息 %d: key=%s, value=%s\n", i+1, msg.key, msg.value)
		
		partitionId, offset, err := networkProducer.Send("test-topic", []byte(msg.key), []byte(msg.value))
		if err != nil {
			log.Printf("发送消息失败: %v", err)
			continue
		}
		
		fmt.Printf("✅ 消息已发送到分区 %d，offset: %d\n", partitionId, offset)
		time.Sleep(500 * time.Millisecond) // 稍微间隔一下
	}
	
	fmt.Println("\n🎉 测试完成！")
}