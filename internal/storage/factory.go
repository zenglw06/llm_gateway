package storage

import (
	"context"

	"github.com/zenglw/llm_gateway/internal/model"
	"github.com/zenglw/llm_gateway/internal/storage/memory"
	"github.com/zenglw/llm_gateway/internal/storage/redis"
	"github.com/zenglw/llm_gateway/pkg/config"
)

// Store 存储接口
type Store interface {
	// Create 创建API Key
	Create(ctx context.Context, apiKey *model.APIKey) error
	// GetByKey 根据Key哈希获取API Key
	GetByKey(ctx context.Context, keyHash string) (*model.APIKey, error)
	// GetByID 根据ID获取API Key
	GetByID(ctx context.Context, id string) (*model.APIKey, error)
	// Update 更新API Key
	Update(ctx context.Context, apiKey *model.APIKey) error
	// Delete 删除API Key
	Delete(ctx context.Context, id string) error
	// ListByUserID 根据用户ID获取API Key列表
	ListByUserID(ctx context.Context, userID string) ([]*model.APIKey, error)
	// GetUserQuota 获取用户配额
	GetUserQuota(ctx context.Context, userID string) (*model.UserQuota, error)
	// UpdateUserQuota 更新用户配额
	UpdateUserQuota(ctx context.Context, quota *model.UserQuota) error
	// IncrementUsage 增加使用量
	IncrementUsage(ctx context.Context, userID string, amount int) (remaining int, err error)
	// ResetUsage 重置使用量
	ResetUsage(ctx context.Context, userID string) error
	// Close 关闭存储
	Close() error
}

// NewStore 根据配置创建存储实例
func NewStore(cfg config.StorageConfig) (Store, error) {
	switch cfg.Type {
	case "redis":
		return redis.NewRedisStore(cfg.Redis)
	case "memory":
		fallthrough
	default:
		return memory.NewStore(), nil
	}
}
