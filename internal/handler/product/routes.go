package product

import (
	"marketplace/internal/adapter/jwt"
	"marketplace/internal/handler/middleware"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func RegisterProductRoutes(rg *gin.RouterGroup, h *productHandler, jwtManager jwt.JWTManager, log *logrus.Logger) {
	publicGroup := rg.Group("/")
	publicGroup.Use(middleware.AccessTokenMiddleware(jwtManager, log))
	{
		publicGroup.GET("/products/title/:title", h.GetByTitle)
		publicGroup.GET("/categories/:categoryID/products", h.List)
	}

	sellerGroup := rg.Group("/")
	sellerGroup.Use(middleware.AccessTokenMiddleware(jwtManager, log))
	sellerGroup.Use(middleware.RequireRole(middleware.UserTypeSeller, log))
	{
		sellerGroup.POST("/categories/:categoryID/products", h.Create)
		sellerGroup.PUT("/products/:productID", h.Update)
		sellerGroup.DELETE("/products/:productID", h.Delete)
	}
}
