package redis

const (
	// KeyPrefix Redis键前缀
	KeyPrefix = "llm_gateway:"

	// APIKeyPrefix API Key前缀
	APIKeyPrefix = KeyPrefix + "api_key:"

	// APIKeyHashPrefix API Key哈希前缀
	APIKeyHashPrefix = KeyPrefix + "api_key_hash:"

	// UserQuotaPrefix 用户配额前缀
	UserQuotaPrefix = KeyPrefix + "user_quota:"
)

// GetAPIKeyKey 获取API Key的键
func GetAPIKeyKey(id string) string {
	return APIKeyPrefix + id
}

// GetAPIKeyHashKey 获取API Key哈希的键
func GetAPIKeyHashKey(hash string) string {
	return APIKeyHashPrefix + hash
}

// GetUserQuotaKey 获取用户配额的键
func GetUserQuotaKey(userID string) string {
	return UserQuotaPrefix + userID
}
