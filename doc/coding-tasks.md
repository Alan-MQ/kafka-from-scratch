# 编码任务指南

## 当前进度
我已经为你创建了基础的项目结构和核心数据结构：
- ✅ Message 结构体 (`internal/common/message.go`)
- ✅ Partition 结构体 (`internal/common/partition.go`)  
- ✅ Topic 结构体框架 (`internal/common/topic.go`) - **有 TODO 等你实现**
- ✅ MemoryBroker 结构体框架 (`internal/broker/memory_broker.go`) - **有 TODO 等你实现**

## 🎯 你的第一个编码任务

### 任务1: 实现 Topic 中的 TODO 方法

#### 📁 文件: `internal/common/topic.go`

#### 🔧 需要实现的方法:

**1. `GetPartitionForKey(key []byte) *Partition`**
```go
// 功能：根据消息的Key选择合适的分区
// 算法步骤：
// 1. 如果 key 为空，返回第一个分区 (t.Partitions[0])
// 2. 使用 hash/fnv 包计算 key 的哈希值
// 3. 用哈希值对分区数量取模，得到分区ID
// 4. 注意处理负数情况（如果结果为负，取绝对值）

// 示例代码框架：
if len(key) == 0 {
    // 返回第一个分区
}
hash := fnv.New32a()
// 写入 key 到 hash
// 计算哈希值
// 对分区数量取模
// 返回对应的分区
```

**2. `ProduceMessage(message *Message) (int32, int64, error)`**
```go
// 功能：将消息发送到合适的分区
// 步骤：
// 1. 使用 GetPartitionForKey 选择分区
// 2. 调用分区的 Append 方法添加消息
// 3. 返回 (分区ID, offset, nil)

partition := t.GetPartitionForKey(message.Key)
// 调用 partition.Append(message)
// 返回结果
```

### 任务2: 实现 MemoryBroker 中的 TODO 方法

#### 📁 文件: `internal/broker/memory_broker.go`

#### 🔧 需要实现的方法:

**1. `CreateTopic(name string, partitions int32) error`**
```go
// 检查 Topic 是否已存在
// 如果存在，返回错误
// 如果不存在，创建新 Topic 并存储
// 记住使用写锁保护并发访问
```

**2. `GetTopic(name string) (*common.Topic, error)`**
```go
// 从 topics map 中查找
// 使用读锁保护
// 如果不存在返回适当的错误
```

**3. `ProduceMessage(topicName string, message *common.Message) (int32, int64, error)`**
```go
// 1. 获取 Topic
// 2. 调用 Topic 的 ProduceMessage 方法
// 3. 返回结果
```

**4. `ConsumeMessages(topicName string, partition int32, offset int64, maxMessages int) ([]*common.Message, error)`**
```go
// 1. 获取 Topic  
// 2. 获取指定的分区
// 3. 调用分区的 GetMessages 方法
```

## 🧪 测试你的实现

实现完成后，创建一个简单的测试文件来验证：

```go
// 在 examples/ 目录下创建 test_basic.go
broker := broker.NewMemoryBroker()

// 创建 Topic
broker.CreateTopic("test-topic", 3)

// 发送消息
message := common.NewMessage([]byte("user-123"), []byte("Hello Kafka!"))
partitionID, offset, err := broker.ProduceMessage("test-topic", message)

// 消费消息
messages, err := broker.ConsumeMessages("test-topic", partitionID, 0, 10)
```

## 💡 提示和注意事项

1. **并发安全**: 使用 `sync.RWMutex` 保护共享数据
2. **错误处理**: 每个方法都要有适当的错误检查
3. **边界条件**: 检查空值、负数等边界情况
4. **代码风格**: 保持与现有代码一致的风格

## 🆘 如果遇到问题

- 不确定某个函数怎么用？问我！
- 编译错误？让我帮你看看
- 逻辑不清楚？我们一起讨论

## ✅ 完成标准

- 所有 TODO 方法都实现完成
- 代码能够编译通过
- 基本的消息发送和接收功能正常

完成这些任务后，我们将进入下一阶段：创建 Producer 和 Consumer 的接口！

你准备好开始编码了吗？从哪个方法开始？