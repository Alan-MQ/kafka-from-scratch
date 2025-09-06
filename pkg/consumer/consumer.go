package consumer

import (
	"fmt"

	"github.com/kafka-from-scratch/internal/broker"
	"github.com/kafka-from-scratch/internal/common"
)

// Consumer 接口定义了消息消费者的基本操作
type Consumer interface {
	// Subscribe 订阅一个或多个Topic
	Subscribe(topics []string) error

	// Poll 拉取消息，返回消息列表
	Poll(maxMessages int) ([]*common.Message, error)

	// Seek 设置消费位置到指定offset
	Seek(topic string, partition int32, offset int64) error

	// Close 关闭Consumer并清理资源
	Close() error
}

// MemoryConsumer 是基于内存Broker的Consumer实现
type MemoryConsumer struct {
	broker           *broker.MemoryBroker
	subscribedTopics []string
	// 当前消费位置 - map[topic][partition]offset
	offsets map[string]map[int32]int64
}

// NewMemoryConsumer 创建一个新的内存版Consumer
func NewMemoryConsumer(b *broker.MemoryBroker) Consumer {
	return &MemoryConsumer{
		broker:           b,
		subscribedTopics: make([]string, 0),
		offsets:          make(map[string]map[int32]int64),
	}
}

// TODO: 你来实现这个方法！
// 功能：订阅Topic列表
// 提示：
// 1. 保存订阅的Topic列表到subscribedTopics
// 2. 为每个Topic初始化offset跟踪 (从0开始)
// 3. 需要获取每个Topic的分区数量来初始化offset map
func (c *MemoryConsumer) Subscribe(topics []string) error {
	// 保存新的订阅列表 (替换语义，与Kafka保持一致)
	c.subscribedTopics = topics
	
	// 只为新Topic初始化offset，保留已有的offset
	for _, topic := range topics {
		if c.offsets[topic] == nil {  // 只初始化新Topic
			topicObj, err := c.broker.GetTopic(topic)
			if err != nil {
				return fmt.Errorf("topic %s not found: %w", topic, err)
			}
			
			// 初始化该topic的offset map
			c.offsets[topic] = make(map[int32]int64)
			
			// 为每个分区设置初始offset为0
			partitionCount := topicObj.GetPartitionCount()
			for i := int32(0); i < partitionCount; i++ {
				c.offsets[topic][i] = 0
			}
		}
	}
	
	return nil
}

// TODO: 你来实现这个方法！
// 功能：拉取消息
// 提示：
// 1. 遍历所有订阅的Topic和分区
// 2. 从当前offset位置消费消息
// 3. 更新offset位置
// 4. 返回所有拉取到的消息
// 5. maxMessages控制总的消息数量上限
func (c *MemoryConsumer) Poll(maxMessages int) ([]*common.Message, error) {
	var allMessages []*common.Message
	messagesCollected := 0

	// 遍历所有订阅的Topic
	for _, topicName := range c.subscribedTopics {
		if messagesCollected >= maxMessages {
			break
		}
		
		// 获取Topic对象
		topicObj, err := c.broker.GetTopic(topicName)
		if err != nil {
			continue  // 跳过不存在的Topic
		}
		
		// 遍历该Topic的所有分区 (当前策略：单Consumer消费所有分区)
		partitionCount := topicObj.GetPartitionCount()
		for partitionID := int32(0); partitionID < partitionCount; partitionID++ {
			if messagesCollected >= maxMessages {
				break
			}
			
			// 获取当前分区的offset
			currentOffset := c.offsets[topicName][partitionID]
			
			// 计算这个分区要拉取多少消息
			remainingMessages := maxMessages - messagesCollected
			
			// 从Broker拉取消息
			messages, err := c.broker.ConsumeMessages(topicName, partitionID, currentOffset, remainingMessages)
			if err != nil {
				continue  // 跳过有问题的分区
			}
			
			// 添加到结果集
			allMessages = append(allMessages, messages...)
			messagesCollected += len(messages)
			
			// 更新offset
			if len(messages) > 0 {
				lastMessage := messages[len(messages)-1]
				c.offsets[topicName][partitionID] = lastMessage.Offset + 1
			}
		}
	}
	
	return allMessages, nil
}

// TODO: 你来实现这个方法！
// 功能：设置指定分区的消费位置
// 提示：
// 1. 检查topic是否已订阅
// 2. 更新offsets map中对应的offset值
func (c *MemoryConsumer) Seek(topic string, partition int32, offset int64) error {
	// 检查topic是否已订阅
	subscribed := false
	for _, subscribedTopic := range c.subscribedTopics {
		if subscribedTopic == topic {
			subscribed = true
			break
		}
	}
	
	if !subscribed {
		return fmt.Errorf("topic %s is not subscribed", topic)
	}
	
	// 检查分区是否存在
	topicObj, err := c.broker.GetTopic(topic)
	if err != nil {
		return err
	}
	
	if partition < 0 || partition >= topicObj.GetPartitionCount() {
		return fmt.Errorf("partition %d does not exist in topic %s", partition, topic)
	}
	
	// 更新offset
	if c.offsets[topic] == nil {
		c.offsets[topic] = make(map[int32]int64)
	}
	c.offsets[topic][partition] = offset
	
	return nil
}

// Close 关闭Consumer
func (c *MemoryConsumer) Close() error {
	// 清理订阅信息
	c.subscribedTopics = nil
	c.offsets = nil
	return nil
}

// ConsumerConfig 配置结构 (为将来扩展预留)
type ConsumerConfig struct {
	// GroupID 消费者组ID (当前版本未使用)
	GroupID string

	// AutoCommit 是否自动提交offset
	AutoCommit bool

	// TODO: 后续阶段会添加更多配置项
}
