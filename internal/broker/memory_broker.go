package broker

import (
	"errors"
	"sync"

	"github.com/kafka-from-scratch/internal/common"
)

// MemoryBroker 是我们第一阶段的内存版消息代理
type MemoryBroker struct {
	topics map[string]*common.Topic
	mu     sync.RWMutex
}

func NewMemoryBroker() *MemoryBroker {
	return &MemoryBroker{
		topics: make(map[string]*common.Topic),
	}
}

// TODO: 你来实现这个方法！
// 功能：创建一个新的Topic
// 提示：
// 1. 检查Topic是否已经存在
// 2. 如果不存在，创建新的Topic
// 3. 将Topic存储到broker的topics map中
// 4. 注意并发安全（使用读写锁）
func (b *MemoryBroker) CreateTopic(name string, partitions int32) error {
	// TODO: 在这里实现Topic创建逻辑
	b.mu.Lock()
	defer b.mu.Unlock()
	if _, exists := b.topics[name]; !exists {
		b.topics[name] = common.NewTopic(name, partitions)
	} else {
		// 疑问2， 如果已经有了怎么处理呢？
	}
	return nil
}

// TODO: 你来实现这个方法！
// 功能：获取指定名称的Topic
// 提示：
// 1. 从topics map中查找Topic
// 2. 如果不存在返回错误
// 3. 注意使用读锁保护
func (b *MemoryBroker) GetTopic(name string) (*common.Topic, error) {
	// TODO: 在这里实现Topic获取逻辑
	b.mu.RLock()
	defer b.mu.RUnlock()
	if t, exists := b.topics[name]; exists {
		return t, nil
	}
	// 是不是应该有个统一的错误处理
	return nil, errors.New("not found")
}

// TODO: 你来实现这个方法！
// 功能：向指定Topic发送消息
// 提示：
// 1. 先获取Topic
// 2. 调用Topic的ProduceMessage方法
// 3. 返回分区ID和offset
func (b *MemoryBroker) ProduceMessage(topicName string, message *common.Message) (int32, int64, error) {
	// TODO: 在这里实现消息生产逻辑
	b.mu.Lock()
	defer b.mu.Unlock()
	topic, ok := b.topics[topicName]
	if !ok {
		return 0, 0, errors.New("topic not found")
	}
	partition := topic.GetPartitionForKey(message.Key)
	partition.Append(message)
	return partition.ID, message.Offset, nil
}

// TODO: 你来实现这个方法！
// 功能：从指定Topic的分区消费消息
// 提示：
// 1. 获取Topic
// 2. 获取指定的分区
// 3. 调用分区的GetMessages方法
// 4. maxMessages参数控制一次最多获取多少条消息
func (b *MemoryBroker) ConsumeMessages(topicName string, partitionId int32, offset int64, maxMessages int) ([]*common.Message, error) {
	// TODO: 在这里实现消息消费逻辑
	b.mu.Lock()
	defer b.mu.Unlock()
	topic, ok := b.topics[topicName]
	if !ok {
		return nil, errors.New("topic not found")
	}
	partition, err := topic.GetPartition(partitionId)
	if err != nil {
		return nil, errors.New("partition not found")
	}
	// 疑问3， 我没理解， 这里只是获取了一份message ， 真正的消息还在 partition.Messages 里面， 怎么算消费了呢？
	return partition.GetMessages(offset, maxMessages)
}

// 这个方法我先给你实现，作为参考
func (b *MemoryBroker) ListTopics() []string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	topics := make([]string, 0, len(b.topics))
	for name := range b.topics {
		topics = append(topics, name)
	}
	return topics
}
