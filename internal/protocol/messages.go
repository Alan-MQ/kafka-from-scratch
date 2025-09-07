package protocol

// RequestType 定义请求类型
type RequestType string

const (
	RequestTypeCreateTopic RequestType = "CREATE_TOPIC"
	RequestTypeProduce     RequestType = "PRODUCE"
	RequestTypeConsume     RequestType = "CONSUME"
	RequestTypeSubscribe   RequestType = "SUBSCRIBE"
	RequestTypeSeek        RequestType = "SEEK"
)

// Request 通用请求结构
type Request struct {
	Type      RequestType `json:"type"`
	RequestID string      `json:"request_id"`
	Data      interface{} `json:"data"`
}

// Response 通用响应结构
type Response struct {
	RequestID string      `json:"request_id"`
	Success   bool        `json:"success"`
	Error     string      `json:"error,omitempty"`
	Data      interface{} `json:"data,omitempty"`
}