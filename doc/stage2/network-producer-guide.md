# 网络版Producer实现指导

## 🎯 你的任务
实现 `pkg/producer/network_producer.go` 中的3个TODO方法。

## 📋 方法实现指导

### 1. `Connect() error` 方法

**功能**: 建立与Broker的TCP连接

**实现步骤**:
```go
func (np *NetworkProducer) Connect() error {
    conn, err := net.Dial("tcp", np.brokerAddress)
    if err != nil {
        return fmt.Errorf("failed to connect to broker %s: %w", np.brokerAddress, err)
    }
    
    np.conn = conn
    fmt.Printf("🔗 已连接到Broker: %s\n", np.brokerAddress)
    return nil
}
```

### 2. `CreateTopic(name string, partitions int32) error` 方法

**功能**: 向Broker发送创建Topic请求

**实现步骤**:
```go
func (np *NetworkProducer) CreateTopic(name string, partitions int32) error {
    // 1. 创建请求数据
    createReq := &protocol.CreateTopicRequest{
        TopicName:    name,
        PartitionNum: partitions,
    }
    
    // 2. 包装到通用请求中
    request := &protocol.Request{
        Type:      protocol.RequestTypeCreateTopic,
        RequestID: "create-topic-1", // 简化版，实际应该用UUID
        Data:      createReq,
    }
    
    // 3. 发送请求
    response, err := np.sendRequest(request)
    if err != nil {
        return err
    }
    
    // 4. 检查响应
    if !response.Success {
        return fmt.Errorf("create topic failed: %s", response.Error)
    }
    
    fmt.Printf("✅ Topic创建成功: %s (%d个分区)\n", name, partitions)
    return nil
}
```

### 3. `Send(topic string, key, value []byte) (int32, int64, error)` 方法

**功能**: 发送消息到指定Topic

**实现步骤**:
```go
func (np *NetworkProducer) Send(topic string, key, value []byte) (int32, int64, error) {
    // 1. 创建请求数据
    produceReq := &protocol.ProduceRequest{
        TopicName:    topic,
        PartitionKey: string(key),
        Value:        string(value),
        Headers:      make(map[string]string),
    }
    
    // 2. 包装到通用请求中
    request := &protocol.Request{
        Type:      protocol.RequestTypeProduce,
        RequestID: "produce-1", // 简化版
        Data:      produceReq,
    }
    
    // 3. 发送请求
    response, err := np.sendRequest(request)
    if err != nil {
        return 0, 0, err
    }
    
    // 4. 检查响应
    if !response.Success {
        return 0, 0, fmt.Errorf("produce message failed: %s", response.Error)
    }
    
    // 5. 解析响应数据
    responseData, _ := json.Marshal(response.Data)
    var produceResp protocol.ProduceResponse
    json.Unmarshal(responseData, &produceResp)
    
    return produceResp.PartitionId, produceResp.Offset, nil
}
```

## 💡 实现提示

### RequestID生成
为了简化，可以使用递增数字或固定字符串：
```go
// 简化版
RequestID: "req-" + strconv.Itoa(rand.Int())

// 或者更简单
RequestID: "1"
```

### 错误处理
确保每个网络操作都有错误处理：
```go
if np.conn == nil {
    return fmt.Errorf("not connected to broker")
}
```

### JSON处理
响应数据的解析需要先序列化再反序列化：
```go
responseData, _ := json.Marshal(response.Data)
var specificResp protocol.SpecificResponse
json.Unmarshal(responseData, &specificResp)
```

## 🧪 测试方法

实现完成后，我们会创建一个简单的测试程序：

```go
// 测试流程
producer := NewNetworkProducer("localhost:9092")
producer.Connect()
producer.CreateTopic("test-topic", 3)
producer.Send("test-topic", []byte("key1"), []byte("Hello Network!"))
producer.Close()
```

## ✅ 完成标准

- [ ] Connect方法能成功连接到Broker
- [ ] CreateTopic能创建新的Topic
- [ ] Send能发送消息并返回正确的分区ID和offset
- [ ] 能够编译通过
- [ ] 基础的错误处理完善

你准备好开始实现了吗？从哪个方法开始？