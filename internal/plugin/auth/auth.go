package auth

import (
    "context"
    "crypto/sha256"
    "encoding/hex"
    "strings"
    "time"

    "github.com/golang-jwt/jwt/v5"
    "github.com/zenglw/llm_gateway/internal/model"
    "github.com/zenglw/llm_gateway/internal/storage"
    "github.com/zenglw/llm_gateway/pkg/errors"
    "github.com/zenglw/llm_gateway/pkg/logger"
)

// Plugin 鉴权插件
type Plugin struct {
    store      storage.Store
    jwtSecret  []byte
    jwtExpire  time.Duration
}

// Config 鉴权插件配置
type Config struct {
    JWTSecret string `mapstructure:"jwt_secret"`
    JWTExpire int    `mapstructure:"jwt_expire"` // 小时
}

// NewPlugin 创建鉴权插件
func NewPlugin(store storage.Store) *Plugin {
    return &Plugin{
        store: store,
    }
}

// Name 返回插件名称
func (p *Plugin) Name() string {
    return "auth"
}

// Init 初始化插件
func (p *Plugin) Init(config map[string]interface{}) error {
    jwtSecret, ok := config["jwt_secret"].(string)
    if !ok {
        jwtSecret = "default-secret"
    }
    p.jwtSecret = []byte(jwtSecret)

    jwtExpire, ok := config["jwt_expire"].(int)
    if !ok {
        jwtExpire = 24 * 7 // 默认7天
    }
    p.jwtExpire = time.Duration(jwtExpire) * time.Hour

    logger.Info("Auth plugin initialized")
    return nil
}

// Close 关闭插件
func (p *Plugin) Close() error {
    return nil
}

// HandleRequest 处理请求鉴权
func (p *Plugin) HandleRequest(ctx context.Context, req *model.LLMRequest) (*model.LLMRequest, error) {
    authHeader := ctx.Value("Authorization").(string)
    if authHeader == "" {
        return nil, errors.New(errors.ErrCodeUnauthorized, "missing authorization header")
    }

    parts := strings.Split(authHeader, " ")
    if len(parts) != 2 {
        return nil, errors.New(errors.ErrCodeUnauthorized, "invalid authorization header format")
    }

    scheme := strings.ToLower(parts[0])
    token := parts[1]

    switch scheme {
    case "bearer":
        // 尝试JWT认证
        userID, err := p.validateJWT(token)
        if err == nil {
            req.UserID = userID
            return req, nil
        }
        // JWT认证失败，尝试API Key认证
        fallthrough
    case "apikey":
        // API Key认证
        apiKey, err := p.validateAPIKey(ctx, token)
        if err != nil {
            return nil, err
        }
        req.UserID = apiKey.UserID
        req.APIKey = token
        return req, nil
    default:
        return nil, errors.New(errors.ErrCodeUnauthorized, "unsupported authorization scheme")
    }
}

// validateAPIKey 验证API Key
func (p *Plugin) validateAPIKey(ctx context.Context, key string) (*model.APIKey, error) {
    // 计算Key的哈希值
    hash := sha256.Sum256([]byte(key))
    keyHash := hex.EncodeToString(hash[:])

    apiKey, err := p.store.GetByKey(ctx, keyHash)
    if err != nil {
        return nil, errors.New(errors.ErrCodeAPIKeyInvalid, "invalid api key")
    }

    // 检查状态
    if apiKey.Status != model.APIKeyStatusEnabled {
        return nil, errors.New(errors.ErrCodeAPIKeyDisabled, "api key is disabled")
    }

    // 检查过期时间
    if apiKey.ExpiresAt != nil && time.Now().After(*apiKey.ExpiresAt) {
        return nil, errors.New(errors.ErrCodeAPIKeyExpired, "api key has expired")
    }

    // 更新最后使用时间
    now := time.Now()
    apiKey.LastUsedAt = &now
    if err := p.store.Update(ctx, apiKey); err != nil {
        logger.Warnf("Failed to update api key last used time: %v", err)
    }

    return apiKey, nil
}

// Claims JWT Claims
type Claims struct {
    UserID string `json:"user_id"`
    jwt.RegisteredClaims
}

// validateJWT 验证JWT Token
func (p *Plugin) validateJWT(token string) (string, error) {
    claims := &Claims{}

    t, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, errors.New(errors.ErrCodeUnauthorized, "invalid token signing method")
        }
        return p.jwtSecret, nil
    })

    if err != nil {
        return "", errors.Wrap(errors.ErrCodeUnauthorized, "invalid token", err)
    }

    if !t.Valid {
        return "", errors.New(errors.ErrCodeUnauthorized, "invalid token")
    }

    return claims.UserID, nil
}

// GenerateJWT 生成JWT Token
func (p *Plugin) GenerateJWT(userID string) (string, error) {
    expirationTime := time.Now().Add(p.jwtExpire)
    claims := &Claims{
        UserID: userID,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(expirationTime),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            Issuer:    "llm-gateway",
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, err := token.SignedString(p.jwtSecret)
    if err != nil {
        return "", errors.Wrap(errors.ErrCodeInternal, "failed to generate token", err)
    }

    return tokenString, nil
}
