package lark

type SimpleMessage struct {
	MsgType string `json:"msg_type"`
	Content Text   `json:"content"`
}

type Text struct {
	Text string `json:"text"`
}

// InitSimpleMessage 初始化一个简单的文本消息
func initSimpleMessage(message string) *SimpleMessage {
	return &SimpleMessage{
		MsgType: "text",
		Content: Text{
			Text: message,
		},
	}
}
