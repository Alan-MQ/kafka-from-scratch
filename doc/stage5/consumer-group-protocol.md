# Consumer Group 协议定义

## 🔌 新增请求类型

```go
// internal/protocol/messages.go 中新增
const (
    // 现有类型...
    RequestTypeJoinGroup      RequestType = "JOIN_GROUP"
    RequestTypeLeaveGroup     RequestType = "LEAVE_GROUP" 
    RequestTypeSyncGroup      RequestType = "SYNC_GROUP"
    RequestTypeHeartbeat      RequestType = "HEARTBEAT"
    RequestTypeCommitOffset   RequestType = "COMMIT_OFFSET"
    RequestTypeGetOffset      RequestType = "GET_OFFSET"
)
```

## 📤 请求数据结构

### JoinGroupRequest
```go
type JoinGroupRequest struct {
    GroupId         string   `json:"group_id"`
    ConsumerId      string   `json:"consumer_id"`
    ClientId        string   `json:"client_id"`
    Topics          []string `json:"topics"`
    SessionTimeout  int32    `json:"session_timeout"` // 毫秒
}
```

**使用场景**：Consumer启动时加入Group

**例子**：
```json
{
    "group_id": "order-processor",
    "consumer_id": "consumer-001",
    "client_id": "my-app-v1.0",
    "topics": ["order-events", "payment-events"],
    "session_timeout": 30000
}
```

---

### LeaveGroupRequest
```go
type LeaveGroupRequest struct {
    GroupId    string `json:"group_id"`
    ConsumerId string `json:"consumer_id"`
}
```

**使用场景**：Consumer优雅退出

---

### SyncGroupRequest
```go
type SyncGroupRequest struct {
    GroupId    string `json:"group_id"`
    ConsumerId string `json:"consumer_id"`
    Generation int32  `json:"generation"`
}
```

**使用场景**：Rebalance后获取分区分配

---

### HeartbeatRequest
```go
type HeartbeatRequest struct {
    GroupId    string `json:"group_id"`
    ConsumerId string `json:"consumer_id"`
    Generation int32  `json:"generation"`
}
```

**使用场景**：定期发送心跳保持活跃

---

### CommitOffsetRequest
```go
type CommitOffsetRequest struct {
    GroupId    string                    `json:"group_id"`
    ConsumerId string                    `json:"consumer_id"`
    Generation int32                     `json:"generation"`
    Offsets    []TopicPartitionOffset   `json:"offsets"`
}

type TopicPartitionOffset struct {
    Topic       string `json:"topic"`
    PartitionId int32  `json:"partition_id"`
    Offset      int64  `json:"offset"`
}
```

**使用场景**：提交消费进度

**例子**：
```json
{
    "group_id": "order-processor",
    "consumer_id": "consumer-001", 
    "generation": 5,
    "offsets": [
        {"topic": "order-events", "partition_id": 0, "offset": 1250},
        {"topic": "order-events", "partition_id": 1, "offset": 890}
    ]
}
```

---

### GetOffsetRequest
```go
type GetOffsetRequest struct {
    GroupId     string `json:"group_id"`
    Topic       string `json:"topic"`
    PartitionId int32  `json:"partition_id"`
}
```

**使用场景**：Consumer启动时查询上次的消费位置

---

## 📥 响应数据结构

### JoinGroupResponse
```go
type JoinGroupResponse struct {
    Success      bool     `json:"success"`
    Error        string   `json:"error,omitempty"`
    ConsumerId   string   `json:"consumer_id"`
    Generation   int32    `json:"generation"`
    LeaderId     string   `json:"leader_id"`      // Group Leader的ID
    Members      []string `json:"members"`        // 所有成员ID列表
}
```

**说明**：
- `LeaderId`：Group中选出的Leader，负责分配分区
- `Generation`：组的版本号，用于检测过期请求

---

### SyncGroupResponse  
```go
type SyncGroupResponse struct {
    Success    bool         `json:"success"`
    Error      string       `json:"error,omitempty"`
    Assignment []Assignment `json:"assignment"`
}

type Assignment struct {
    Topic       string `json:"topic"`
    PartitionId int32  `json:"partition_id"`
}
```

**例子**：
```json
{
    "success": true,
    "assignment": [
        {"topic": "order-events", "partition_id": 0},
        {"topic": "order-events", "partition_id": 3},
        {"topic": "payment-events", "partition_id": 1}
    ]
}
```

---

### HeartbeatResponse
```go
type HeartbeatResponse struct {
    Success           bool   `json:"success"`
    Error             string `json:"error,omitempty"`
    RebalanceRequired bool   `json:"rebalance_required"`
}
```

**说明**：
- `RebalanceRequired=true` 时，Consumer需要重新加入Group

---

### GetOffsetResponse
```go
type GetOffsetResponse struct {
    Success bool  `json:"success"`
    Error   string `json:"error,omitempty"`
    Offset  int64  `json:"offset"`
}
```

---

## 🔄 典型交互流程

### 场景1：Consumer首次加入
```
Consumer                          Broker
   │                                │
   │──── JoinGroupRequest ────────→ │
   │                                │ GroupCoordinator.HandleJoin()
   │                                │ - 创建/更新ConsumerGroup
   │                                │ - 添加Member
   │                                │ - 触发Rebalance
   │                                │
   │←─── JoinGroupResponse ─────────│ 
   │     {generation: 1}            │
   │                                │
   │──── SyncGroupRequest ─────────→│
   │     {generation: 1}            │ - 执行分区分配算法
   │                                │ - 返回分配结果
   │←─── SyncGroupResponse ────────│
   │     {assignment: [P0,P1]}      │
   │                                │
   │──── GetOffsetRequest ─────────→│ 查询上次消费位置
   │←─── GetOffsetResponse ────────│ {offset: 0} (首次为0)
   │                                │
   │ 开始消费...                     │
```

### 场景2：定期心跳
```
Consumer                          Broker  
   │                                │
   │──── HeartbeatRequest ─────────→│
   │     {generation: 1}            │ - 更新Member.LastSeen
   │                                │ - 检查是否需要Rebalance
   │←─── HeartbeatResponse ────────│
   │     {rebalance_required: false}│
   │                                │
   │ 继续消费...                     │
```

### 场景3：新Consumer加入触发Rebalance
```
原有Consumer                      Broker                新Consumer
      │                            │                         │
      │──── HeartbeatRequest ──────→│←──── JoinGroupRequest ──│
      │                            │ - 触发Rebalance         │
      │←─── HeartbeatResponse ─────│ - generation++          │
      │     {rebalance_required:   │                         │
      │      true}                 │                         │
      │                            │                         │
      │──── JoinGroupRequest ──────→│ 等待所有Member重新加入   │
      │                            │                         │
      │←─── JoinGroupResponse ─────│──── JoinGroupResponse →│
      │     {generation: 2}        │     {generation: 2}     │
      │                            │                         │
      │──── SyncGroupRequest ──────→│←──── SyncGroupRequest ──│
      │                            │ 重新分配分区             │
      │←─── SyncGroupResponse ─────│──── SyncGroupResponse →│
      │     {assignment: [P0]}     │     {assignment: [P1]}  │
```

---

## ⚠️ 错误码定义

```go
const (
    ErrorUnknownGroup        = "UNKNOWN_GROUP"
    ErrorInvalidGeneration   = "INVALID_GENERATION" 
    ErrorUnknownMember      = "UNKNOWN_MEMBER"
    ErrorRebalanceInProgress = "REBALANCE_IN_PROGRESS"
    ErrorInvalidAssignment   = "INVALID_ASSIGNMENT"
)
```

**使用示例**：
```json
{
    "success": false,
    "error": "INVALID_GENERATION: expected generation 3, got 2"
}
```

这样你对协议层面有更清晰的认识了吗？接下来我们开始实现基础框架！