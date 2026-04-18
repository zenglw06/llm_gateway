package service

import (
    "context"

    "github.com/zenglw/llm_gateway/internal/model"
    "github.com/zenglw/llm_gateway/internal/storage"
    "github.com/zenglw/llm_gateway/pkg/errors"
)

// QuotaService 配额服务
type QuotaService struct {
    store storage.Store
}

// NewQuotaService 创建配额服务
func NewQuotaService(store storage.Store) *QuotaService {
    return &QuotaService{
        store: store,
    }
}

// GetUserQuota 获取用户配额
func (s *QuotaService) GetUserQuota(ctx context.Context, userID string) (*model.QuotaResponse, error) {
    quota, err := s.store.GetUserQuota(ctx, userID)
    if err != nil {
        return nil, err
    }

    return &model.QuotaResponse{
        UserID:        quota.UserID,
        TotalRequests: quota.TotalRequests,
        UsedRequests:  quota.UsedRequests,
        Remaining:     quota.TotalRequests - quota.UsedRequests,
        DailyLimit:    quota.DailyLimit,
        DailyUsed:     quota.DailyUsed,
        DailyRemaining: quota.DailyLimit - quota.DailyUsed,
        MonthlyLimit:  quota.MonthlyLimit,
        MonthlyUsed:   quota.MonthlyUsed,
        MonthlyRemaining: quota.MonthlyLimit - quota.MonthlyUsed,
    }, nil
}

// UpdateUserQuota 更新用户配额
func (s *QuotaService) UpdateUserQuota(ctx context.Context, userID string, req *model.UpdateQuotaRequest) (*model.QuotaResponse, error) {
    quota, err := s.store.GetUserQuota(ctx, userID)
    if err != nil {
        return nil, err
    }

    // 更新字段
    if req.TotalRequests != nil {
        quota.TotalRequests = *req.TotalRequests
    }
    if req.DailyLimit != nil {
        quota.DailyLimit = *req.DailyLimit
    }
    if req.MonthlyLimit != nil {
        quota.MonthlyLimit = *req.MonthlyLimit
    }

    // 确保已使用量不超过新的限额
    if quota.UsedRequests > quota.TotalRequests {
        quota.UsedRequests = quota.TotalRequests
    }
    if quota.DailyUsed > quota.DailyLimit {
        quota.DailyUsed = quota.DailyLimit
    }
    if quota.MonthlyUsed > quota.MonthlyLimit {
        quota.MonthlyUsed = quota.MonthlyLimit
    }

    if err := s.store.UpdateUserQuota(ctx, quota); err != nil {
        return nil, err
    }

    return &model.QuotaResponse{
        UserID:        quota.UserID,
        TotalRequests: quota.TotalRequests,
        UsedRequests:  quota.UsedRequests,
        Remaining:     quota.TotalRequests - quota.UsedRequests,
        DailyLimit:    quota.DailyLimit,
        DailyUsed:     quota.DailyUsed,
        DailyRemaining: quota.DailyLimit - quota.DailyUsed,
        MonthlyLimit:  quota.MonthlyLimit,
        MonthlyUsed:   quota.MonthlyUsed,
        MonthlyRemaining: quota.MonthlyLimit - quota.MonthlyUsed,
    }, nil
}

// ResetUserQuota 重置用户使用量
func (s *QuotaService) ResetUserQuota(ctx context.Context, userID string) error {
    return s.store.ResetUsage(ctx, userID)
}

// CheckQuota 检查用户配额
func (s *QuotaService) CheckQuota(ctx context.Context, userID string) (int, error) {
    quota, err := s.store.GetUserQuota(ctx, userID)
    if err != nil {
        return 0, err
    }

    // 计算剩余配额
    remaining := min(
        quota.TotalRequests - quota.UsedRequests,
        quota.DailyLimit - quota.DailyUsed,
        quota.MonthlyLimit - quota.MonthlyUsed,
    )

    if remaining <= 0 {
        return 0, errors.New(errors.ErrCodeQuotaExhausted, "quota exhausted")
    }

    return remaining, nil
}

// ConsumeQuota 消费配额
func (s *QuotaService) ConsumeQuota(ctx context.Context, userID string, amount int) (int, error) {
    if amount <= 0 {
        amount = 1 // 默认消费1个配额
    }

    return s.store.IncrementUsage(ctx, userID, amount)
}
