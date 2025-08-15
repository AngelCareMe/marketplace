package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"marketplace/internal/adapter/bcrypt"
	"marketplace/internal/adapter/jwt"
	"marketplace/internal/adapter/postgres/customer"
	productAdapter "marketplace/internal/adapter/postgres/product"
	"marketplace/internal/adapter/postgres/seller"
	"marketplace/internal/adapter/postgres/token"
	"marketplace/internal/adapter/postgres/user"
	"marketplace/internal/handler/auth"
	"marketplace/internal/handler/product"
	usecase "marketplace/internal/usecase/auth"
	usecaseProduct "marketplace/internal/usecase/product"
	"marketplace/pkg/config"
	adapter "marketplace/pkg/pgxpool"

	"github.com/go-playground/validator/v10"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {
	// Флаги для config
	configPath := flag.String("config", "config.yaml", "path to config file")
	flag.Parse()

	// Инициализация Viper
	viper.SetConfigFile(*configPath)
	viper.AddConfigPath(".")
	viper.SetEnvPrefix("app")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("failed to read config: %v", err)
	}

	var cfg config.Config
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("failed to unmarshal config: %v", err)
	}

	// Инициализация logrus напрямую
	rawLogger := logrus.New()
	rawLogger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	rawLogger.SetLevel(logrus.InfoLevel)

	ctx := context.Background()
	pool, err := adapter.InitDBPool(ctx, &cfg, rawLogger)
	if err != nil {
		rawLogger.Fatalf("failed to init DB pool: %v", err)
	}
	defer pool.Close()

	// Репозитории
	userRepo := user.NewUserRepository(pool, rawLogger)
	customerRepo := customer.NewCustomerRepository(pool, rawLogger)
	sellerRepo := seller.NewSellerRepository(pool, rawLogger)
	tokenRepo := token.NewTokenRepository(pool, rawLogger)
	productRepo := productAdapter.NewProductRepository(pool, rawLogger)

	// Менеджеры
	bcryptManager := bcrypt.NewBcryptManager(rawLogger, 12)
	jwtManager := jwt.NewJWTManager(tokenRepo, rawLogger, cfg)

	// Usecase
	authUsecase := usecase.NewAuthUsecase(userRepo, customerRepo, sellerRepo, tokenRepo, jwtManager, bcryptManager, rawLogger)
	productUsecase := usecaseProduct.NewProductUsecase(productRepo, rawLogger, validator.New())

	// Handler
	authHandler := auth.NewAuthHandler(authUsecase, rawLogger)
	productHandler := product.NewProductHandler(productUsecase, rawLogger)

	// Gin router
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())

	// Группа маршрутов
	apiGroup := r.Group("/")
	auth.RegisterAuthRoutes(apiGroup, authHandler, jwtManager, rawLogger)
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "alive"})
	})
	product.RegisterProductRoutes(apiGroup, productHandler, jwtManager, rawLogger)
	r.POST("/test", func(c *gin.Context) {
		var data map[string]interface{}
		c.BindJSON(&data)
		c.JSON(http.StatusOK, gin.H{"success": true, "received": data})
	})

	// HTTP server
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			rawLogger.Fatalf("server listen failed: %v", err)
		}
	}()

	rawLogger.Infof("server started on %s", addr)

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	rawLogger.Info("shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		rawLogger.Fatalf("server shutdown failed: %v", err)
	}

	rawLogger.Info("server exited gracefully")
}
