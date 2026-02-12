package model

import (
	"database/sql/driver"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type Role string

const (
	RoleAdmin Role = "admin"
	RoleAgent Role = "agent"
	RoleUser  Role = "user"
)

func (r *Role) Scan(value interface{}) error {
	if value == nil {
		*r = RoleUser
		return nil
	}

	switch v := value.(type) {
	case []byte:
		*r = Role(string(v))
	case string:
		*r = Role(v)
	default:
		return fmt.Errorf("invalid type for Role: %T", value)
	}

	switch *r {
	case RoleAdmin, RoleAgent, RoleUser:
		return nil
	default:
		*r = RoleUser
		return nil
	}
}

func (r Role) Value() (driver.Value, error) {
	return string(r), nil
}

func (r Role) String() string {
	return string(r)
}

func (r Role) MarshalJSON() ([]byte, error) {
	return []byte(`"` + string(r) + `"`), nil
}

func (r *Role) UnmarshalJSON(data []byte) error {
	if len(data) >= 2 {
		data = data[1 : len(data)-1]
	}

	*r = Role(data)

	switch *r {
	case RoleAdmin, RoleAgent, RoleUser:
		return nil
	default:
		return fmt.Errorf("invalid role: %s", data)
	}
}

type User struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	TenantID     uint           `gorm:"index;default:1" json:"tenant_id"`
	Email        string         `gorm:"unique;not null" json:"email"`
	PasswordHash string         `gorm:"not null" json:"-"`
	FullName     string         `json:"full_name"`
	Phone        string         `json:"phone"`
	Avatar       string         `json:"avatar"`
	Role         Role           `gorm:"type:enum('admin','agent','user');default:'user'" json:"role"`
	IsActive     bool           `gorm:"default:true" json:"is_active"`
	LastLoginAt  *time.Time     `json:"last_login_at"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}
