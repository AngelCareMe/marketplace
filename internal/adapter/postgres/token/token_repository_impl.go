package token

import (
	"context"
	"errors"
	"marketplace/internal/entity"
	appErrors "marketplace/pkg/errors"

	sq "github.com/Masterminds/squirrel"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

type tokenRepository struct {
	pool   *pgxpool.Pool
	logger *logrus.Logger
}

func NewTokenRepository(pool *pgxpool.Pool, logger *logrus.Logger) *tokenRepository {
	return &tokenRepository{
		pool:   pool,
		logger: logger,
	}
}

func (r *tokenRepository) GetRefreshTokenByUserID(ctx context.Context, userID string) (*entity.RefreshToken, error) {
	query, args, err := psql.
		Select(
			"user_id",
			"token",
			"expires_at",
			"is_revoked",
			"created_at",
			"updated_at",
		).
		From("tokens").
		Where(sq.Eq{"user_id": userID}).
		ToSql()
	if err != nil {
		r.logger.WithFields(logrus.Fields{
			"method":  "GetRefreshTokenByUserID",
			"user_id": userID,
			"error":   err,
		}).Error("failed to build SQL query")
		return nil, appErrors.ErrInternal
	}

	var t entity.RefreshToken

	row := r.pool.QueryRow(ctx, query, args...)
	if err := row.Scan(
		&t.UserID,
		&t.Token,
		&t.ExpiresAt,
		&t.IsRevoked,
		&t.CreatedAt,
		&t.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.logger.WithFields(logrus.Fields{
				"method":  "GetRefreshTokenByUserID",
				"user_id": userID,
			}).Info("refresh token not found")
			return nil, appErrors.ErrNotFound
		}
		r.logger.WithFields(logrus.Fields{
			"method":  "GetRefreshTokenByUserID",
			"user_id": userID,
			"error":   err,
		}).Error("failed to scan row")
		return nil, appErrors.ErrInternal
	}

	return &t, nil
}

func (r *tokenRepository) UpsertRefreshToken(ctx context.Context, token *entity.RefreshToken) error {
	query, args, err := psql.
		Insert("tokens").
		Columns(
			"user_id",
			"token",
			"expires_at",
			"is_revoked",
			"created_at",
			"updated_at",
		).
		Values(
			token.UserID,
			token.Token,
			token.ExpiresAt,
			token.IsRevoked,
			token.CreatedAt,
			token.UpdatedAt,
		).
		Suffix(`
			ON CONFLICT (user_id) DO UPDATE 
			SET token = EXCLUDED.token,
				expires_at = EXCLUDED.expires_at,
				is_revoked = EXCLUDED.is_revoked,
				updated_at = EXCLUDED.updated_at
		`).
		ToSql()
	if err != nil {
		r.logger.WithFields(logrus.Fields{
			"method":  "UpsertRefreshToken",
			"user_id": token.UserID,
			"error":   err,
		}).Error("failed to build SQL upsert query")
		return appErrors.ErrInternal
	}

	_, err = r.pool.Exec(ctx, query, args...)
	if err != nil {
		r.logger.WithFields(logrus.Fields{
			"method":  "UpsertRefreshToken",
			"user_id": token.UserID,
			"error":   err,
		}).Error("failed to execute upsert query")
		return appErrors.ErrInternal
	}

	r.logger.WithFields(logrus.Fields{
		"method":  "UpsertRefreshToken",
		"user_id": token.UserID,
	}).Info("refresh token successfully upserted")

	return nil
}
