# Kafka 核心概念

## 概述
Apache Kafka 是一个分布式流处理平台，主要用于构建实时数据管道和流应用程序。理解其核心概念是实现 Mini Kafka 的基础。

## 核心概念

### 1. Message (消息)
消息是 Kafka 中数据传输的基本单位。

**特性**:
- 由 Key、Value 和 Headers 组成
- Key 用于分区路由 (可选)
- Value 是实际的消息内容
- Headers 存储元数据信息

**在我们的实现中**:
```go
type Message struct {
    Key       []byte    // 消息键 (用于分区)
    Value     []byte    // 消息内容
    Headers   map[string]string // 头部信息
    Timestamp time.Time // 时间戳
    Offset    int64     // 偏移量 (由 Broker 分配)
}
```

### 2. Topic (主题)
Topic 是消息的分类，类似于数据库中的表。

**特性**:
- 逻辑上的消息分类
- 由多个 Partition 组成
- Producer 发送消息到特定 Topic
- Consumer 从特定 Topic 消费消息

**命名规则**:
- 只能包含字母、数字、点(.)、下划线(_)、连字符(-)
- 不能超过 249 个字符

### 3. Partition (分区)
Partition 是 Topic 的物理分片，是 Kafka 并行处理的基础。

**重要特性**:
- 每个 Partition 是一个有序的消息队列
- 消息在 Partition 内部有序，但 Topic 层面无序
- 每个 Partition 只能被同一消费者组中的一个 Consumer 消费
- 通过 Key 的哈希值决定消息发送到哪个 Partition

**分区的好处**:
- **并行处理**: 多个 Consumer 可以同时处理不同分区
- **水平扩展**: 可以通过增加分区数来提高吞吐量
- **负载均衡**: 消息分散到不同分区

### 4. Offset (偏移量)
Offset 是消息在 Partition 中的唯一标识和位置信息。

**特性**:
- 从 0 开始的递增整数
- 每个 Partition 独立维护
- Consumer 通过 Offset 跟踪消费进度
- 可以重置 Offset 来重新消费历史消息

### 5. Producer (生产者)
Producer 负责向 Kafka 发送消息。

**工作流程**:
1. 选择目标 Topic
2. 根据 Key 或轮询策略选择 Partition
3. 发送消息到指定的 Broker
4. 接收 Broker 的确认响应

**关键配置**:
- `acks`: 确认级别 (0, 1, all)
- `retries`: 重试次数
- `batch.size`: 批量发送大小

### 6. Consumer (消费者)
Consumer 负责从 Kafka 消费消息。

**工作流程**:
1. 订阅一个或多个 Topic
2. 从分配的 Partition 拉取消息
3. 处理消息
4. 提交 Offset (手动或自动)

**消费模式**:
- **Push**: Broker 主动推送 (Kafka 不采用)
- **Pull**: Consumer 主动拉取 (Kafka 采用)

### 7. Consumer Group (消费者组)
Consumer Group 是多个 Consumer 的集合，共同消费一个 Topic。

**重要特性**:
- 同一消费者组中的 Consumer 不会重复消费同一条消息
- 每个 Partition 只能被组内一个 Consumer 消费
- 支持自动负载均衡和故障转移
- 不同消费者组之间相互独立

**Rebalance (重平衡)**:
- Consumer 加入或离开时触发
- 重新分配 Partition 给 Consumer
- 短暂中断消费，但保证最终一致性

### 8. Broker (代理)
Broker 是 Kafka 集群中的服务器节点。

**职责**:
- 存储和管理 Topic 的 Partition
- 处理 Producer 的写请求
- 处理 Consumer 的读请求
- 维护集群元数据

**特性**:
- 每个 Broker 都有唯一的 ID
- 可以水平扩展
- 通过副本机制保证高可用

## 消息流转过程

### 生产流程
```
Producer → 选择 Topic → 选择 Partition → 发送到 Broker → 返回 Offset
```

### 消费流程  
```
Consumer → 订阅 Topic → 分配 Partition → 拉取消息 → 处理消息 → 提交 Offset
```

## 与传统消息队列的区别

| 特性 | 传统 MQ | Kafka |
|------|---------|--------|
| 消息模型 | 点对点/发布订阅 | 发布订阅 |
| 消息存储 | 消费后删除 | 持久化存储 |
| 消费模式 | Push | Pull |
| 顺序保证 | 队列级别 | 分区级别 |
| 扩展性 | 有限 | 高度可扩展 |

## 在我们的 Mini Kafka 中

### 第一阶段实现
- 基础的 Message 结构
- 内存版 Topic 管理
- 简单的 Producer/Consumer
- 单分区模式

### 后续阶段扩展
- 多分区支持
- 持久化存储
- 网络通信
- 消费者组
- 副本机制

## 学习要点

1. **理解分区概念**: 分区是 Kafka 高性能的关键
2. **掌握 Offset 管理**: Offset 是消费进度的核心
3. **理解消费者组**: 实现负载均衡和高可用的基础
4. **学习拉取模式**: 不同于传统 MQ 的推送模式

这些概念将在我们的实现过程中逐步体现和深化理解。