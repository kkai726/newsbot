package lark

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// PushToLark 将消息推送到飞书
func PushToLark(webhookURL, message string) error {
	// 创建消息体
	payload := initSimpleMessage(message)

	// 序列化为 JSON
	messageBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("消息序列化失败: %v", err)
	}

	// 发送 POST 请求到飞书的 webhook
	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(messageBytes))
	if err != nil {
		return fmt.Errorf("推送消息失败: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("推送失败: %s", resp.Status)
	}

	return nil
}
