package main

import (
	"fmt"
	"log"
	"time"

	"github.com/kafka-from-scratch/pkg/consumer"
)

func main() {
	// 创建网络版Consumer
	networkConsumer := consumer.NewNetworkConsumer("localhost:9092")
	
	fmt.Println("🔗 Consumer正在连接到Broker...")
	if err := networkConsumer.Connect(); err != nil {
		log.Fatalf("连接失败: %v", err)
	}
	defer networkConsumer.Close()
	
	fmt.Println("✅ Consumer连接成功！")
	
	// 1. 订阅Topic
	fmt.Println("\n📝 订阅Topic 'test-topic'...")
	if err := networkConsumer.Subscribe([]string{"test-topic"}); err != nil {
		log.Fatalf("订阅失败: %v", err)
	}
	fmt.Println("✅ 订阅成功！")
	
	// 2. 从不同分区消费消息
	fmt.Println("\n📥 开始消费消息...")
	
	partitions := []int32{0, 1, 2} // 3个分区
	for _, partitionId := range partitions {
		fmt.Printf("\n--- 分区 %d ---\n", partitionId)
		
		messages, err := networkConsumer.Consume("test-topic", partitionId, 0, 10)
		if err != nil {
			log.Printf("消费分区 %d 失败: %v", partitionId, err)
			continue
		}
		
		if len(messages) == 0 {
			fmt.Printf("分区 %d 暂无消息\n", partitionId)
		} else {
			for i, msg := range messages {
				fmt.Printf("消息 %d: key=%s, value=%s, offset=%d, timestamp=%s\n", 
					i+1, msg.Key, msg.Value, msg.Offset, msg.Timestamp)
			}
		}
		
		time.Sleep(200 * time.Millisecond)
	}
	
	// 3. 测试Seek操作
	fmt.Println("\n🎯 测试Seek操作...")
	if err := networkConsumer.Seek("test-topic", 2, 0); err != nil {
		log.Printf("Seek失败: %v", err)
	} else {
		fmt.Println("✅ Seek到分区2的offset 0成功")
		
		// 重新消费
		messages, err := networkConsumer.Consume("test-topic", 2, 0, 1)
		if err != nil {
			log.Printf("Seek后消费失败: %v", err)
		} else if len(messages) > 0 {
			msg := messages[0]
			fmt.Printf("Seek后消费到: key=%s, value=%s, offset=%d\n", 
				msg.Key, msg.Value, msg.Offset)
		}
	}
	
	fmt.Println("\n🎉 Consumer测试完成！")
}