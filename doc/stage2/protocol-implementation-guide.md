# 协议实现指导

## 🎯 你的第一个任务：完成协议数据结构

### 📁 需要完成的文件
1. `internal/protocol/requests.go` - 请求结构
2. `internal/protocol/responses.go` - 响应结构

### 🔧 具体实现指导

#### 1. CreateTopicRequest 
```go
type CreateTopicRequest struct {
    TopicName      string `json:"topic_name"`
    PartitionCount int32  `json:"partition_count"`
}
```
**作用**: 客户端请求创建新的Topic

#### 2. ProduceRequest
```go
type ProduceRequest struct {
    TopicName string            `json:"topic_name"`
    Key       []byte            `json:"key,omitempty"`
    Value     []byte            `json:"value"`
    Headers   map[string]string `json:"headers,omitempty"`
}
```
**作用**: 客户端发送消息到指定Topic

#### 3. ConsumeRequest  
```go
type ConsumeRequest struct {
    TopicName   string `json:"topic_name"`
    PartitionID int32  `json:"partition_id"`
    Offset      int64  `json:"offset"`
    MaxMessages int    `json:"max_messages"`
}
```
**作用**: 客户端从指定分区消费消息

#### 4. SubscribeRequest
```go
type SubscribeRequest struct {
    Topics []string `json:"topics"`
}
```
**作用**: 客户端订阅一个或多个Topic

#### 5. SeekRequest
```go
type SeekRequest struct {
    TopicName   string `json:"topic_name"`
    PartitionID int32  `json:"partition_id"`
    Offset      int64  `json:"offset"`
}
```
**作用**: 客户端设置消费位置

## 响应结构

#### 1. CreateTopicResponse
```go
type CreateTopicResponse struct {
    Message string `json:"message"`
}
```

#### 2. ProduceResponse  
```go
type ProduceResponse struct {
    PartitionID int32 `json:"partition_id"`
    Offset      int64 `json:"offset"`
}
```

#### 3. ConsumeResponse
```go
type ConsumeResponse struct {
    Messages []*common.Message `json:"messages"`
}
```

#### 4. SubscribeResponse
```go
type SubscribeResponse struct {
    Message string `json:"message"`
}
```

#### 5. SeekResponse  
```go
type SeekResponse struct {
    Message string `json:"message"`
}
```

## 💡 实现提示

1. **JSON标签**: 使用合适的json标签
2. **omitempty**: 可选字段使用omitempty
3. **类型选择**: 注意int32 vs int64的选择
4. **导入**: 需要导入common包来使用Message类型

## ✅ 完成标准
- 所有结构体定义完成
- JSON标签正确  
- 能够编译通过
- 字段类型与现有代码一致

你准备好开始实现了吗？从哪个结构体开始？