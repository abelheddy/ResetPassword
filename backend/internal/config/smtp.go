package config

import "time"

type SMTPConfig struct {
	ID        int       `json:"id,omitempty"`
	Host      string    `json:"host" binding:"required"`
	Port      int       `json:"port" binding:"required"`
	Username  string    `json:"username" binding:"required"`
	Password  string    `json:"password" binding:"required"`
	FromEmail string    `json:"from_email" binding:"required"`
	IsActive  bool      `json:"is_active,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}