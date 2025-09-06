package producer

import (
	"github.com/kafka-from-scratch/internal/broker"
	"github.com/kafka-from-scratch/internal/common"
)

// Producer 接口定义了消息生产者的基本操作
type Producer interface {
	// Send 发送消息到指定Topic
	Send(topic string, message *common.Message) (partition int32, offset int64, err error)

	// SendWithKey 发送带Key的消息，用于分区选择
	SendWithKey(topic string, key, value []byte) (partition int32, offset int64, err error)

	// Close 关闭Producer并清理资源
	Close() error
}

// MemoryProducer 是基于内存Broker的Producer实现
type MemoryProducer struct {
	broker *broker.MemoryBroker
}

// NewMemoryProducer 创建一个新的内存版Producer
func NewMemoryProducer(b *broker.MemoryBroker) Producer {
	return &MemoryProducer{
		// 疑问1 我觉得这里不能拷贝锁， 应该直接使用指针， 我注意到你用了拷贝， 是有什么特别的考虑吗？
		broker: b,
	}
}

// TODO: 你来实现这个方法！
// 功能：发送消息到指定Topic
// 提示：
// 1. 直接调用broker的ProduceMessage方法
// 2. 返回分区ID、offset和错误
func (p *MemoryProducer) Send(topic string, message *common.Message) (int32, int64, error) {
	// TODO: 实现消息发送逻辑
	return p.broker.ProduceMessage(topic, message)
}

// TODO: 你来实现这个方法！
// 功能：发送带Key的消息
// 提示：
// 1. 使用common.NewMessage创建消息对象
// 2. 调用Send方法发送消息
func (p *MemoryProducer) SendWithKey(topic string, key, value []byte) (int32, int64, error) {
	// TODO: 实现带Key的消息发送
	return p.Send(topic, common.NewMessage(key, value))
}

// Close 关闭Producer (当前内存版本无需特殊处理)
func (p *MemoryProducer) Close() error {
	// 内存版本无需特殊的清理工作
	return nil
}

// ProducerConfig 配置结构 (为将来扩展预留)
type ProducerConfig struct {
	// RetryTimes 重试次数
	RetryTimes int

	// BatchSize 批量发送大小
	BatchSize int

	// TODO: 后续阶段会添加更多配置项
}
