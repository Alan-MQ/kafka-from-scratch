# Consumer Group 机制设计详解

## 🤔 为什么需要Consumer Group？

### 现在的问题
在我们当前的实现中，Consumer需要手动指定：
```go
// 用户必须知道topic有哪些分区，手动指定从哪个分区消费
messages, err := consumer.Consume("test-topic", 2, 0, 10) // 手动指定分区2
```

这带来的问题：
1. **用户负担重**：必须了解topic内部结构
2. **容易出错**：分区ID写错、重复消费、遗漏分区
3. **难以扩缩容**：新增Consumer时需要手动重新分配
4. **无法负载均衡**：多个Consumer可能都读同一个分区

### Consumer Group的解决方案

```go
// 用户只需要指定groupId和要订阅的topic，分区分配完全自动化
consumer := NewGroupConsumer("my-app-group", "localhost:9092")
consumer.Subscribe([]string{"order-topic", "user-topic"})
messages := consumer.Poll() // 自动从分配给自己的分区消费
```

---

## 🏗️ 核心概念详解

### 1. Consumer Group
**定义**：一组Consumer实例，共同消费一个或多个Topic，实现负载均衡。

**核心特性**：
- Group内每个分区只能被一个Consumer消费（独占）
- Consumer数量可以动态变化
- 自动故障转移和负载均衡

### 2. Group Coordinator  
**定义**：Broker端负责管理Consumer Group的组件。

**职责**：
- 管理Group成员（加入/离开）
- 执行分区分配算法
- 协调Rebalance过程
- 管理Offset提交

### 3. Rebalance（重平衡）
**定义**：重新分配分区给Group内Consumer的过程。

**触发条件**：
- 新Consumer加入Group
- Consumer离开Group（正常退出或故障）
- 订阅的Topic分区数量变化

---

## 🎯 实际场景分析

### 场景1：订单处理系统

**背景**：
- Topic: `order-events` (6个分区)
- Consumer Group: `order-processor`
- 业务需求：每个订单事件只能被处理一次

**部署演进：**

#### 初始状态：1个Consumer
```
Topic: order-events (6 partitions)
┌─────┬─────┬─────┬─────┬─────┬─────┐
│ P0  │ P1  │ P2  │ P3  │ P4  │ P5  │
└─────┴─────┴─────┴─────┴─────┴─────┘
    │     │     │     │     │     │
    └─────┴─────┴─────┴─────┴─────┘
            Consumer-A (处理所有分区)
```

#### 扩容到3个Consumer：自动Rebalance
```
Topic: order-events (6 partitions)
┌─────┬─────┬─────┬─────┬─────┬─────┐
│ P0  │ P1  │ P2  │ P3  │ P4  │ P5  │
└─────┴─────┴─────┴─────┴─────┴─────┘
    │     │     │     │     │     │
    └─────┘     └─────┘     └─────┘
  Consumer-A   Consumer-B   Consumer-C
   (P0,P1)      (P2,P3)      (P4,P5)
```

#### Consumer-B故障：自动故障转移
```
Topic: order-events (6 partitions)
┌─────┬─────┬─────┬─────┬─────┬─────┐
│ P0  │ P1  │ P2  │ P3  │ P4  │ P5  │
└─────┴─────┴─────┴─────┴─────┴─────┘
    │     │     │     │     │     │
    └─────┴─────┘     └─────┴─────┘
     Consumer-A      Consumer-C
    (P0,P1,P2)       (P3,P4,P5)
```

### 场景2：多应用消费同一Topic

**背景**：`user-activity` Topic需要被多个业务消费

```
Topic: user-activity (4 partitions)
┌─────┬─────┬─────┬─────┐
│ P0  │ P1  │ P2  │ P3  │
└─────┴─────┴─────┴─────┘
    │     │     │     │
    └─────┴─────┘     │  Group: analytics-team
     Consumer-A1      │   └─ Consumer-A1: P0,P1
                      │   └─ Consumer-A2: P2,P3  
                      │
    ┌─────────────────┘
    │     │     │     │
    └─────┴─────┴─────┘  Group: recommendation-team
     Consumer-B1         └─ Consumer-B1: P0,P1,P2,P3
```

**关键点**：
- 不同Group之间完全独立
- 每个Group都会收到所有消息
- Group内部负载均衡，Group之间互不影响

---

## 🔧 技术架构设计

### 系统组件图

```
┌─────────────────┐    ┌─────────────────────────────┐
│   Consumer-1    │    │         Broker              │
│                 │    │  ┌───────────────────────┐   │
│ ┌─────────────┐ │    │  │  Group Coordinator    │   │
│ │GroupConsumer│◄┼────┼─►│                       │   │
│ │             │ │    │  │ ┌─── ConsumerGroup ─┐ │   │
│ └─────────────┘ │    │  │ │   - groupId       │ │   │
└─────────────────┘    │  │ │   - members[]     │ │   │
                       │  │ │   - assignment    │ │   │
┌─────────────────┐    │  │ │   - offsets       │ │   │
│   Consumer-2    │    │  │ └───────────────────┘ │   │
│                 │    │  └───────────────────────┘   │
│ ┌─────────────┐ │    │                             │
│ │GroupConsumer│◄┼────┤                             │
│ │             │ │    │  ┌───────────────────────┐   │
│ └─────────────┘ │    │  │    Topic Manager      │   │
└─────────────────┘    │  │  (existing)           │   │
                       │  └───────────────────────┘   │
                       └─────────────────────────────┘
```

### 数据结构设计

```go
// Group Coordinator：管理所有Consumer Group
type GroupCoordinator struct {
    groups map[string]*ConsumerGroup  // groupId -> group
    mutex  sync.RWMutex
}

// Consumer Group：单个消费组的信息
type ConsumerGroup struct {
    GroupId    string
    Members    map[string]*GroupMember    // consumerId -> member info
    Assignment map[string][]Assignment   // consumerId -> assigned partitions  
    Offsets    map[string]map[int32]int64 // topic -> partition -> committed offset
    Topics     []string                   // subscribed topics
    State      GroupState                 // Stable, Rebalancing, Dead
    mutex      sync.RWMutex
}

// Group Member：单个消费者的信息
type GroupMember struct {
    ConsumerId   string
    ClientId     string  
    LastSeen     time.Time
    Subscriptions []string  // 订阅的topic列表
}

// Assignment：分区分配信息
type Assignment struct {
    Topic       string
    PartitionId int32
}

// Group State：消费组状态
type GroupState int
const (
    StateStable GroupState = iota      // 稳定状态，正常消费
    StateRebalancing                   // 重平衡中
    StateDead                         // 组已解散
)
```

---

## 🔄 工作流程详解

### 1. Consumer加入Group流程

```sequence
Consumer->Broker: JoinGroupRequest
  {
    groupId: "order-processor",
    consumerId: "consumer-001", 
    topics: ["order-events"]
  }

Broker->Broker: GroupCoordinator.HandleJoin()
  1. 创建/获取ConsumerGroup
  2. 添加Member到group.Members
  3. 更新group.Topics (合并所有成员订阅)
  4. 标记group.State = Rebalancing

Broker->Consumer: JoinGroupResponse
  {
    success: true,
    memberId: "consumer-001",
    generation: 1
  }

Broker->Broker: 触发Rebalance
  1. 等待所有成员重新加入
  2. 执行分区分配算法
  3. 发送新的分配方案

Broker->Consumer: SyncGroupResponse
  {
    assignment: [
      {topic: "order-events", partitionId: 0},
      {topic: "order-events", partitionId: 1}
    ]
  }
```

### 2. Rebalance算法（Round Robin策略）

**输入**：
- Group成员：[consumer-A, consumer-B, consumer-C]
- 订阅Topics：["order-events"] 
- order-events分区：[P0, P1, P2, P3, P4, P5]

**算法步骤**：
```
1. 收集所有分区：[P0, P1, P2, P3, P4, P5]
2. 按Member数量平均分配：
   - 总分区数：6
   - Consumer数量：3  
   - 每个Consumer分配：6/3 = 2个分区
   
3. 依次分配：
   - consumer-A: [P0, P3]  (索引 0, 3)
   - consumer-B: [P1, P4]  (索引 1, 4) 
   - consumer-C: [P2, P5]  (索引 2, 5)
```

**伪代码实现**：
```go
func (gc *GroupCoordinator) assignPartitionsRoundRobin(
    members []string, 
    topicPartitions []TopicPartition,
) map[string][]Assignment {
    assignment := make(map[string][]Assignment)
    
    // 初始化每个成员的分配列表
    for _, member := range members {
        assignment[member] = make([]Assignment, 0)
    }
    
    // 轮询分配分区
    for i, partition := range topicPartitions {
        memberIndex := i % len(members)
        member := members[memberIndex]
        assignment[member] = append(assignment[member], Assignment{
            Topic:       partition.Topic,
            PartitionId: partition.PartitionId,
        })
    }
    
    return assignment
}
```

### 3. Offset管理流程

#### Consumer提交Offset
```go
// Consumer端
consumer.CommitOffset("order-events", 2, 1500) 

// 发送请求到Broker
CommitOffsetRequest{
    groupId: "order-processor",
    consumerId: "consumer-001", 
    offsets: [
        {topic: "order-events", partitionId: 2, offset: 1500}
    ]
}
```

#### Broker端处理
```go
func (gc *GroupCoordinator) HandleCommitOffset(req *CommitOffsetRequest) {
    group := gc.groups[req.GroupId]
    
    // 验证Consumer是否拥有这个分区
    if !gc.isAssignedToConsumer(group, req.ConsumerId, req.Topic, req.PartitionId) {
        return errors.New("partition not assigned to this consumer")
    }
    
    // 更新offset
    if group.Offsets[req.Topic] == nil {
        group.Offsets[req.Topic] = make(map[int32]int64)
    }
    group.Offsets[req.Topic][req.PartitionId] = req.Offset
}
```

#### Consumer获取起始Offset
```go
// Consumer启动时查询上次提交的offset
GetOffsetRequest{
    groupId: "order-processor", 
    topic: "order-events",
    partitionId: 2
}

// Broker响应
GetOffsetResponse{
    offset: 1500  // 从这里继续消费
}
```

---

## 🛡️ 错误处理和边界情况

### 1. Consumer故障检测

**心跳机制**：
```go
// Consumer定期发送心跳
type HeartbeatRequest struct {
    GroupId    string
    ConsumerId string
    Generation int32
}

// Broker检测超时
func (gc *GroupCoordinator) detectDeadMembers() {
    for groupId, group := range gc.groups {
        for consumerId, member := range group.Members {
            if time.Since(member.LastSeen) > 30*time.Second {
                // 踢出超时成员，触发Rebalance
                gc.removeMember(groupId, consumerId)
                gc.triggerRebalance(groupId)
            }
        }
    }
}
```

### 2. 重复消费预防

**场景**：Rebalance过程中Consumer还在处理消息
```
时间轴：
T1: Consumer-A处理分区P1的消息（offset 100-105）
T2: Rebalance开始，P1分配给Consumer-B  
T3: Consumer-A还在处理，Consumer-B开始从offset 100消费
T4: 重复处理！
```

**解决方案**：Generation机制
```go
type ConsumeRequest struct {
    // ... 其他字段
    Generation int32  // 消费时的generation版本
}

// Broker验证
if req.Generation != group.Generation {
    return errors.New("stale generation, please rejoin group")
}
```

### 3. Split-brain预防

**问题**：网络分区导致同一Consumer在多个Broker上注册

**解决方案**：
- Consumer只能连接一个Broker
- 使用唯一的ConsumerId
- Session超时机制

---

## 📊 性能考虑

### 1. Rebalance性能优化

**问题**：频繁Rebalance影响性能

**优化策略**：
```go
// 批量处理加入请求
type RebalanceCoordinator struct {
    pendingJoins    []JoinRequest
    rebalanceTimer  *time.Timer
    batchWindow     time.Duration  // 5秒窗口期
}

func (rc *RebalanceCoordinator) scheduleRebalance() {
    // 延迟触发，收集批量变更
    rc.rebalanceTimer = time.AfterFunc(rc.batchWindow, rc.doRebalance)
}
```

### 2. Offset存储优化

**当前方案**：内存存储
```go
// 内存中的offset map
group.Offsets[topic][partition] = offset
```

**未来优化**：
- 定期持久化到磁盘
- 使用内部Topic `__consumer_offsets` 存储
- 压缩重复的offset记录

---

## 🔮 实现计划

### Phase 1：基础框架
- [ ] GroupCoordinator数据结构
- [ ] Consumer Group管理
- [ ] 基础的Join/Leave Group协议

### Phase 2：分区分配
- [ ] Round Robin分配算法
- [ ] Rebalance流程实现
- [ ] 分配结果验证

### Phase 3：Offset管理
- [ ] Offset提交和查询
- [ ] Consumer启动时的Offset恢复
- [ ] 权限验证（Consumer只能提交自己分区的Offset）

### Phase 4：心跳和故障检测
- [ ] 心跳协议实现
- [ ] 超时检测和自动踢出
- [ ] Generation机制防止重复消费

### Phase 5：测试验证
- [ ] 单Consumer场景测试
- [ ] 多Consumer负载均衡测试  
- [ ] 故障转移测试
- [ ] 边界情况测试

---

你觉得这个设计怎么样？有哪些地方需要我进一步解释的？我们可以从Phase 1开始，一步步实现。