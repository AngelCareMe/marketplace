package auth

import (
	"marketplace/internal/adapter/jwt"
	"marketplace/internal/handler/middleware"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func RegisterAuthRoutes(rg *gin.RouterGroup, h *AuthHandler, jwtManager jwt.JWTManager, log *logrus.Logger) {
	auth := rg.Group("/auth")

	auth.POST("/register", h.Register)
	auth.POST("/login", h.Login)

	auth.PUT("/update-auth", middleware.AccessTokenMiddleware(jwtManager, log), h.UpdateAuth)

	auth.PUT("/update-profile", middleware.AccessTokenMiddleware(jwtManager, log), h.UpdateProfile)
	auth.DELETE("/delete", middleware.AccessTokenMiddleware(jwtManager, log), h.DeleteUser)
}
