package common

import (
	"fmt"
	"hash/maphash"
	"sync"
)

type Topic struct {
	Name       string
	Partitions []*Partition
	mu         sync.RWMutex
}

func NewTopic(name string, numPartitions int32) *Topic {
	if numPartitions <= 0 {
		numPartitions = 1
	}

	partitions := make([]*Partition, numPartitions)
	for i := int32(0); i < numPartitions; i++ {
		partitions[i] = NewPartition(i)
	}

	return &Topic{
		Name:       name,
		Partitions: partitions,
	}
}

func (t *Topic) GetPartition(partitionID int32) (*Partition, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if partitionID < 0 || partitionID >= int32(len(t.Partitions)) {
		return nil, fmt.Errorf("partition %d does not exist in topic %s", partitionID, t.Name)
	}

	return t.Partitions[partitionID], nil
}

// TODO: 你来实现这个方法！
// 功能：根据消息的Key选择合适的分区
// 提示：
// 1. 如果key为空，返回第一个分区
// 2. 使用hash函数计算key的哈希值
// 3. 用哈希值对分区数量取模，得到分区ID
// 4. 注意处理负数情况
func (t *Topic) GetPartitionForKey(key []byte) *Partition {
	// TODO: 在这里实现分区选择逻辑
	if len(key) <= 0 {
		return t.Partitions[0]
	}
	count := t.GetPartitionCount()
	var h maphash.Hash
	h.WriteString(string(key))
	return t.Partitions[int64(h.Sum64())%int64(count)]
}

func (t *Topic) GetPartitionCount() int32 {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return int32(len(t.Partitions))
}

// TODO: 你来实现这个方法！
// 功能：将消息发送到合适的分区
// 提示：
// 1. 使用GetPartitionForKey选择分区
// 2. 调用分区的Append方法添加消息
// 3. 返回分区ID和消息的offset
func (t *Topic) ProduceMessage(message *Message) (int32, int64, error) {
	// TODO: 在这里实现消息生产逻辑
	partition := t.GetPartitionForKey(message.Key)
	partition.Append(message)
	return partition.ID, message.Offset, nil
}
