# TCP服务器实现指导

## 🎯 你的任务
实现 `internal/server/tcp_server.go` 中的3个TODO方法。

## 📋 方法实现指导

### 1. `Start() error` 方法

**功能**: 启动TCP服务器并监听连接

**实现步骤**:
```go
func (s *TCPServer) Start() error {
    // 1. 创建监听器
    listener, err := net.Listen("tcp", s.address)
    if err != nil {
        return fmt.Errorf("failed to start server: %w", err)
    }
    s.listener = listener
    
    fmt.Printf("TCP服务器启动，监听地址: %s\n", s.address)
    
    // 2. 循环接受连接
    for {
        conn, err := listener.Accept()
        if err != nil {
            // 服务器关闭时会到这里
            return err
        }
        
        // 3. 为每个连接启动goroutine
        go s.handleConnection(conn)
    }
}
```

### 2. `handleConnection(conn net.Conn)` 方法

**功能**: 处理单个客户端连接

**实现步骤**:
```go
func (s *TCPServer) handleConnection(conn net.Conn) {
    defer conn.Close()
    
    clientAddr := conn.RemoteAddr().String()
    fmt.Printf("客户端连接: %s\n", clientAddr)
    
    // 创建JSON解码器和编码器
    decoder := json.NewDecoder(conn)
    encoder := json.NewEncoder(conn)
    
    for {
        // 1. 读取请求
        var request protocol.Request
        if err := decoder.Decode(&request); err != nil {
            fmt.Printf("解析请求失败: %v\n", err)
            break
        }
        
        fmt.Printf("收到请求: %s (ID: %s)\n", request.Type, request.RequestID)
        
        // 2. 处理请求
        response := s.handleRequest(&request)
        
        // 3. 发送响应
        if err := encoder.Encode(response); err != nil {
            fmt.Printf("发送响应失败: %v\n", err)
            break
        }
    }
    
    fmt.Printf("客户端断开: %s\n", clientAddr)
}
```

### 3. `handleRequest(request *protocol.Request)` 方法

**功能**: 根据请求类型分发处理

**实现步骤**:
```go
func (s *TCPServer) handleRequest(request *protocol.Request) *protocol.Response {
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
```

## 🔧 你还需要实现的辅助方法

实现了上面3个方法后，你还需要实现具体的处理方法：

### handleCreateTopic 方法
```go
func (s *TCPServer) handleCreateTopic(request *protocol.Request) *protocol.Response {
    // TODO: 你来实现
    // 1. 将request.Data解析为CreateTopicRequest
    // 2. 调用broker.CreateTopic
    // 3. 返回CreateTopicResponse
}
```

### handleProduce 方法
```go
func (s *TCPServer) handleProduce(request *protocol.Request) *protocol.Response {
    // TODO: 你来实现  
    // 1. 将request.Data解析为ProduceRequest
    // 2. 创建Message对象
    // 3. 调用broker.ProduceMessage
    // 4. 返回ProduceResponse
}
```

### handleConsume 方法
```go
func (s *TCPServer) handleConsume(request *protocol.Request) *protocol.Response {
    // TODO: 你来实现
    // 1. 将request.Data解析为ConsumeRequest  
    // 2. 调用broker.ConsumeMessages
    // 3. 转换Message为NetworkMessage
    // 4. 返回ConsumeResponse
}
```

## 💡 实现提示

### JSON解析技巧
```go
// 将interface{}解析为具体类型
var createReq protocol.CreateTopicRequest
reqData, _ := json.Marshal(request.Data)
json.Unmarshal(reqData, &createReq)
```

### 错误处理
- 每个方法都要有适当的错误处理
- 使用 `s.createErrorResponse()` 创建错误响应
- 使用 `s.createSuccessResponse()` 创建成功响应

### 数据转换
- common.Message → protocol.NetworkMessage
- 注意时间格式转换

## ✅ 完成标准
- [ ] Start方法能正确启动服务器
- [ ] handleConnection能处理多个并发连接
- [ ] handleRequest能正确分发不同类型请求
- [ ] 所有辅助处理方法实现完成
- [ ] 能够编译通过
- [ ] 基础的错误处理完善

你准备好开始实现了吗？从哪个方法开始？