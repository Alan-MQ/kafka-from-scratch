package producer

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/google/uuid"
	"github.com/kafka-from-scratch/internal/protocol"
)

// NetworkProducer 网络版Producer，通过TCP连接与Broker通信
type NetworkProducer struct {
	brokerAddress string
	conn          net.Conn
}

// NewNetworkProducer 创建网络版Producer
func NewNetworkProducer(brokerAddress string) *NetworkProducer {
	return &NetworkProducer{
		brokerAddress: brokerAddress,
	}
}

// TODO: 你来实现这个方法！
// 功能：连接到Broker
// 提示：
// 1. 使用 net.Dial("tcp", np.brokerAddress) 建立连接
// 2. 保存连接到 np.conn
// 3. 处理连接错误
func (np *NetworkProducer) Connect() error {
	// TODO: 实现连接逻辑
	conn, err := net.Dial("tcp", np.brokerAddress)
	if err != nil {
		return err
	}
	np.conn = conn
	return nil
}

// TODO: 你来实现这个方法！
// 功能：向指定Topic发送消息
// 提示：
// 1. 创建ProduceRequest
// 2. 包装到protocol.Request中
// 3. JSON序列化并发送
// 4. 读取并解析响应
// 5. 返回分区ID和offset
func (np *NetworkProducer) Send(topic string, key, value []byte) (int32, int64, error) {
	// TODO: 实现消息发送逻辑

	// TopicName    string            `json:"topic_name"`
	// PartitionKey string            `json:"partition_key"`
	// Value        string            `json:"value"`
	// Headers      map[string]string `json:"headers"`
	// 1. 创建请求数据
	produceReq := &protocol.ProduceRequest{
		TopicName:    topic,
		PartitionKey: string(key),
		Value:        string(value),
		Headers:      make(map[string]string),
	}

	// 2. 包装到通用请求
	request := &protocol.Request{
		Type:      protocol.RequestTypeProduce,
		RequestID: uuid.New().String(),
		Data:      produceReq,
	}

	res, err := np.sendRequest(request)
	if err != nil {
		return 0, 0, err
	}

	if !res.Success {
		return 0, 0, fmt.Errorf("produce message failed: %s", res.Error)
	}

	respData, _ := json.Marshal(res.Data)
	var produceResp protocol.ProduceResponse
	json.Unmarshal(respData, &produceResp)
	return produceResp.PartitionId, produceResp.Offset, nil
}

// TODO: 你来实现这个方法！
// 功能：创建Topic
// 提示：
// 1. 创建CreateTopicRequest
// 2. 发送请求并处理响应
func (np *NetworkProducer) CreateTopic(name string, partitions int32) error {
	// TODO: 实现Topic创建逻辑
	sendReq := &protocol.CreateTopicRequest{
		TopicName:    name,
		PartitionNum: partitions,
	}

	request := &protocol.Request{
		Type:      protocol.RequestTypeCreateTopic,
		RequestID: uuid.New().String(),
		Data:      sendReq,
	}

	res, err := np.sendRequest(request)
	if err != nil {
		return err
	}

	if !res.Success {
		return fmt.Errorf("create topic err since %s", res.Error)
	}
	respData, _ := json.Marshal(res.Data)
	var createResp protocol.CreateTopicResponse
	json.Unmarshal(respData, &createResp)
	fmt.Printf("Result of create topic is %d\n", createResp.Result)
	return nil
}

// Close 关闭连接
func (np *NetworkProducer) Close() error {
	if np.conn != nil {
		return np.conn.Close()
	}
	return nil
}

// 辅助方法：发送请求并接收响应
func (np *NetworkProducer) sendRequest(req *protocol.Request) (*protocol.Response, error) {
	if np.conn == nil {
		return nil, fmt.Errorf("not connected to broker")
	}

	// 发送请求
	encoder := json.NewEncoder(np.conn)
	if err := encoder.Encode(req); err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// 接收响应
	var response protocol.Response
	decoder := json.NewDecoder(np.conn)
	if err := decoder.Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to receive response: %w", err)
	}

	return &response, nil
}
