package memory

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"

	"github.com/zenglw/llm_gateway/internal/model"
	"github.com/zenglw/llm_gateway/pkg/errors"
	"github.com/zenglw/llm_gateway/pkg/utils"
)

// Store 内存存储实现
type Store struct {
	apiKeys    map[string]*model.APIKey // key: api key id
	apiKeyHash map[string]string        // key: key hash, value: api key id
	quotas     map[string]*model.UserQuota
	mu         sync.RWMutex
}

// NewStore 创建内存存储实例
func NewStore() *Store {
	return &Store{
		apiKeys:    make(map[string]*model.APIKey),
		apiKeyHash: make(map[string]string),
		quotas:     make(map[string]*model.UserQuota),
	}
}

// Create 创建API Key
func (s *Store) Create(ctx context.Context, apiKey *model.APIKey) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 生成ID
	if apiKey.ID == "" {
		apiKey.ID = utils.RandomString(16)
	}

	// 计算Key的哈希值
	hash := sha256.Sum256([]byte(apiKey.Key))
	keyHash := hex.EncodeToString(hash[:])
	apiKey.KeyHash = keyHash

	// 设置创建和更新时间
	now := time.Now()
	if apiKey.CreatedAt.IsZero() {
		apiKey.CreatedAt = now
	}
	if apiKey.UpdatedAt.IsZero() {
		apiKey.UpdatedAt = now
	}

	// 保存
	s.apiKeys[apiKey.ID] = apiKey
	s.apiKeyHash[keyHash] = apiKey.ID

	// 如果用户配额不存在，创建默认配额
	if _, ok := s.quotas[apiKey.UserID]; !ok {
		s.quotas[apiKey.UserID] = &model.UserQuota{
			UserID:        apiKey.UserID,
			TotalRequests: 10000, // 默认总配额10000
			DailyLimit:    1000,  // 默认每日限额1000
			MonthlyLimit:  10000, // 默认每月限额10000
			LastResetDate: now,
			UpdatedAt:     now,
		}
	}

	return nil
}

// GetByKey 根据Key获取API Key
func (s *Store) GetByKey(ctx context.Context, keyHash string) (*model.APIKey, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	id, ok := s.apiKeyHash[keyHash]
	if !ok {
		return nil, errors.New(errors.ErrCodeAPIKeyInvalid, "api key not found")
	}

	apiKey, ok := s.apiKeys[id]
	if !ok {
		return nil, errors.New(errors.ErrCodeAPIKeyInvalid, "api key not found")
	}

	// 返回副本，防止外部修改
	copyKey := *apiKey
	return &copyKey, nil
}

// GetByID 根据ID获取API Key
func (s *Store) GetByID(ctx context.Context, id string) (*model.APIKey, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	apiKey, ok := s.apiKeys[id]
	if !ok {
		return nil, errors.New(errors.ErrCodeNotFound, "api key not found")
	}

	// 返回副本
	copyKey := *apiKey
	return &copyKey, nil
}

// Update 更新API Key
func (s *Store) Update(ctx context.Context, apiKey *model.APIKey) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.apiKeys[apiKey.ID]
	if !ok {
		return errors.New(errors.ErrCodeNotFound, "api key not found")
	}

	// 更新时间
	apiKey.UpdatedAt = time.Now()

	// 保存
	s.apiKeys[apiKey.ID] = apiKey
	return nil
}

// Delete 删除API Key
func (s *Store) Delete(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	apiKey, ok := s.apiKeys[id]
	if !ok {
		return errors.New(errors.ErrCodeNotFound, "api key not found")
	}

	// 删除哈希映射
	delete(s.apiKeyHash, apiKey.KeyHash)
	// 删除API Key
	delete(s.apiKeys, id)

	return nil
}

// ListByUserID 根据用户ID获取API Key列表
func (s *Store) ListByUserID(ctx context.Context, userID string) ([]*model.APIKey, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var list []*model.APIKey
	for _, apiKey := range s.apiKeys {
		if apiKey.UserID == userID {
			// 返回副本
			copyKey := *apiKey
			list = append(list, &copyKey)
		}
	}

	return list, nil
}

// GetUserQuota 获取用户配额
func (s *Store) GetUserQuota(ctx context.Context, userID string) (*model.UserQuota, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	quota, ok := s.quotas[userID]
	if !ok {
		// 如果不存在，创建默认配额
		now := time.Now()
		quota = &model.UserQuota{
			UserID:        userID,
			TotalRequests: 10000,
			DailyLimit:    1000,
			MonthlyLimit:  10000,
			LastResetDate: now,
			UpdatedAt:     now,
		}
		s.quotas[userID] = quota
	}

	// 检查是否需要重置每日/每月配额
	now := time.Now()
	if now.Day() != quota.LastResetDate.Day() || now.Month() != quota.LastResetDate.Month() || now.Year() != quota.LastResetDate.Year() {
		// 跨天，重置每日使用量
		quota.DailyUsed = 0
		if now.Month() != quota.LastResetDate.Month() || now.Year() != quota.LastResetDate.Year() {
			// 跨月，重置每月使用量
			quota.MonthlyUsed = 0
		}
		quota.LastResetDate = now
		quota.UpdatedAt = now
	}

	// 返回副本
	copyQuota := *quota
	return &copyQuota, nil
}

// UpdateUserQuota 更新用户配额
func (s *Store) UpdateUserQuota(ctx context.Context, quota *model.UserQuota) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	quota.UpdatedAt = time.Now()
	s.quotas[quota.UserID] = quota
	return nil
}

// IncrementUsage 增加使用量
func (s *Store) IncrementUsage(ctx context.Context, userID string, amount int) (remaining int, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	quota, ok := s.quotas[userID]
	if !ok {
		now := time.Now()
		quota = &model.UserQuota{
			UserID:        userID,
			TotalRequests: 10000,
			DailyLimit:    1000,
			MonthlyLimit:  10000,
			LastResetDate: now,
			UpdatedAt:     now,
		}
		s.quotas[userID] = quota
	}

	// 检查配额
	if quota.UsedRequests+amount > quota.TotalRequests {
		return 0, errors.New(errors.ErrCodeQuotaExhausted, "total quota exhausted")
	}
	if quota.DailyUsed+amount > quota.DailyLimit {
		return 0, errors.New(errors.ErrCodeQuotaExhausted, "daily quota exhausted")
	}
	if quota.MonthlyUsed+amount > quota.MonthlyLimit {
		return 0, errors.New(errors.ErrCodeQuotaExhausted, "monthly quota exhausted")
	}

	// 增加使用量
	quota.UsedRequests += amount
	quota.DailyUsed += amount
	quota.MonthlyUsed += amount
	quota.UpdatedAt = time.Now()

	// 计算剩余配额
	remaining = min(
		quota.TotalRequests-quota.UsedRequests,
		quota.DailyLimit-quota.DailyUsed,
		quota.MonthlyLimit-quota.MonthlyUsed,
	)

	return remaining, nil
}

// ResetUsage 重置使用量
func (s *Store) ResetUsage(ctx context.Context, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	quota, ok := s.quotas[userID]
	if !ok {
		return errors.New(errors.ErrCodeNotFound, "user quota not found")
	}

	now := time.Now()
	quota.UsedRequests = 0
	quota.DailyUsed = 0
	quota.MonthlyUsed = 0
	quota.LastResetDate = now
	quota.UpdatedAt = now

	return nil
}

// Close 关闭存储
func (s *Store) Close() error {
	// 内存存储不需要关闭
	return nil
}
