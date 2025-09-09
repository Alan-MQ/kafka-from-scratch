package protocol

// TODO: 你来实现这个文件！
// 定义所有的响应数据结构

// CreateTopicResponse 创建Topic响应
type CreateTopicResponse struct {
	// TODO: 你来定义字段
	// 提示: 可能只需要确认信息
	Result int8 `json:"result"` // 0 表示没问题
}

// ProduceResponse 生产消息响应
type ProduceResponse struct {
	// TODO: 你来定义字段
	// 提示: 需要返回分区ID和offset

	PartitionId int32 `json:"partition_id"`
	Offset      int64 `json:"offset"`
	Result      int8  `json:"result"` // 0 表示没问题
}

// ConsumeResponse 消费消息响应
type ConsumeResponse struct {
	// TODO: 你来定义字段
	// 提示: 需要返回消息列表
	Messages []*NetworkMessage `json:"messages"`
	Result   int8              `json:"result"` // 0 表示没问题
}

type NetworkMessage struct {
	Key       string            `json:"key,omitempty"`
	Value     string            `json:"value"`
	Headers   map[string]string `json:"headers,omitempty"`
	Offset    int64             `json:"offset"`
	Timestamp string            `json:"timestamp"`
}

// SubscribeResponse 订阅响应
type SubscribeResponse struct {
	// TODO: 你来定义字段
	// 提示: 可能只需要确认信息
	Result int8 `json:"result"` // 0 表示没问题
}

// SeekResponse Seek操作响应
type SeekResponse struct {
	// TODO: 你来定义字段
	// 提示: 可能只需要确认信息
	Result int8 `json:"result"` // 0 表示没问题
}

// ==================== Consumer Group 协议响应 ====================

// JoinGroupResponse Consumer加入Group的响应
type JoinGroupResponse struct {
	ConsumerId string   `json:"consumer_id"`
	Generation int32    `json:"generation"`
	LeaderId   string   `json:"leader_id"`  // Group Leader的ID
	Members    []string `json:"members"`    // 所有成员ID列表
}

// LeaveGroupResponse Consumer离开Group的响应
type LeaveGroupResponse struct {
	// 简单确认即可
}

// SyncGroupResponse 分区分配响应
type SyncGroupResponse struct {
	Assignment []Assignment `json:"assignment"`
}

// Assignment 分区分配信息
type Assignment struct {
	Topic       string `json:"topic"`
	PartitionId int32  `json:"partition_id"`
}

// HeartbeatResponse 心跳响应
type HeartbeatResponse struct {
	RebalanceRequired bool `json:"rebalance_required"`
}

// CommitOffsetResponse 提交offset响应
type CommitOffsetResponse struct {
	// 简单确认即可
}

// GetOffsetResponse 获取offset响应
type GetOffsetResponse struct {
	Offset int64 `json:"offset"`
}
