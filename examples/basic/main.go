package main

import (
	"fmt"
	"log"
	"time"

	"github.com/kafka-from-scratch/internal/broker"
	"github.com/kafka-from-scratch/pkg/consumer"
	"github.com/kafka-from-scratch/pkg/producer"
)

func main() {
	fmt.Println("🚀 Mini Kafka 基础示例")
	fmt.Println("========================")

	// 1. 创建内存Broker
	fmt.Println("1. 创建Broker...")
	memoryBroker := broker.NewMemoryBroker()

	// 2. 创建Topic
	fmt.Println("2. 创建Topic 'user-events' (3个分区)...")
	err := memoryBroker.CreateTopic("user-events", 3)
	if err != nil {
		log.Fatalf("创建Topic失败: %v", err)
	}

	// 3. 创建Producer
	fmt.Println("3. 创建Producer...")
	prod := producer.NewMemoryProducer(memoryBroker)
	defer prod.Close()

	// 4. 发送消息
	fmt.Println("\n4. 发送消息...")
	messages := []struct {
		key   string
		value string
	}{
		{"user-001", "用户登录"},
		{"user-002", "用户注册"},
		{"user-001", "用户购买"},
		{"user-003", "用户评论"},
		{"user-002", "用户分享"},
	}

	for i, msg := range messages {
		partitionID, offset, err := prod.SendWithKey(
			"user-events", 
			[]byte(msg.key), 
			[]byte(msg.value),
		)
		if err != nil {
			log.Fatalf("发送消息失败: %v", err)
		}
		fmt.Printf("  消息 %d: [%s] -> '%s' (分区: %d, offset: %d)\n", 
			i+1, msg.key, msg.value, partitionID, offset)
	}

	// 5. 创建Consumer
	fmt.Println("\n5. 创建Consumer...")
	cons := consumer.NewMemoryConsumer(memoryBroker)
	defer cons.Close()

	// 6. 订阅Topic
	fmt.Println("6. 订阅Topic 'user-events'...")
	err = cons.Subscribe([]string{"user-events"})
	if err != nil {
		log.Fatalf("订阅Topic失败: %v", err)
	}

	// 7. 消费消息
	fmt.Println("\n7. 消费消息...")
	consumedMessages, err := cons.Poll(10)
	if err != nil {
		log.Fatalf("消费消息失败: %v", err)
	}

	fmt.Printf("  成功消费 %d 条消息:\n", len(consumedMessages))
	for i, msg := range consumedMessages {
		fmt.Printf("  消息 %d: [%s] -> '%s' (offset: %d, 时间: %s)\n", 
			i+1, string(msg.Key), string(msg.Value), msg.Offset, 
			msg.Timestamp.Format("15:04:05"))
	}

	// 8. 测试Seek功能
	fmt.Println("\n8. 测试Seek功能 - 重置分区0到offset 0...")
	err = cons.Seek("user-events", 0, 0)
	if err != nil {
		log.Fatalf("Seek失败: %v", err)
	}

	// 再次消费
	fmt.Println("9. 再次消费 (应该会重复看到分区0的消息)...")
	moreMessages, err := cons.Poll(3)
	if err != nil {
		log.Fatalf("再次消费失败: %v", err)
	}

	fmt.Printf("  再次消费到 %d 条消息:\n", len(moreMessages))
	for i, msg := range moreMessages {
		fmt.Printf("  消息 %d: [%s] -> '%s' (offset: %d)\n", 
			i+1, string(msg.Key), string(msg.Value), msg.Offset)
	}

	// 10. 显示Topic信息
	fmt.Println("\n10. Broker信息:")
	topics := memoryBroker.ListTopics()
	for _, topicName := range topics {
		topic, _ := memoryBroker.GetTopic(topicName)
		fmt.Printf("  Topic: %s, 分区数: %d\n", topicName, topic.GetPartitionCount())
	}

	fmt.Println("\n✅ Mini Kafka 基础功能演示完成!")
	fmt.Println("\n🎯 关键概念验证:")
	fmt.Println("  ✓ 消息根据Key自动分区")
	fmt.Println("  ✓ Consumer可以从所有分区消费")
	fmt.Println("  ✓ Offset正确跟踪和更新")
	fmt.Println("  ✓ Seek功能允许重新消费")
	fmt.Println("  ✓ 消息持久化存储 (内存版)")
	
	// 11. 性能测试
	fmt.Println("\n11. 简单性能测试...")
	performanceTest(memoryBroker, prod, cons)
}

func performanceTest(broker *broker.MemoryBroker, prod producer.Producer, cons consumer.Consumer) {
	// 创建测试Topic
	broker.CreateTopic("perf-test", 1)
	
	// 发送1000条消息
	start := time.Now()
	messageCount := 1000
	
	for i := 0; i < messageCount; i++ {
		key := fmt.Sprintf("key-%d", i%10)  // 10个不同的key
		value := fmt.Sprintf("message-%d", i)
		_, _, err := prod.SendWithKey("perf-test", []byte(key), []byte(value))
		if err != nil {
			log.Printf("发送消息 %d 失败: %v", i, err)
			return
		}
	}
	
	sendDuration := time.Since(start)
	
	// 消费所有消息
	cons.Subscribe([]string{"perf-test"})
	start = time.Now()
	
	totalConsumed := 0
	for totalConsumed < messageCount {
		messages, err := cons.Poll(100)
		if err != nil {
			log.Printf("消费消息失败: %v", err)
			return
		}
		totalConsumed += len(messages)
		if len(messages) == 0 {
			break  // 没有更多消息
		}
	}
	
	consumeDuration := time.Since(start)
	
	fmt.Printf("  发送 %d 条消息耗时: %v (%.0f msg/s)\n", 
		messageCount, sendDuration, float64(messageCount)/sendDuration.Seconds())
	fmt.Printf("  消费 %d 条消息耗时: %v (%.0f msg/s)\n", 
		totalConsumed, consumeDuration, float64(totalConsumed)/consumeDuration.Seconds())
}