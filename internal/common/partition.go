package common

import (
	"sync"
)

type Partition struct {
	ID       int32
	Messages []*Message
	mu       sync.RWMutex
}

func NewPartition(id int32) *Partition {
	return &Partition{
		ID:       id,
		Messages: make([]*Message, 0),
	}
}

func (p *Partition) Append(message *Message) int64 {
	p.mu.Lock()
	defer p.mu.Unlock()

	offset := int64(len(p.Messages))
	message.Offset = offset
	p.Messages = append(p.Messages, message)

	return offset
}

func (p *Partition) GetMessages(startOffset int64, maxMessages int) ([]*Message, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if startOffset < 0 || startOffset >= int64(len(p.Messages)) {
		return []*Message{}, nil
	}

	endOffset := startOffset + int64(maxMessages)
	if endOffset > int64(len(p.Messages)) {
		endOffset = int64(len(p.Messages))
	}

	messages := make([]*Message, endOffset-startOffset)
	// 疑问1 这里是 深拷贝的意思吗？ 意思是 GetMessages 希望获取从start 到 end 这一段的所有消息
	// 但是 希望后续获得这些消息的对象 如果修改了消息， 不会改到p.Messages 里面的内容？
	copy(messages, p.Messages[startOffset:endOffset])

	return messages, nil
}

func (p *Partition) GetLatestOffset() int64 {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return int64(len(p.Messages))
}
