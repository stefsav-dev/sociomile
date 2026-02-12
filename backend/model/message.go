package model

import "time"

type Message struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	ConversationID uint      `json:"conversation_id"`
	SenderType     string    `gorm:"type:enum('customer','agent')" json:"sender_type"`
	Message        string    `gorm:"type:text" json:"message"`
	IsRead         bool      `gorm:"default:false" json:"is_read"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
