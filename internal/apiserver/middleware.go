package apiserver

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/zenglw/llm_gateway/pkg/errors"
	"github.com/zenglw/llm_gateway/pkg/logger"
)

// responseWriter 自定义响应Writer，用于捕获响应内容
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// LoggerMiddleware 日志中间件
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// 读取请求体
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			// 重新设置请求体，供后续处理使用
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// 包装响应Writer
		writer := &responseWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBufferString(""),
		}
		c.Writer = writer

		// 处理请求
		c.Next()

		// 计算耗时
		duration := time.Since(start)
		statusCode := c.Writer.Status()

		// 记录日志
		logger.Infof("%s %s %d %s", method, path, statusCode, duration)

		// 如果是错误，记录详细信息
		if statusCode >= 400 {
			logger.Errorf("Request failed: %s %s, status: %d, request: %s, response: %s",
				method, path, statusCode, string(requestBody), writer.body.String())
		}
	}
}

// CORS 跨域中间件
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		c.Header("Access-Control-Max-Age", "86400")

		// 处理OPTIONS请求
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	}
}

// AuthMiddleware 认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "missing authorization header",
			})
			c.Abort()
			return
		}

		// 将Authorization头存入上下文
		ctx := context.WithValue(c.Request.Context(), "Authorization", authHeader)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// PanicRecoveryMiddleware panic恢复中间件
func PanicRecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Errorf("Panic recovered: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "internal server error",
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}

// ErrorHandlerMiddleware 错误处理中间件
func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 检查是否有错误
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err

			// 判断是否是自定义错误
			if customErr, ok := err.(*errors.Error); ok {
				c.JSON(http.StatusBadRequest, gin.H{
					"code":    customErr.Code,
					"message": customErr.Message,
				})
				return
			}

			// 默认错误
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    errors.ErrCodeInternal,
				"message": err.Error(),
			})
		}
	}
}

// ValidateContentTypeMiddleware 验证Content-Type中间件
func ValidateContentTypeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "POST" || c.Request.Method == "PUT" {
			contentType := c.GetHeader("Content-Type")
			if !strings.Contains(contentType, "application/json") {
				c.JSON(http.StatusUnsupportedMediaType, gin.H{
					"error": "unsupported content type, expected application/json",
				})
				c.Abort()
				return
			}
		}
		c.Next()
	}
}
