package consumer

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/google/uuid"
	"github.com/kafka-from-scratch/internal/protocol"
)

// NetworkConsumer 网络版Consumer，通过TCP连接与Broker通信
type NetworkConsumer struct {
	brokerAddress string
	conn          net.Conn
	topics        []string                     // 已订阅的Topics
	offsets       map[string]map[int32]int64   // topic -> partition -> offset
}

// NewNetworkConsumer 创建网络版Consumer
func NewNetworkConsumer(brokerAddress string) *NetworkConsumer {
	return &NetworkConsumer{
		brokerAddress: brokerAddress,
		topics:        make([]string, 0),
		offsets:       make(map[string]map[int32]int64),
	}
}

// TODO: 你来实现这个方法！
// 功能：连接到Broker
// 提示：和Producer的Connect方法类似
func (nc *NetworkConsumer) Connect() error {
	// TODO: 实现连接逻辑
	conn, err := net.Dial("tcp", nc.brokerAddress)
	if err != nil {
		return err
	}
	nc.conn = conn
	return nil
}

// TODO: 你来实现这个方法！
// 功能：订阅Topic列表
// 提示：
// 1. 创建SubscribeRequest
// 2. 发送请求并处理响应
// 3. 更新本地的topics列表
func (nc *NetworkConsumer) Subscribe(topics []string) error {
	// TODO: 实现订阅逻辑
	subscribeReq := &protocol.SubscribeRequest{
		Topics: topics,
	}

	request := &protocol.Request{
		Type:      protocol.RequestTypeSubscribe,
		RequestID: uuid.New().String(),
		Data:      subscribeReq,
	}

	res, err := nc.sendRequest(request)
	if err != nil {
		return err
	}
	if !res.Success {
		return fmt.Errorf("subscribe failed: %s", res.Error)
	}
	
	// 更新本地topics列表并初始化offset
	nc.topics = topics
	for _, topic := range topics {
		if nc.offsets[topic] == nil {
			nc.offsets[topic] = make(map[int32]int64)
		}
	}
	return nil
}

// TODO: 你来实现这个方法！
// 功能：从指定Topic和分区消费消息
// 提示：
// 1. 创建ConsumeRequest
// 2. 发送请求并解析响应
// 3. 返回消息列表
func (nc *NetworkConsumer) Consume(topic string, partitionId int32, offset int64, maxMessages int) ([]*protocol.NetworkMessage, error) {
	// TODO: 实现消费逻辑
	consumeReq := &protocol.ConsumeRequest{
		TopicName:   topic,
		PartitionId: partitionId,
		Offset:      offset,
		MaxMessages: maxMessages,
	}

	request := &protocol.Request{
		Type:      protocol.RequestTypeConsume,
		RequestID: uuid.New().String(),
		Data:      consumeReq,
	}

	res, err := nc.sendRequest(request)
	if err != nil {
		return nil, err
	}
	if !res.Success {
		return nil, fmt.Errorf("consume failed: %s", res.Error)
	}
	
	// 解析响应
	respData, _ := json.Marshal(res.Data)
	var consumeResp protocol.ConsumeResponse
	json.Unmarshal(respData, &consumeResp)
	
	// 更新本地offset
	if nc.offsets[topic] == nil {
		nc.offsets[topic] = make(map[int32]int64)
	}
	if len(consumeResp.Messages) > 0 {
		// 设置为最后一条消息的offset + 1
		lastMsg := consumeResp.Messages[len(consumeResp.Messages)-1]
		nc.offsets[topic][partitionId] = lastMsg.Offset + 1
	}
	
	return consumeResp.Messages, nil
}

// TODO: 你来实现这个方法！
// 功能：设置消费位置（Seek操作）
// 提示：
// 1. 创建SeekRequest
// 2. 发送请求并处理响应
func (nc *NetworkConsumer) Seek(topic string, partitionId int32, offset int64) error {
	// 创建Seek请求
	seekReq := &protocol.SeekRequest{
		Topic:       topic,
		PartitionId: partitionId,
		Offset:      offset,
	}
	
	request := &protocol.Request{
		Type:      protocol.RequestTypeSeek,
		RequestID: uuid.New().String(),
		Data:      seekReq,
	}
	
	// 发送请求进行服务端验证
	res, err := nc.sendRequest(request)
	if err != nil {
		return err
	}
	if !res.Success {
		return fmt.Errorf("seek failed: %s", res.Error)
	}
	
	// 更新本地offset状态
	if nc.offsets[topic] == nil {
		nc.offsets[topic] = make(map[int32]int64)
	}
	nc.offsets[topic][partitionId] = offset
	
	return nil
}

// Close 关闭连接
func (nc *NetworkConsumer) Close() error {
	if nc.conn != nil {
		return nc.conn.Close()
	}
	return nil
}

// 辅助方法：发送请求并接收响应（和Producer类似）
func (nc *NetworkConsumer) sendRequest(req *protocol.Request) (*protocol.Response, error) {
	if nc.conn == nil {
		return nil, fmt.Errorf("not connected to broker")
	}

	// 发送请求
	encoder := json.NewEncoder(nc.conn)
	if err := encoder.Encode(req); err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// 接收响应
	var response protocol.Response
	decoder := json.NewDecoder(nc.conn)
	if err := decoder.Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to receive response: %w", err)
	}

	return &response, nil
}
