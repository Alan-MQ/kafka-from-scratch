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
