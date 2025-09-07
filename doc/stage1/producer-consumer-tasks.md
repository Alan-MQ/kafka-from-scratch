# Producer/Consumer 实现任务

## 🎯 当前任务概览

我已经创建了Producer和Consumer的接口框架，现在需要你来实现核心的TODO方法。这些接口将成为我们第一个完整示例的基础。

## 📝 你需要实现的方法

### 🏭 Producer 实现

#### 📁 文件: `pkg/producer/producer.go`

**1. `Send(topic string, message *common.Message) (int32, int64, error)`**

```go
// 实现思路：
func (p *MemoryProducer) Send(topic string, message *common.Message) (int32, int64, error) {
    // 直接调用broker的ProduceMessage方法
    return p.broker.ProduceMessage(topic, message)
}
```

**2. `SendWithKey(topic string, key, value []byte) (int32, int64, error)`**

```go
// 实现思路：
func (p *MemoryProducer) SendWithKey(topic string, key, value []byte) (int32, int64, error) {
    // 1. 创建Message对象
    message := common.NewMessage(key, value)
    
    // 2. 调用Send方法
    return p.Send(topic, message)
}
```

### 🛒 Consumer 实现

#### 📁 文件: `pkg/consumer/consumer.go`

**1. `Subscribe(topics []string) error`**

```go
// 实现思路：
func (c *MemoryConsumer) Subscribe(topics []string) error {
    // 1. 保存订阅的Topic列表
    c.subscribedTopics = topics
    
    // 2. 为每个Topic初始化offset跟踪
    for _, topicName := range topics {
        // 获取Topic信息
        topic, err := c.broker.GetTopic(topicName)
        if err != nil {
            return err
        }
        
        // 初始化该topic的offset map
        if c.offsets[topicName] == nil {
            c.offsets[topicName] = make(map[int32]int64)
        }
        
        // 为每个分区设置初始offset为0
        partitionCount := topic.GetPartitionCount()
        for i := int32(0); i < partitionCount; i++ {
            c.offsets[topicName][i] = 0
        }
    }
    
    return nil
}
```

**2. `Poll(maxMessages int) ([]*common.Message, error)`**

```go
// 实现思路：
func (c *MemoryConsumer) Poll(maxMessages int) ([]*common.Message, error) {
    var allMessages []*common.Message
    messagesCollected := 0
    
    // 遍历所有订阅的Topic
    for _, topicName := range c.subscribedTopics {
        if messagesCollected >= maxMessages {
            break
        }
        
        // 遍历该Topic的所有分区
        topicOffsets := c.offsets[topicName]
        for partitionID, currentOffset := range topicOffsets {
            if messagesCollected >= maxMessages {
                break
            }
            
            // 计算这个分区要拉取多少消息
            remainingMessages := maxMessages - messagesCollected
            
            // 从Broker拉取消息
            messages, err := c.broker.ConsumeMessages(topicName, partitionID, currentOffset, remainingMessages)
            if err != nil {
                return nil, err
            }
            
            // 添加到结果集
            allMessages = append(allMessages, messages...)
            messagesCollected += len(messages)
            
            // 更新offset
            if len(messages) > 0 {
                // offset更新为最后一条消息的offset + 1
                lastMessage := messages[len(messages)-1]
                c.offsets[topicName][partitionID] = lastMessage.Offset + 1
            }
        }
    }
    
    return allMessages, nil
}
```

**3. `Seek(topic string, partition int32, offset int64) error`**

```go
// 实现思路：
func (c *MemoryConsumer) Seek(topic string, partition int32, offset int64) error {
    // 1. 检查topic是否已订阅
    subscribed := false
    for _, subscribedTopic := range c.subscribedTopics {
        if subscribedTopic == topic {
            subscribed = true
            break
        }
    }
    
    if !subscribed {
        return fmt.Errorf("topic %s is not subscribed", topic)
    }
    
    // 2. 检查分区是否存在
    topicObj, err := c.broker.GetTopic(topic)
    if err != nil {
        return err
    }
    
    if partition < 0 || partition >= topicObj.GetPartitionCount() {
        return fmt.Errorf("partition %d does not exist in topic %s", partition, topic)
    }
    
    // 3. 更新offset
    if c.offsets[topic] == nil {
        c.offsets[topic] = make(map[int32]int64)
    }
    c.offsets[topic][partition] = offset
    
    return nil
}
```

## 🧪 测试你的实现

实现完成后，我们将创建一个完整的示例来测试：

```go
// 创建Broker
broker := broker.NewMemoryBroker()
broker.CreateTopic("test-topic", 3)

// 创建Producer
producer := producer.NewMemoryProducer(broker)

// 发送消息
partitionID, offset, err := producer.SendWithKey("test-topic", 
    []byte("user-123"), []byte("Hello Kafka!"))

// 创建Consumer
consumer := consumer.NewMemoryConsumer(broker)
consumer.Subscribe([]string{"test-topic"})

// 消费消息
messages, err := consumer.Poll(10)
```

## 💡 实现提示

1. **错误处理**: 每个方法都要有合适的错误检查
2. **边界条件**: 注意空值、负数等情况
3. **并发安全**: 当前版本依赖Broker层的并发安全
4. **Offset管理**: Consumer需要正确跟踪和更新消费进度

## ❓ 可能遇到的问题

1. **导入路径**: 注意正确的包导入路径
2. **指针 vs 值**: 注意Broker是指针传递
3. **Map初始化**: 确保nested map被正确初始化
4. **Offset更新时机**: 在消费消息后及时更新offset

## ✅ 完成标准

- [ ] 所有TODO方法实现完成
- [ ] 代码能够编译通过
- [ ] Producer能正确发送消息
- [ ] Consumer能正确订阅和消费消息
- [ ] Offset跟踪功能正常

实现完成后，我们就可以编写第一个完整的示例程序了！

你准备好开始了吗？从哪个方法开始？