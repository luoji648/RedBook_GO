package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"redbook-go/internal/auth"
	"redbook-go/internal/model"
	"redbook-go/internal/pkg/logger"
)

func CORS() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}

func Auth(db *gorm.DB, tokens *auth.TokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}
		claims, err := tokens.ParseToken(c.GetHeader("authorization"))
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if tokens.IsBlacklisted(c.Request.Context(), claims.ID) {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		var user model.User
		if err := db.WithContext(c.Request.Context()).First(&user, claims.UserID).Error; err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if user.TokenVersion != claims.TV {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Set(auth.CurrentUserKey, &user)
		c.Next()
	}
}

func OptionalAuth(db *gorm.DB, tokens *auth.TokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}
		claims, err := tokens.ParseToken(c.GetHeader("authorization"))
		if err != nil {
			c.Next()
			return
		}
		if tokens.IsBlacklisted(c.Request.Context(), claims.ID) {
			c.Next()
			return
		}
		var user model.User
		if err := db.WithContext(c.Request.Context()).First(&user, claims.UserID).Error; err != nil {
			c.Next()
			return
		}
		if user.TokenVersion == claims.TV {
			c.Set(auth.CurrentUserKey, &user)
		}
		c.Next()
	}
}

func CurrentUser(c *gin.Context) (*model.User, bool) {
	v, ok := c.Get(auth.CurrentUserKey)
	if !ok {
		return nil, false
	}
	u, ok := v.(*model.User)
	return u, ok
}

func CurrentUserID(c *gin.Context) (int64, bool) {
	u, ok := CurrentUser(c)
	if !ok {
		return 0, false
	}
	return u.ID, true
}

func ContextTimeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
		cancel()
	}
}

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = randomID()
		}
		c.Set("requestID", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		if logger.Log == nil {
			return
		}
		userID := int64(0)
		if u, ok := CurrentUser(c); ok {
			userID = u.ID
		}
		logger.Log.Info("http request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.String("request_id", c.GetString("requestID")),
			zap.Int64("user_id", userID),
			zap.Duration("latency", time.Since(start)),
		)
	}
}

func randomID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
