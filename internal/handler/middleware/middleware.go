package middleware

import (
	"fmt"
	"marketplace/internal/adapter/jwt"
	"marketplace/pkg/dto"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
)

const (
	ContextUserID   = "userID"
	ContextUserType = "userType"
)

func AccessTokenMiddleware(jwtManager jwt.JWTManager, logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			logger.Warnf("AccessTokenMiddleware: missing or invalid Authorization header")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid Authorization header"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		parsedToken, err := jwtlib.Parse(tokenString, func(t *jwtlib.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwtlib.SigningMethodHMAC); !ok {
				logger.Errorf("AccessTokenMiddleware: unexpected signing method: %v", t.Header["alg"])
				return nil, fmt.Errorf("unexpected signing method")
			}
			return []byte(jwtManager.Secret()), nil
		})
		if err != nil || !parsedToken.Valid {
			logger.WithFields(map[string]interface{}{
				"token": tokenString,
				"error": err,
			}).Error("AccessTokenMiddleware: failed to parse or validate access token")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid access token"})
			return
		}

		claims, ok := parsedToken.Claims.(jwtlib.MapClaims)
		if !ok {
			logger.Error("AccessTokenMiddleware: failed to cast token claims")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
			return
		}

		userID, ok := claims["user_id"].(string)
		userType, ok2 := claims["user_type"].(string)
		if !ok || !ok2 {
			logger.WithFields(map[string]interface{}{
				"user_id":   claims["user_id"],
				"user_type": claims["user_type"],
			}).Warn("AccessTokenMiddleware: missing claims")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing claims"})
			return
		}

		if err := jwtManager.ValidateAccessToken(tokenString); err != nil {
			logger.WithFields(map[string]interface{}{
				"user_id": userID,
				"error":   err,
			}).Error("AccessTokenMiddleware: token validation failed")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		logger.WithFields(map[string]interface{}{
			"user_id":   userID,
			"user_type": userType,
		}).Info("AccessTokenMiddleware: token validated successfully")

		c.Set("userID", userID)
		c.Set("userType", userType)
		c.Next()
	}
}

func RefreshTokenMiddleware(jwtManager jwt.JWTManager, logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dto.RefreshTokenRequest
		if err := c.ShouldBindJSON(&req); err != nil || req.RefreshToken == "" {
			logger.Warn("RefreshTokenMiddleware: invalid refresh token payload")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid refresh token payload"})
			return
		}

		parsedToken, err := jwtlib.Parse(req.RefreshToken, func(t *jwtlib.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwtlib.SigningMethodHMAC); !ok {
				logger.Errorf("RefreshTokenMiddleware: unexpected signing method: %v", t.Header["alg"])
				return nil, fmt.Errorf("unexpected signing method")
			}
			return []byte(jwtManager.Secret()), nil
		})
		if err != nil || !parsedToken.Valid {
			logger.WithFields(map[string]interface{}{
				"token": req.RefreshToken,
				"error": err,
			}).Error("RefreshTokenMiddleware: failed to parse or validate refresh token")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
			return
		}

		claims, ok := parsedToken.Claims.(jwtlib.MapClaims)
		if !ok {
			logger.Error("RefreshTokenMiddleware: failed to cast token claims")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
			return
		}

		userID, ok := claims["user_id"].(string)
		userType, ok2 := claims["user_type"].(string)
		if !ok || !ok2 {
			logger.WithFields(map[string]interface{}{
				"user_id":   claims["user_id"],
				"user_type": claims["user_type"],
			}).Warn("RefreshTokenMiddleware: missing claims")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing claims"})
			return
		}

		if err := jwtManager.ValidateRefreshToken(c.Request.Context(), req.RefreshToken); err != nil {
			logger.WithFields(map[string]interface{}{
				"user_id": userID,
				"error":   err,
			}).Error("RefreshTokenMiddleware: token validation failed")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		logger.WithFields(map[string]interface{}{
			"user_id":   userID,
			"user_type": userType,
		}).Info("RefreshTokenMiddleware: token validated successfully")

		c.Set("userID", userID)
		c.Set("userType", userType)
		c.Next()
	}
}
