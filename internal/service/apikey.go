package service

import (
    "context"
    "crypto/rand"
    "encoding/hex"

    "github.com/zenglw/llm_gateway/internal/model"
    "github.com/zenglw/llm_gateway/internal/storage"
    "github.com/zenglw/llm_gateway/pkg/errors"
)

// APIKeyService API Key服务
type APIKeyService struct {
    store storage.Store
}

// NewAPIKeyService 创建API Key服务
func NewAPIKeyService(store storage.Store) *APIKeyService {
    return &APIKeyService{
        store: store,
    }
}

// CreateAPIKey 创建API Key
func (s *APIKeyService) CreateAPIKey(ctx context.Context, req *model.CreateAPIKeyRequest) (*model.APIKey, error) {
    // 生成随机API Key
    key, err := generateAPIKey()
    if err != nil {
        return nil, errors.Wrap(errors.ErrCodeInternal, "failed to generate api key", err)
    }

    apiKey := &model.APIKey{
        UserID:     req.UserID,
        Key:        key, // 明文Key，只在创建时返回
        Name:       req.Name,
        Status:     model.APIKeyStatusEnabled,
        Permissions: req.Permissions,
        AllowedIPs: req.AllowedIPs,
        ExpiresAt:  req.ExpiresAt,
    }

    if err := s.store.Create(ctx, apiKey); err != nil {
        return nil, err
    }

    return apiKey, nil
}

// GetAPIKey 获取API Key详情
func (s *APIKeyService) GetAPIKey(ctx context.Context, id string) (*model.APIKeyResponse, error) {
    apiKey, err := s.store.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }

    return convertToAPIKeyResponse(apiKey), nil
}

// ListAPIKeys 获取用户API Key列表
func (s *APIKeyService) ListAPIKeys(ctx context.Context, userID string) ([]*model.APIKeyResponse, error) {
    apiKeys, err := s.store.ListByUserID(ctx, userID)
    if err != nil {
        return nil, err
    }

    resp := make([]*model.APIKeyResponse, len(apiKeys))
    for i, apiKey := range apiKeys {
        resp[i] = convertToAPIKeyResponse(apiKey)
    }

    return resp, nil
}

// UpdateAPIKey 更新API Key
func (s *APIKeyService) UpdateAPIKey(ctx context.Context, id string, req *model.UpdateAPIKeyRequest) (*model.APIKeyResponse, error) {
    apiKey, err := s.store.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }

    // 更新字段
    if req.Name != nil {
        apiKey.Name = *req.Name
    }
    if req.Status != nil {
        apiKey.Status = *req.Status
    }
    if req.Permissions != nil {
        apiKey.Permissions = *req.Permissions
    }
    if req.AllowedIPs != nil {
        apiKey.AllowedIPs = *req.AllowedIPs
    }
    if req.ExpiresAt != nil {
        apiKey.ExpiresAt = *req.ExpiresAt
    }

    if err := s.store.Update(ctx, apiKey); err != nil {
        return nil, err
    }

    return convertToAPIKeyResponse(apiKey), nil
}

// DeleteAPIKey 删除API Key
func (s *APIKeyService) DeleteAPIKey(ctx context.Context, id string) error {
    return s.store.Delete(ctx, id)
}

// convertToAPIKeyResponse 转换为响应对象
func convertToAPIKeyResponse(apiKey *model.APIKey) *model.APIKeyResponse {
    return &model.APIKeyResponse{
        ID:         apiKey.ID,
        UserID:     apiKey.UserID,
        Name:       apiKey.Name,
        Status:     apiKey.Status,
        Permissions: apiKey.Permissions,
        AllowedIPs: apiKey.AllowedIPs,
        ExpiresAt:  apiKey.ExpiresAt,
        CreatedAt:  apiKey.CreatedAt,
        UpdatedAt:  apiKey.UpdatedAt,
        LastUsedAt: apiKey.LastUsedAt,
    }
}

// generateAPIKey 生成API Key
func generateAPIKey() (string, error) {
    b := make([]byte, 16)
    _, err := rand.Read(b)
    if err != nil {
        return "", err
    }
    return "sk-" + hex.EncodeToString(b), nil
}
