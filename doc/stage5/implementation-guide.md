# Consumer Group 实现指南 🛠️

## 🎉 框架已就绪！

恭喜！Consumer Group的完整框架已经搭建完成。现在你有了：

### ✅ 已完成的组件：

1. **协议层定义** (`internal/protocol/`)
   - 6种Consumer Group请求类型
   - 完整的请求/响应数据结构
   - 与现有协议完美集成

2. **核心数据结构** (`internal/coordinator/group_coordinator.go`)
   - `GroupCoordinator`: 管理所有Consumer Group
   - `ConsumerGroup`: 单个Group的完整状态
   - `GroupMember`: Consumer成员信息
   - 状态机和Generation机制

3. **网络集成** (`internal/server/tcp_server.go`)
   - TCP服务器已集成GroupCoordinator
   - 6个Consumer Group请求处理器框架
   - 错误处理和响应机制

4. **学习资源**
   - 📚 深度设计文档 (3个)
   - 🧠 12道考察题 + 奖励题
   - 💡 大量实现提示和考察点

---

## 🎯 你的任务清单

### 第一优先级：核心功能实现

#### 1. 实现 `HandleJoinGroup` 方法 ⭐⭐⭐
**位置**: `group_coordinator.go:102`

**流程提示**:
```go
func (gc *GroupCoordinator) HandleJoinGroup(req *protocol.JoinGroupRequest) (*protocol.JoinGroupResponse, error) {
    // 1. 获取或创建ConsumerGroup
    // 2. 创建GroupMember并添加到group.Members
    // 3. 合并Topics到group.Topics
    // 4. 如果是新成员或Topic变化，触发Rebalance
    // 5. 选择Leader (第一个加入的成员)
    // 6. 返回JoinGroupResponse
}
```

**考察点**: 
- Q2: 什么时候需要触发Rebalance？
- Q7: 重复Consumer ID如何处理？

#### 2. 实现 `performRebalance` 方法 ⭐⭐⭐
**位置**: `group_coordinator.go:243`

**算法提示**:
```go
func (gc *GroupCoordinator) performRebalance(group *ConsumerGroup) error {
    // 1. 收集所有分区：遍历group.Topics，调用getTopicPartitions
    // 2. Round Robin分配：partitions[i] -> members[i % len(members)]
    // 3. 更新group.Assignment
    // 4. 增加group.Generation
    // 5. 设置group.State = StateStable
}
```

**考察点**: 
- Q4: Round Robin算法的具体实现
- Q5: 不同订阅Topic的处理

#### 3. 实现 `getTopicPartitions` 方法 ⭐⭐
**位置**: `group_coordinator.go:285`

**实现提示**:
```go
func (gc *GroupCoordinator) getTopicPartitions(topicName string) ([]protocol.Assignment, error) {
    topic, err := gc.broker.GetTopic(topicName)
    if err != nil {
        return nil, err
    }
    
    var assignments []protocol.Assignment
    for i := int32(0); i < topic.GetPartitionCount(); i++ {
        assignments = append(assignments, protocol.Assignment{
            Topic: topicName,
            PartitionId: i,
        })
    }
    return assignments, nil
}
```

### 第二优先级：状态管理

#### 4. 实现 `HandleSyncGroup` 方法 ⭐⭐
**位置**: `group_coordinator.go:130`

**验证逻辑**:
- 检查Generation有效性
- 返回Consumer的分区分配
- 处理Rebalancing状态

#### 5. 实现 `HandleHeartbeat` 方法 ⭐⭐
**位置**: `group_coordinator.go:142`

**核心功能**:
- 更新LastHeartbeat时间
- 检查是否需要Rebalance
- Generation验证

### 第三优先级：Offset管理

#### 6. 实现 `HandleCommitOffset` 方法 ⭐⭐
**位置**: `group_coordinator.go:158`

**权限控制**:
- 验证Consumer是否拥有该分区
- Generation检查
- 更新group.Offsets

#### 7. 实现 `HandleGetOffset` 方法 ⭐
**位置**: `group_coordinator.go:173`

**简单查询**:
- 从group.Offsets查找
- 不存在返回0

### 第四优先级：健壮性

#### 8. 实现 `checkDeadMembers` 方法 ⭐⭐
**位置**: `group_coordinator.go:299`

**超时检测**:
- 遍历所有Member
- 检查LastHeartbeat vs SessionTimeout
- 移除超时成员并触发Rebalance

#### 9. 实现TCP服务器处理器 ⭐
**位置**: `tcp_server.go:238-273`

**标准模式**:
```go
func (s *TCPServer) handleJoinGroup(request *protocol.Request) *protocol.Response {
    reqData, _ := json.Marshal(request.Data)
    var joinReq protocol.JoinGroupRequest
    json.Unmarshal(reqData, &joinReq)
    
    resp, err := s.groupCoordinator.HandleJoinGroup(&joinReq)
    if err != nil {
        return s.createErrorResponse(request.RequestID, err)
    }
    return s.createSuccessResponse(request.RequestID, resp)
}
```

---

## 🧠 学习策略建议

### 开始前：回答考察题
在写代码前，先花时间回答 `implementation-quiz.md` 中的考察题，特别是：
- Q1: 心跳检测频率 
- Q2: Rebalance触发时机
- Q3: Generation机制
- Q4: Round Robin算法
- Q5: Offset权限控制

### 实现顺序建议：
1. **先实现简单方法**: `getTopicPartitions` → `HandleGetOffset`
2. **核心流程**: `HandleJoinGroup` → `performRebalance` 
3. **状态管理**: `HandleSyncGroup` → `HandleHeartbeat`
4. **完善功能**: `HandleCommitOffset` → `checkDeadMembers`
5. **网络层**: TCP服务器处理器们
6. **端到端测试**: 完整Consumer Group流程

### 调试技巧：
- 多使用fmt.Printf观察状态变化
- 先单Consumer测试，再多Consumer测试
- 重点关注Generation和Assignment的正确性

---

## 🏆 成功标准

完成实现后，你应该能够：

### 功能验证 ✅
- [ ] Consumer可以成功JoinGroup
- [ ] 多Consumer自动分区分配
- [ ] Consumer故障时自动Rebalance  
- [ ] Offset正确提交和获取
- [ ] 心跳机制正常工作

### 边界情况处理 ✅
- [ ] 重复ConsumerId处理
- [ ] SessionTimeout验证
- [ ] Generation过期拒绝
- [ ] Topic不存在处理
- [ ] 空Group清理

### 性能表现 ✅
- [ ] 心跳检测不占用过多CPU
- [ ] Rebalance不会无限循环
- [ ] 内存使用合理，无明显泄漏

---

## 🚀 准备开始？

你现在拥有：
- ✨ 完整的架构框架
- 📖 详细的实现指南  
- 🧠 深度的概念理解
- 💡 充足的代码提示
- 🎯 明确的任务清单

**是时候把理论变成代码了！从第一个TODO开始，让Consumer Group活起来！** 

Good luck! 💪✨