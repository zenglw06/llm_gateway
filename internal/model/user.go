package model

import "time"

// User 用户信息
type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// APIKey API Key信息
type APIKey struct {
	ID          string     `json:"id"`
	UserID      string     `json:"user_id"`
	Key         string     `json:"key,omitempty"` // 仅在创建时返回
	KeyHash     string     `json:"-"`             // 存储的哈希值
	Name        string     `json:"name"`
	Status      int        `json:"status"` // 1:启用 2:禁用
	Permissions []string   `json:"permissions,omitempty"`
	AllowedIPs  []string   `json:"allowed_ips,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
}

const (
	APIKeyStatusEnabled  = 1
	APIKeyStatusDisabled = 2
)

// CreateAPIKeyRequest 创建API Key请求
type CreateAPIKeyRequest struct {
	UserID      string     `json:"user_id" binding:"required"`
	Name        string     `json:"name" binding:"required"`
	Permissions []string   `json:"permissions"`
	AllowedIPs  []string   `json:"allowed_ips"`
	ExpiresAt   *time.Time `json:"expires_at"`
}

// UpdateAPIKeyRequest 更新API Key请求
type UpdateAPIKeyRequest struct {
	Name        *string     `json:"name"`
	Status      *int        `json:"status"`
	Permissions *[]string   `json:"permissions"`
	AllowedIPs  *[]string   `json:"allowed_ips"`
	ExpiresAt   **time.Time `json:"expires_at"`
}

// APIKeyResponse API Key响应
type APIKeyResponse struct {
	ID          string     `json:"id"`
	UserID      string     `json:"user_id"`
	Name        string     `json:"name"`
	Status      int        `json:"status"`
	Permissions []string   `json:"permissions,omitempty"`
	AllowedIPs  []string   `json:"allowed_ips,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
}
