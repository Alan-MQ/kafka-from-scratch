# 阶段2: 网络通信协议设计

## 🎯 目标
从单机内存版本迈向真正的分布式系统，实现Producer/Consumer通过网络与Broker通信。

## 🤔 协议设计的核心问题

### 1. 选择协议格式
**选项对比**:
- **JSON**: 人类可读，调试简单，但性能较低
- **Binary**: 性能最优，但实现复杂
- **Protobuf**: 平衡性能和开发效率

**我们的选择**: 先用JSON实现，便于调试和学习

### 2. 消息结构设计
需要支持的操作:
- CreateTopic
- ProduceMessage  
- ConsumeMessages
- Subscribe
- Seek

## 📋 第一个任务：设计请求/响应结构

### 你需要实现的结构体 (在 `internal/protocol/` 目录)

```go
// 请求类型枚举
type RequestType string

const (
    RequestTypeCreateTopic    RequestType = "CREATE_TOPIC"
    RequestTypeProduce        RequestType = "PRODUCE"
    RequestTypeConsume        RequestType = "CONSUME"
    RequestTypeSubscribe      RequestType = "SUBSCRIBE"
    RequestTypeSeek           RequestType = "SEEK"
)

// 通用请求结构
type Request struct {
    Type      RequestType `json:"type"`
    RequestID string      `json:"request_id"`
    Data      interface{} `json:"data"`
}

// 通用响应结构
type Response struct {
    RequestID string      `json:"request_id"`
    Success   bool        `json:"success"`
    Error     string      `json:"error,omitempty"`
    Data      interface{} `json:"data,omitempty"`
}
```

### TODO: 你来定义具体的请求数据结构

在 `internal/protocol/requests.go` 文件中实现：

1. **CreateTopicRequest** - 创建Topic的请求数据
2. **ProduceRequest** - 生产消息的请求数据  
3. **ConsumeRequest** - 消费消息的请求数据
4. **SubscribeRequest** - 订阅Topic的请求数据
5. **SeekRequest** - 设置offset的请求数据

### TODO: 你来定义响应数据结构

在 `internal/protocol/responses.go` 文件中实现：

1. **CreateTopicResponse** - 创建Topic的响应
2. **ProduceResponse** - 生产消息的响应
3. **ConsumeResponse** - 消费消息的响应
4. **SubscribeResponse** - 订阅的响应
5. **SeekResponse** - Seek操作的响应

## 💡 实现提示

### CreateTopicRequest 应该包含什么？
- topic name
- partition count

### ProduceRequest 应该包含什么？
- topic name
- message (key, value, headers)

### ConsumeRequest 应该包含什么？
- topic name
- partition id
- offset
- max messages

你觉得这个设计方向如何？准备好开始实现第一个协议文件了吗？