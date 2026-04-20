package utils

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"math/big"
)

// GenerateAPIKey 生成随机API Key
func GenerateAPIKey() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		// 如果随机生成失败，使用备用方式
		return "ak_" + RandomString(32)
	}
	return "ak_" + hex.EncodeToString(b)
}

// RandomString 生成指定长度的随机字符串
func RandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		b[i] = charset[n.Int64()]
	}
	return string(b)
}

// Contains 判断字符串是否在数组中
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Ptr 返回指针
func Ptr[T any](v T) *T {
	return &v
}

// MapToStruct map转换为struct
func MapToStruct(m map[string]interface{}, s interface{}) error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, s)
}
