package model

import (
	"time"

	"gorm.io/gorm"
)

type BlacklistedToken struct {
	gorm.Model
	Token     string    `gorm:"type:text;uniqueIndex;not null"`
	ExpiresAt time.Time `gorm:"index;not null"`
}

func (BlacklistedToken) TableName() string {
	return "blacklisted_tokens"
}
