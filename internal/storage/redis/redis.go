package redis

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/zenglw/llm_gateway/internal/model"
	"github.com/zenglw/llm_gateway/pkg/config"
	"github.com/zenglw/llm_gateway/pkg/errors"
	"github.com/zenglw/llm_gateway/pkg/utils"
)

// RedisStore Redis存储实现
type RedisStore struct {
	client *redis.Client
}

// NewRedisStore 创建Redis存储实例
func NewRedisStore(cfg config.RedisConfig) (*RedisStore, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &RedisStore{client: client}, nil
}

// Create 创建API Key
func (s *RedisStore) Create(ctx context.Context, apiKey *model.APIKey) error {
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

	// 序列化API Key
	data, err := json.Marshal(apiKey)
	if err != nil {
		return err
	}

	// 开启事务
	tx := s.client.TxPipeline()

	// 保存API Key
	apiKeyKey := GetAPIKeyKey(apiKey.ID)
	tx.Set(ctx, apiKeyKey, data, 0) // 永不过期

	// 保存哈希映射
	hashKey := GetAPIKeyHashKey(keyHash)
	tx.Set(ctx, hashKey, apiKey.ID, 0)

	// 如果用户配额不存在，创建默认配额
	quotaKey := GetUserQuotaKey(apiKey.UserID)
	exists, err := s.client.Exists(ctx, quotaKey).Result()
	if err != nil {
		return err
	}
	if exists == 0 {
		quota := &model.UserQuota{
			UserID:        apiKey.UserID,
			TotalRequests: 10000, // 默认总配额10000
			DailyLimit:    1000,  // 默认每日限额1000
			MonthlyLimit:  10000, // 默认每月限额10000
			LastResetDate: now,
			UpdatedAt:     now,
		}
		quotaData, err := json.Marshal(quota)
		if err != nil {
			return err
		}
		tx.Set(ctx, quotaKey, quotaData, 0)
	}

	// 执行事务
	_, err = tx.Exec(ctx)
	return err
}

// GetByKey 根据Key获取API Key
func (s *RedisStore) GetByKey(ctx context.Context, keyHash string) (*model.APIKey, error) {
	hashKey := GetAPIKeyHashKey(keyHash)
	id, err := s.client.Get(ctx, hashKey).Result()
	if err == redis.Nil {
		return nil, errors.New(errors.ErrCodeAPIKeyInvalid, "api key not found")
	} else if err != nil {
		return nil, err
	}

	apiKeyKey := GetAPIKeyKey(id)
	data, err := s.client.Get(ctx, apiKeyKey).Result()
	if err == redis.Nil {
		return nil, errors.New(errors.ErrCodeAPIKeyInvalid, "api key not found")
	} else if err != nil {
		return nil, err
	}

	var apiKey model.APIKey
	if err := json.Unmarshal([]byte(data), &apiKey); err != nil {
		return nil, err
	}

	// 返回副本
	copyKey := apiKey
	return &copyKey, nil
}

// GetByID 根据ID获取API Key
func (s *RedisStore) GetByID(ctx context.Context, id string) (*model.APIKey, error) {
	apiKeyKey := GetAPIKeyKey(id)
	data, err := s.client.Get(ctx, apiKeyKey).Result()
	if err == redis.Nil {
		return nil, errors.New(errors.ErrCodeNotFound, "api key not found")
	} else if err != nil {
		return nil, err
	}

	var apiKey model.APIKey
	if err := json.Unmarshal([]byte(data), &apiKey); err != nil {
		return nil, err
	}

	// 返回副本
	copyKey := apiKey
	return &copyKey, nil
}

// Update 更新API Key
func (s *RedisStore) Update(ctx context.Context, apiKey *model.APIKey) error {
	// 检查是否存在
	apiKeyKey := GetAPIKeyKey(apiKey.ID)
	exists, err := s.client.Exists(ctx, apiKeyKey).Result()
	if err != nil {
		return err
	}
	if exists == 0 {
		return errors.New(errors.ErrCodeNotFound, "api key not found")
	}

	// 更新时间
	apiKey.UpdatedAt = time.Now()

	// 序列化
	data, err := json.Marshal(apiKey)
	if err != nil {
		return err
	}

	// 保存
	return s.client.Set(ctx, apiKeyKey, data, 0).Err()
}

// Delete 删除API Key
func (s *RedisStore) Delete(ctx context.Context, id string) error {
	// 先获取API Key，得到KeyHash
	apiKey, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// 开启事务
	tx := s.client.TxPipeline()

	// 删除API Key
	apiKeyKey := GetAPIKeyKey(id)
	tx.Del(ctx, apiKeyKey)

	// 删除哈希映射
	hashKey := GetAPIKeyHashKey(apiKey.KeyHash)
	tx.Del(ctx, hashKey)

	// 执行事务
	_, err = tx.Exec(ctx)
	return err
}

// ListByUserID 根据用户ID获取API Key列表
func (s *RedisStore) ListByUserID(ctx context.Context, userID string) ([]*model.APIKey, error) {
	// 注意：Redis不支持按字段查询，这里需要使用SCAN或者二级索引
	// 简单实现：遍历所有API Key，匹配用户ID，生产环境建议使用二级索引
	var list []*model.APIKey

	iter := s.client.Scan(ctx, 0, APIKeyPrefix+"*", 0).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		data, err := s.client.Get(ctx, key).Result()
		if err != nil {
			continue
		}

		var apiKey model.APIKey
		if err := json.Unmarshal([]byte(data), &apiKey); err != nil {
			continue
		}

		if apiKey.UserID == userID {
			// 返回副本
			copyKey := apiKey
			list = append(list, &copyKey)
		}
	}

	if err := iter.Err(); err != nil {
		return nil, err
	}

	return list, nil
}

// GetUserQuota 获取用户配额
func (s *RedisStore) GetUserQuota(ctx context.Context, userID string) (*model.UserQuota, error) {
	quotaKey := GetUserQuotaKey(userID)
	data, err := s.client.Get(ctx, quotaKey).Result()
	if err == redis.Nil {
		// 如果不存在，创建默认配额
		now := time.Now()
		quota := &model.UserQuota{
			UserID:        userID,
			TotalRequests: 10000,
			DailyLimit:    1000,
			MonthlyLimit:  10000,
			LastResetDate: now,
			UpdatedAt:     now,
		}
		quotaData, err := json.Marshal(quota)
		if err != nil {
			return nil, err
		}
		if err := s.client.Set(ctx, quotaKey, quotaData, 0).Err(); err != nil {
			return nil, err
		}
		return quota, nil
	} else if err != nil {
		return nil, err
	}

	var quota model.UserQuota
	if err := json.Unmarshal([]byte(data), &quota); err != nil {
		return nil, err
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

		// 保存更新
		quotaData, err := json.Marshal(quota)
		if err != nil {
			return nil, err
		}
		if err := s.client.Set(ctx, quotaKey, quotaData, 0).Err(); err != nil {
			return nil, err
		}
	}

	// 返回副本
	copyQuota := quota
	return &copyQuota, nil
}

// UpdateUserQuota 更新用户配额
func (s *RedisStore) UpdateUserQuota(ctx context.Context, quota *model.UserQuota) error {
	quotaKey := GetUserQuotaKey(quota.UserID)
	quota.UpdatedAt = time.Now()
	data, err := json.Marshal(quota)
	if err != nil {
		return err
	}
	return s.client.Set(ctx, quotaKey, data, 0).Err()
}

// IncrementUsage 增加使用量
func (s *RedisStore) IncrementUsage(ctx context.Context, userID string, amount int) (remaining int, err error) {
	// 使用Lua脚本保证原子性
	script := redis.NewScript(`
		local key = KEYS[1]
		local amount = tonumber(ARGV[1])
		local now = tonumber(ARGV[2])

		local data = redis.call('GET', key)
		if not data then
			-- 配额不存在，创建默认配额
			local quota = {
				UserID = ARGV[3],
				TotalRequests = 10000,
				DailyLimit = 1000,
				MonthlyLimit = 10000,
				UsedRequests = 0,
				DailyUsed = 0,
				MonthlyUsed = 0,
				LastResetDate = now,
				UpdatedAt = now
			}
			redis.call('SET', key, cjson.encode(quota))
			data = redis.call('GET', key)
		end

		local quota = cjson.decode(data)
		local reset_day = tonumber(os.date("%d", quota.LastResetDate))
		local reset_month = tonumber(os.date("%m", quota.LastResetDate))
		local reset_year = tonumber(os.date("%Y", quota.LastResetDate))
		local current_day = tonumber(os.date("%d", now))
		local current_month = tonumber(os.date("%m", now))
		local current_year = tonumber(os.date("%Y", now))

		-- 检查是否需要重置配额
		if current_day ~= reset_day or current_month ~= reset_month or current_year ~= reset_year then
			quota.DailyUsed = 0
			if current_month ~= reset_month or current_year ~= reset_year then
				quota.MonthlyUsed = 0
			end
			quota.LastResetDate = now
		end

		-- 检查配额是否足够
		if quota.UsedRequests + amount > quota.TotalRequests then
			return -1 -- 总配额不足
		end
		if quota.DailyUsed + amount > quota.DailyLimit then
			return -2 -- 日配额不足
		end
		if quota.MonthlyUsed + amount > quota.MonthlyLimit then
			return -3 -- 月配额不足
		end

		-- 增加使用量
		quota.UsedRequests = quota.UsedRequests + amount
		quota.DailyUsed = quota.DailyUsed + amount
		quota.MonthlyUsed = quota.MonthlyUsed + amount
		quota.UpdatedAt = now

		-- 保存更新
		redis.call('SET', key, cjson.encode(quota))

		-- 返回剩余配额
		local remaining_total = quota.TotalRequests - quota.UsedRequests
		local remaining_daily = quota.DailyLimit - quota.DailyUsed
		local remaining_monthly = quota.MonthlyLimit - quota.MonthlyUsed

		return math.min(remaining_total, remaining_daily, remaining_monthly)
	`)

	quotaKey := GetUserQuotaKey(userID)
	now := time.Now().Unix()

	res, err := script.Run(ctx, s.client, []string{quotaKey}, amount, now, userID).Int()
	if err != nil {
		return 0, err
	}

	if res < 0 {
		switch res {
		case -1:
			return 0, errors.New(errors.ErrCodeQuotaExhausted, "total quota exhausted")
		case -2:
			return 0, errors.New(errors.ErrCodeQuotaExhausted, "daily quota exhausted")
		case -3:
			return 0, errors.New(errors.ErrCodeQuotaExhausted, "monthly quota exhausted")
		}
	}

	return res, nil
}

// ResetUsage 重置使用量
func (s *RedisStore) ResetUsage(ctx context.Context, userID string) error {
	quotaKey := GetUserQuotaKey(userID)
	quota, err := s.GetUserQuota(ctx, userID)
	if err != nil {
		return err
	}

	now := time.Now()
	quota.UsedRequests = 0
	quota.DailyUsed = 0
	quota.MonthlyUsed = 0
	quota.LastResetDate = now
	quota.UpdatedAt = now

	data, err := json.Marshal(quota)
	if err != nil {
		return err
	}

	return s.client.Set(ctx, quotaKey, data, 0).Err()
}

// Close 关闭存储
func (s *RedisStore) Close() error {
	return s.client.Close()
}
