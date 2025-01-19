package message

import "time"

type CreateMessageRequest struct {
	PhoneNumber string `json:"phoneNumber" validate:"required,e164"`
	Content     string `json:"content" validate:"required,min=1,max=40,startsnotwith= "`
}

type CreateMessageResponse struct {
	Id          string     `json:"id"`
	PhoneNumber string     `json:"phoneNumber"`
	Content     string     `json:"content"`
	Status      string     `json:"status"`
	CreatedAt   *time.Time `json:"createdAt"`
}
