package model

import (
	"time"
)

type Channel struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	TenantID        uint      `json:"tenant_id"`
	CustomerID      uint      `json:"customer_id"`
	Status          string    `gorm:"type:enum('open','assigned','closed');default:'open'" json:"status"`
	AssignedAgentID uint      `json:"assigned_agent_id"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (Channel) TableName() string {
	return "channels"
}
