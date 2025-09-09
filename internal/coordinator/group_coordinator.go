package coordinator

import (
	"fmt"
	"sync"
	"time"

	"github.com/kafka-from-scratch/internal/broker"
	"github.com/kafka-from-scratch/internal/protocol"
)

// ==================== 核心数据结构 ====================

// GroupState Consumer Group的状态
type GroupState int

const (
	StateStable      GroupState = iota // 稳定状态，正常消费
	StateRebalancing                   // 重平衡中
	StateDead                          // 组已解散
)

const SessionTimeout = 30000 // 30000 ms = 30s
func (s GroupState) String() string {
	switch s {
	case StateStable:
		return "Stable"
	case StateRebalancing:
		return "Rebalancing"
	case StateDead:
		return "Dead"
	default:
		return "Unknown"
	}
}

// GroupMember Consumer Group中的成员信息
type GroupMember struct {
	ConsumerId     string
	ClientId       string
	Topics         []string  // 该成员订阅的Topics
	LastHeartbeat  time.Time // 最后一次心跳时间
	SessionTimeout int64     // 会话超时时间(毫秒)
}

// ConsumerGroup Consumer Group的完整信息
type ConsumerGroup struct {
	GroupId    string
	Members    map[string]*GroupMember          // consumerId -> member info
	Assignment map[string][]protocol.Assignment // consumerId -> assigned partitions
	Offsets    map[string]map[int32]int64       // topic -> partition -> committed offset
	Topics     []string                         // 合并所有成员订阅的topics
	State      GroupState
	Generation int32  // 组版本号
	LeaderId   string // 当前Leader成员的ID
	mutex      sync.RWMutex
}

// GroupCoordinator 管理所有Consumer Group的协调器
type GroupCoordinator struct {
	groups map[string]*ConsumerGroup // groupId -> group
	broker *broker.MemoryBroker      // 访问Topic和分区信息
	mutex  sync.RWMutex

	// 心跳检测
	heartbeatChecker *time.Ticker
	stopChan         chan struct{}
}

// ==================== 构造函数 ====================

// NewGroupCoordinator 创建新的Group Coordinator
func NewGroupCoordinator(broker *broker.MemoryBroker) *GroupCoordinator {
	gc := &GroupCoordinator{
		groups:   make(map[string]*ConsumerGroup),
		broker:   broker,
		stopChan: make(chan struct{}),
	}

	// 启动心跳检测
	gc.startHeartbeatChecker()

	return gc
}

// ==================== 🤔 考察题1: 心跳检测设计 ====================
// 问题: 心跳检测应该多久运行一次？太频繁 vs 太慢各有什么问题？
// 提示: 考虑网络延迟、系统负载、故障检测及时性的权衡

// startHeartbeatChecker 启动心跳检测goroutine
func (gc *GroupCoordinator) startHeartbeatChecker() {
	// TODO: 你来决定心跳检测的频率
	// 思考: 多久检测一次比较合适？为什么？
	// 查了下文档 我觉得设置为 sessionTimeout 设置为心跳检测的3倍吧 这样偶尔丢失一两条消息也能容忍
	// 所以我觉得30s ok ,你这里 Coordinator 心跳检测的频率 其实就是 sessionTimeout 对吗？?
	interval := time.Second * 30 // 先设个默认值，你可以调整

	gc.heartbeatChecker = time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-gc.heartbeatChecker.C:
				gc.checkDeadMembers()
			case <-gc.stopChan:
				return
			}
		}
	}()
}

// ==================== 主要业务方法 ====================

// HandleJoinGroup 处理Consumer加入Group的请求
func (gc *GroupCoordinator) HandleJoinGroup(req *protocol.JoinGroupRequest) (*protocol.JoinGroupResponse, error) {
	gc.mutex.Lock()
	defer gc.mutex.Unlock()

	// TODO: 你来实现JoinGroup的核心逻辑
	// 提示流程:
	// 1. 获取或创建ConsumerGroup
	// 2. 添加新成员到group.Members
	// 3. 更新group.Topics（合并所有成员的订阅）
	// 4. 检查是否需要触发Rebalance
	// 5. 返回响应

	fmt.Printf("📥 Handling JoinGroup: groupId=%s, consumerId=%s, topics=%v\n",
		req.GroupId, req.ConsumerId, req.Topics)

	// 获取或创建group
	group := gc.getOrCreateGroup(req.GroupId)

	// TODO: 实现具体逻辑
	resp := &protocol.JoinGroupResponse{
		ConsumerId: req.ConsumerId,
		LeaderId:   group.LeaderId,
		Generation: group.Generation,
	}

	members := make([]string, 0, len(group.Members))
	for memberId, _ := range group.Members {
		members = append(members, memberId)
	}
	resp.Members = members
	if _, exists := group.Members[req.ConsumerId]; exists {
		return resp, nil
	}
	// 如果不存在， 证明这是第一个进入ConsumerGroup的 Consumer， 那么他应该是Leader 也是member?
	resp.Members = append(resp.Members, req.ConsumerId)

	group.Members[req.ConsumerId] = &GroupMember{
		ConsumerId: req.ConsumerId,
		// ClientId: , // 啥事clientId ？ 不懂
		Topics:         make([]string, 0),
		LastHeartbeat:  time.Now(),
		SessionTimeout: SessionTimeout,
	}

	// 下面这个 操作真费劲， 为啥不直接用map 来管理Topics 这种array呢？
	reqTopicSet := make(map[string]struct{})
	needToAppend := make(map[string]struct{})
	for _, t := range req.Topics {
		reqTopicSet[t] = struct{}{}
	}
	for _, topic := range group.Topics {
		if _, exists := reqTopicSet[topic]; !exists {
			needToAppend[topic] = struct{}{}
		}
	}
	for topic, _ := range needToAppend {
		group.Topics = append(group.Topics, topic)
	}
	// 我们不做增量判断rebalance 吧， 只要加入consumer 就一定rebalance
	go gc.performRebalance(group)
	return resp, nil
}

// ==================== 🤔 考察题2: Rebalance触发时机 ====================
// 问题: 什么情况下需要触发Rebalance？
// A) 新成员加入
// B) 成员离开
// C) 成员订阅的Topic发生变化
// D) Topic的分区数量变化
// E) 以上全部
//
// 进阶问题: 如果一个空闲的Consumer离开，还需要Rebalance吗？为什么？

// HandleSyncGroup 处理获取分区分配的请求
func (gc *GroupCoordinator) HandleSyncGroup(req *protocol.SyncGroupRequest) (*protocol.SyncGroupResponse, error) {
	gc.mutex.RLock()
	defer gc.mutex.RUnlock()

	group, exists := gc.groups[req.GroupId]
	if !exists {
		return nil, fmt.Errorf("unknown group: %s", req.GroupId)
	}

	// TODO: 你来实现SyncGroup的逻辑
	// 提示流程:
	// 1. 验证Generation是否有效
	// 2. 检查group状态是否为Stable
	// 3. 返回该Consumer的分区分配
	// 4. 如果group正在Rebalancing，可能需要等待或返回错误
	// 疑问， 这里这个函数不是consumer 要求handle 这个partition 对吗？ 事实上consumer 根本不应该这样做
	// 应该等待分配， 所以我理解这里只是consumer 想同步信息， 确认这个分区的Assignment 情况

	fmt.Printf("📤 Handling SyncGroup: groupId=%s, consumerId=%s, generation=%d\n",
		req.GroupId, req.ConsumerId, req.Generation)

	resp := &protocol.SyncGroupResponse{
		Assignment: make([]protocol.Assignment, 0),
	}
	// TODO: 实现具体逻辑
	if req.Generation != group.Generation {
		return nil, fmt.Errorf("please rejoin the group")
	}
	if group.State != StateStable {
		// return resp, nil
		return nil, fmt.Errorf("rebalanceing please wait")
	}
	resp.Assignment = group.Assignment[req.ConsumerId]
	return resp, nil
}

// ==================== 🤔 考察题3: Generation机制设计 ====================
// 问题: 为什么需要Generation机制？如果没有Generation会发生什么？
// 场景: Consumer-A正在处理消息，此时发生Rebalance，Consumer-A的分区被分配给了Consumer-B
//       Consumer-A处理完消息后提交Offset，这时候会发生什么？

// HandleHeartbeat 处理心跳请求
func (gc *GroupCoordinator) HandleHeartbeat(req *protocol.HeartbeatRequest) (*protocol.HeartbeatResponse, error) {
	gc.mutex.Lock()
	defer gc.mutex.Unlock()

	// TODO: 你来实现Heartbeat处理逻辑
	// 提示流程:
	// 1. 验证Group和Member存在
	// 2. 验证Generation是否有效
	// 3. 更新Member的LastHeartbeat时间
	// 4. 检查是否需要通知Consumer进行Rebalance
	// 5. 返回HeartbeatResponse

	fmt.Printf("💗 Handling Heartbeat: groupId=%s, consumerId=%s, generation=%d\n",
		req.GroupId, req.ConsumerId, req.Generation)

	// TODO: 实现具体逻辑
	panic("TODO: 实现Heartbeat逻辑")
}

// ==================== 🤔 考察题4: Offset管理权限控制 ====================
// 问题: Consumer-A想提交分区P1的offset，但P1实际上分配给了Consumer-B，应该如何处理？
// A) 直接拒绝，返回错误
// B) 接受提交，覆盖现有offset
// C) 检查Generation，如果过期则拒绝
// D) A和C都对
//
// 进阶思考: 如果允许任意Consumer提交任意分区的offset，会有什么安全问题？

// HandleCommitOffset 处理offset提交请求
func (gc *GroupCoordinator) HandleCommitOffset(req *protocol.CommitOffsetRequest) (*protocol.CommitOffsetResponse, error) {
	gc.mutex.Lock()
	defer gc.mutex.Unlock()

	// TODO: 你来实现CommitOffset逻辑
	// 提示流程:
	// 1. 验证Group、Member、Generation
	// 2. 验证Consumer是否拥有要提交offset的分区
	// 3. 更新group.Offsets
	// 4. 返回确认响应

	fmt.Printf("📌 Handling CommitOffset: groupId=%s, consumerId=%s, offsets=%v\n",
		req.GroupId, req.ConsumerId, req.Offsets)

	// TODO: 实现具体逻辑
	panic("TODO: 实现CommitOffset逻辑")
}

// HandleGetOffset 处理获取offset请求
func (gc *GroupCoordinator) HandleGetOffset(req *protocol.GetOffsetRequest) (*protocol.GetOffsetResponse, error) {
	gc.mutex.RLock()
	defer gc.mutex.RUnlock()

	// TODO: 你来实现GetOffset逻辑
	// 提示: 从group.Offsets中查找对应的offset，如果不存在返回0

	fmt.Printf("📍 Handling GetOffset: groupId=%s, topic=%s, partition=%d\n",
		req.GroupId, req.Topic, req.PartitionId)

	// TODO: 实现具体逻辑
	panic("TODO: 实现GetOffset逻辑")
}

// ==================== 🤔 考察题5: Rebalance算法选择 ====================
// 问题: Round Robin算法的优缺点是什么？还有哪些分区分配算法？
// 场景: 3个Consumer，Topic-A有4个分区，Topic-B有2个分区
//       用Round Robin算法，Consumer1会分配到哪些分区？
//
// 算法挑战: 如何处理Consumer订阅不同Topic的情况？

// performRebalance 执行重平衡操作
func (gc *GroupCoordinator) performRebalance(group *ConsumerGroup) error {
	fmt.Printf("🔄 Starting rebalance for group %s, generation %d -> %d\n",
		group.GroupId, group.Generation, group.Generation+1)

	// TODO: 你来实现Rebalance算法
	// 提示流程:
	// 1. 收集所有需要分配的分区（来自group.Topics）
	// 2. 使用Round Robin算法分配分区给Members
	// 3. 更新group.Assignment
	// 4. 增加group.Generation
	// 5. 设置group.State = StateStable

	// 增加generation
	group.Generation++

	// TODO: 实现分区分配算法
	panic("TODO: 实现Rebalance算法")
}

// ==================== 辅助方法 ====================

// getOrCreateGroup 获取或创建Consumer Group
func (gc *GroupCoordinator) getOrCreateGroup(groupId string) *ConsumerGroup {
	group, exists := gc.groups[groupId]
	if !exists {
		group = &ConsumerGroup{
			GroupId:    groupId,
			Members:    make(map[string]*GroupMember),
			Assignment: make(map[string][]protocol.Assignment),
			Offsets:    make(map[string]map[int32]int64),
			Topics:     make([]string, 0),
			State:      StateStable,
			Generation: 0,
		}
		gc.groups[groupId] = group
		fmt.Printf("✨ Created new consumer group: %s\n", groupId)
	}
	return group
}

// 🤔 考察题: 这个方法的作用是什么？为什么Rebalance算法需要它？
// getTopicPartitions 获取指定topic的所有分区信息
func (gc *GroupCoordinator) getTopicPartitions(topicName string) ([]protocol.Assignment, error) {
	// TODO: 你来实现这个方法
	// 提示:
	// 1. 通过gc.broker.GetTopic(topicName)获取topic
	// 2. 遍历topic的所有分区
	// 3. 创建Assignment列表返回
	// 4. 如果topic不存在，返回适当的错误

	fmt.Printf("🔍 Getting partitions for topic: %s\n", topicName)

	// TODO: 实现获取分区信息的逻辑
	panic("TODO: 实现getTopicPartitions方法")
}

// checkDeadMembers 检查超时的成员
func (gc *GroupCoordinator) checkDeadMembers() {
	gc.mutex.Lock()
	defer gc.mutex.Unlock()

	// TODO: 你来实现超时检测逻辑
	// 提示:
	// 1. 遍历所有group的所有member
	// 2. 检查LastHeartbeat是否超过SessionTimeout
	// 3. 移除超时成员并触发Rebalance

	fmt.Println("⏰ Checking for dead members...")
	for groupId, group := range gc.groups {
		needRebalance := false
		for _, member := range group.Members {
			if time.Now().Before(member.LastHeartbeat.Add(time.Duration(member.SessionTimeout) * time.Millisecond)) {
				continue
			}
			needRebalance = true
			// 怎么移除member呢？ 直接写成nil？
			gc.groups[groupId] = nil
		}
		if needRebalance {
			gc.performRebalance(group)
		}
	}

}

// Stop 停止GroupCoordinator
func (gc *GroupCoordinator) Stop() {
	if gc.heartbeatChecker != nil {
		gc.heartbeatChecker.Stop()
	}
	close(gc.stopChan)
	fmt.Println("🛑 GroupCoordinator stopped")
}

// ==================== 🏆 奖励题: 边界情况处理 ====================
//
// 1. 如果Consumer的SessionTimeout设置为0，应该如何处理？
// 2. 如果同一个ConsumerId重复JoinGroup，应该如何处理？
// 3. 如果Group中所有Consumer都离开了，Group应该如何清理？
// 4. 如果Rebalance过程中有新的Consumer加入，应该如何处理？
// 5. 如果Topic不存在，Consumer Group应该如何处理？
//
// 思考这些问题，并在实现中考虑这些边界情况的处理！
