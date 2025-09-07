package protocol

// TODO: 你来实现这个文件！
// 定义所有的请求数据结构

// CreateTopicRequest 创建Topic请求
type CreateTopicRequest struct {
	// TODO: 你来定义字段
	// 提示: 需要topic名称和分区数
	TopicName    string `json:"topic_name"`
	PartitionNum int32  `json:"partition_num"`
}

// ProduceRequest 生产消息请求
type ProduceRequest struct {
	// TODO: 你来定义字段
	// 提示: 需要topic名称、消息的key、value、headers

	TopicName    string            `json:"topic_name"`
	PartitionKey string            `json:"partition_key"`
	Value        string            `json:"value"`
	Headers      map[string]string `json:"headers"`
}

// ConsumeRequest 消费消息请求
type ConsumeRequest struct {
	// TODO: 你来定义字段
	// 提示: 需要topic名称、分区ID、起始offset、最大消息数
	// 疑问， 这里为什么是直接用partition_id， 所以consumer 可以从任意分区消费吗？ consumer 传错了怎么办
	// 不应该是kafka 控制哪个消费者可以从哪些地方consume吗？ 这个 ConsumeRequest 看起来像是可以从 指定的PartitionId 消费
	// 还是说我的理解有问题？ 真实的kafka 也是这样的吗？
	TopicName     string `json:"topic_name"`
	PartitionId   int32  `json:"partition_id"`
	Offset        int64  `json:"offset"`
	MaxMessages   int    `json:"max_messages"`
	ConsumerGroup string `json:"consumer_group"` // 预留字段 阶段5 在用
}

// SubscribeRequest 订阅请求
type SubscribeRequest struct {
	// TODO: 你来定义字段
	// 提示: 需要topic列表
	Topics []string `json:"topics"`
}

// SeekRequest Seek操作请求
type SeekRequest struct {
	// TODO: 你来定义字段
	// 提示: 需要topic名称、分区ID、目标offset
	Topic       string `json:"topic"`
	PartitionId int32  `json:"partition_id"`
	Offset      int64  `json:"offset"`
}
