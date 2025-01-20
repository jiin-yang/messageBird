package webhook

import "github.com/google/uuid"

type SendMessageRequest struct {
	To      string `json:"to"`
	Content string `json:"content"`
}

// Burasi webhook ile bagimlilik icerir. Eger webhook response JSON configure etmek istersek burayi da degistirmeliyiz
type SendMessageResponseFromWebhook struct {
	State      string    `json:"state"`
	ResponseId uuid.UUID `json:"responseId"`
}
