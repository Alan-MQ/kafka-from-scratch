package server

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/kafka-from-scratch/internal/broker"
	"github.com/kafka-from-scratch/internal/common"
	"github.com/kafka-from-scratch/internal/protocol"
)

// TCPServer TCP服务器，负责处理网络连接和请求
type TCPServer struct {
	address string
	broker  *broker.MemoryBroker

	listener net.Listener
	clients  map[string]net.Conn
	mu       sync.RWMutex
}

// NewTCPServer 创建新的TCP服务器
func NewTCPServer(address string, broker *broker.MemoryBroker) *TCPServer {
	return &TCPServer{
		address: address,
		broker:  broker,
		clients: make(map[string]net.Conn),
	}
}

// TODO: 你来实现这个方法！
// 功能：启动TCP服务器监听
// 提示：
// 1. 使用 net.Listen("tcp", s.address) 创建监听器
// 2. 启动一个 goroutine 处理新连接
// 3. 每个新连接启动一个 handleConnection goroutine
func (s *TCPServer) Start() error {
	// TODO: 实现服务器启动逻辑
	listener, err := net.Listen("tcp", s.address)
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}
	s.listener = listener
	fmt.Printf("tcp server listening %s\n", s.address)
	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}

		go s.handleConnection(conn)
	}
}

// TODO: 你来实现这个方法！
// 功能：处理客户端连接
// 提示：
// 1. 循环读取客户端请求
// 2. 解析JSON请求到protocol.Request
// 3. 调用handleRequest处理请求
// 4. 将响应发送回客户端
// 5. 处理连接错误和关闭
func (s *TCPServer) handleConnection(conn net.Conn) {
	// TODO: 实现连接处理逻辑
	defer conn.Close()

	clientAddr := conn.RemoteAddr().String()
	fmt.Printf("client addr %s\n", clientAddr)

	decoder := json.NewDecoder(conn)
	encoder := json.NewEncoder(conn)

	for {
		var request protocol.Request
		if err := decoder.Decode(&request); err != nil {
			fmt.Printf("decode error %v", err)
			break
		}
		fmt.Printf("recieved %s (ID: %s)\n", request.Type, request.RequestID)

		response := s.handleRequest(&request)

		if err := encoder.Encode(response); err != nil {
			fmt.Printf("reply error %v\n", err)
			break
		}
	}
	fmt.Printf("client disconnected %s\n", clientAddr)

}

// TODO: 你来实现这个方法！
// 功能：根据请求类型分发处理
// 提示：
// 1. 根据request.Type进行switch分发
// 2. 解析request.Data到具体的请求类型
// 3. 调用对应的处理方法
// 4. 返回protocol.Response
func (s *TCPServer) handleRequest(request *protocol.Request) *protocol.Response {
	// TODO: 实现请求处理分发逻辑

	switch request.Type {
	case protocol.RequestTypeCreateTopic:
		return s.handleCreateTopic(request)
	case protocol.RequestTypeProduce:
		return s.handleProduce(request)
	case protocol.RequestTypeConsume:
		return s.handleConsume(request)
	case protocol.RequestTypeSubscribe:
		return s.handleSubscribe(request)
	case protocol.RequestTypeSeek:
		return s.handleSeek(request)
	default:
		return s.createErrorResponse(request.RequestID, 
			fmt.Errorf("unknown request type: %s", request.Type))
	}
}

func (s *TCPServer) handleSeek(request *protocol.Request) *protocol.Response {
	// Seek操作的处理：实际上是为客户端的下一次Consume设置起始位置
	// 服务器端只需要确认offset的有效性，真正的Seek逻辑在客户端Consumer中
	
	// 解析请求数据
	reqData, _ := json.Marshal(request.Data)
	var seekReq protocol.SeekRequest
	json.Unmarshal(reqData, &seekReq)
	
	// 验证Topic和分区是否存在
	topic, err := s.broker.GetTopic(seekReq.Topic)
	if err != nil {
		return s.createErrorResponse(request.RequestID, 
			fmt.Errorf("topic not found: %s", seekReq.Topic))
	}
	
	_, err = topic.GetPartition(seekReq.PartitionId)
	if err != nil {
		return s.createErrorResponse(request.RequestID, err)
	}
	
	// 返回成功响应，表示offset设置请求已确认
	return s.createSuccessResponse(request.RequestID, &protocol.SeekResponse{
		Result: 0,
	})
}
func (s *TCPServer) handleSubscribe(request *protocol.Request) *protocol.Response {
	// Subscribe操作的处理：验证Topic是否存在
	reqData, _ := json.Marshal(request.Data)
	var subReq protocol.SubscribeRequest
	json.Unmarshal(reqData, &subReq)
	
	// 验证所有Topic是否存在
	for _, topicName := range subReq.Topics {
		_, err := s.broker.GetTopic(topicName)
		if err != nil {
			return s.createErrorResponse(request.RequestID, 
				fmt.Errorf("topic not found: %s", topicName))
		}
	}
	
	// 返回成功响应
	return s.createSuccessResponse(request.RequestID, &protocol.SubscribeResponse{
		Result: 0,
	})
}

func (s *TCPServer) handleConsume(request *protocol.Request) *protocol.Response {
	reqData, _ := json.Marshal(request.Data)
	var data protocol.ConsumeRequest
	json.Unmarshal(reqData, &data)
	
	messages, err := s.broker.ConsumeMessages(data.TopicName, data.PartitionId, data.Offset, data.MaxMessages)
	if err != nil {
		return s.createErrorResponse(request.RequestID, err)
	}
	
	// 转换为NetworkMessage格式
	networkMessages := make([]*protocol.NetworkMessage, len(messages))
	for i, msg := range messages {
		networkMessages[i] = &protocol.NetworkMessage{
			Key:       string(msg.Key),
			Value:     string(msg.Value),
			Headers:   msg.Headers,
			Offset:    msg.Offset,
			Timestamp: msg.Timestamp.Format(time.RFC3339),
		}
	}
	
	return s.createSuccessResponse(request.RequestID, &protocol.ConsumeResponse{
		Messages: networkMessages,
		Result:   0,
	})
}
func (s *TCPServer) handleProduce(request *protocol.Request) *protocol.Response {
	reqData, _ := json.Marshal(request.Data)
	var data protocol.ProduceRequest
	json.Unmarshal(reqData, &data)
	
	message := &common.Message{
		Key:       []byte(data.PartitionKey),
		Value:     []byte(data.Value),
		Headers:   data.Headers,
		Timestamp: time.Now(),
	}
	
	partitionID, offset, err := s.broker.ProduceMessage(data.TopicName, message)
	if err != nil {
		return s.createErrorResponse(request.RequestID, err)
	}
	
	return s.createSuccessResponse(request.RequestID, &protocol.ProduceResponse{
		PartitionId: partitionID,
		Offset:      offset,
		Result:      0,
	})
}

func (s *TCPServer) handleCreateTopic(request *protocol.Request) *protocol.Response {
	reqData, _ := json.Marshal(request.Data)
	var data protocol.CreateTopicRequest
	json.Unmarshal(reqData, &data)
	
	err := s.broker.CreateTopic(data.TopicName, data.PartitionNum)
	if err != nil {
		return s.createErrorResponse(request.RequestID, err)
	}
	
	return s.createSuccessResponse(request.RequestID, &protocol.CreateTopicResponse{
		Result: 0,
	})
}

// Stop 停止服务器
func (s *TCPServer) Stop() error {
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

// 辅助方法：创建成功响应
func (s *TCPServer) createSuccessResponse(requestID string, data interface{}) *protocol.Response {
	return &protocol.Response{
		RequestID: requestID,
		Success:   true,
		Data:      data,
	}
}

// 辅助方法：创建错误响应
func (s *TCPServer) createErrorResponse(requestID string, err error) *protocol.Response {
	return &protocol.Response{
		RequestID: requestID,
		Success:   false,
		Error:     err.Error(),
	}
}
