package common

import (
	"time"
)

type Message struct {
	Key       []byte            
	Value     []byte            
	Headers   map[string]string 
	Timestamp time.Time         
	Offset    int64             
}

func NewMessage(key, value []byte) *Message {
	return &Message{
		Key:       key,
		Value:     value,
		Headers:   make(map[string]string),
		Timestamp: time.Now(),
		Offset:    -1, 
	}
}

func (m *Message) SetHeader(key, value string) {
	if m.Headers == nil {
		m.Headers = make(map[string]string)
	}
	m.Headers[key] = value
}

func (m *Message) GetHeader(key string) (string, bool) {
	if m.Headers == nil {
		return "", false
	}
	value, exists := m.Headers[key]
	return value, exists
}