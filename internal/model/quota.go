package model

import "time"

// UserQuota 用户配额信息
type UserQuota struct {
	UserID        string    `json:"user_id"`
	TotalRequests int       `json:"total_requests"`  // 总请求数配额
	UsedRequests  int       `json:"used_requests"`   // 已使用请求数
	DailyLimit    int       `json:"daily_limit"`     // 每日请求限额
	DailyUsed     int       `json:"daily_used"`      // 今日已使用
	MonthlyLimit  int       `json:"monthly_limit"`   // 每月请求限额
	MonthlyUsed   int       `json:"monthly_used"`    // 本月已使用
	LastResetDate time.Time `json:"last_reset_date"` // 上次重置日期
	UpdatedAt     time.Time `json:"updated_at"`
}

// QuotaResponse 配额查询响应
type QuotaResponse struct {
	UserID           string `json:"user_id"`
	TotalRequests    int    `json:"total_requests"`
	UsedRequests     int    `json:"used_requests"`
	Remaining        int    `json:"remaining"`
	DailyLimit       int    `json:"daily_limit"`
	DailyUsed        int    `json:"daily_used"`
	DailyRemaining   int    `json:"daily_remaining"`
	MonthlyLimit     int    `json:"monthly_limit"`
	MonthlyUsed      int    `json:"monthly_used"`
	MonthlyRemaining int    `json:"monthly_remaining"`
}

// UpdateQuotaRequest 更新配额请求
type UpdateQuotaRequest struct {
	TotalRequests *int `json:"total_requests"`
	DailyLimit    *int `json:"daily_limit"`
	MonthlyLimit  *int `json:"monthly_limit"`
}
