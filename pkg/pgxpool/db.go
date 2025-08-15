package adapter

import (
	"context"
	"fmt"
	"marketplace/pkg/config"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

func BuildDSN(cfg *config.Config) string {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.DB.User,
		cfg.DB.Password,
		cfg.DB.Host,
		cfg.DB.Port,
		cfg.DB.Name,
		cfg.DB.SSLMode,
	)

	return dsn
}

func InitDBPool(ctx context.Context, cfg *config.Config, log *logrus.Logger) (*pgxpool.Pool, error) {
	dsn := BuildDSN(cfg)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pool's config: %w", err)
	}

	poolConfig.MaxConns = 20
	poolConfig.MinConns = 5
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = time.Minute * 30

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		log.WithError(err).Error("Failed to initializate database pool")
		return nil, fmt.Errorf("failed to initializate database pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		log.WithFields(logrus.Fields{
			"dsn": dsn,
		}).Errorf("Failed to ping database: %v", err)
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.WithFields(logrus.Fields{
		"dsn": dsn,
	}).Infof("Successfuly connected to database")

	return pool, nil
}
