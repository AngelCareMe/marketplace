package seller

import (
	"context"
	"fmt"
	"marketplace/internal/entity"
	appError "marketplace/pkg/errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

type sellerRepository struct {
	pool   *pgxpool.Pool
	logger *logrus.Logger
}

func NewSellerRepository(pool *pgxpool.Pool, logger *logrus.Logger) *sellerRepository {
	return &sellerRepository{pool: pool, logger: logger}
}

func (r *sellerRepository) UpdateProfile(ctx context.Context, profile *entity.SellerProfile) (err error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return appError.NewAppError("TX_BEGIN_FAIL", "could not begin transaction", err)
	}
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(ctx); rbErr != nil {
				r.logger.WithError(rbErr).Error("failed to rollback tx")
			}
		} else if cmErr := tx.Commit(ctx); cmErr != nil {
			err = appError.NewAppError("TX_COMMIT_FAIL", "could not commit transaction", cmErr)
		}
	}()

	sQuery, sArgs, err := psql.
		Update("sellers").
		Set("company_name", profile.CompanyName).
		Set("rating", profile.Rating).
		Where(sq.Eq{"user_id": profile.ID}).
		ToSql()
	if err != nil {
		r.logger.WithError(err).Error("failed to build seller update query")
		return appError.NewAppError("SQL_BUILD_ERROR", "could not build seller update query", err)
	}

	if _, err = tx.Exec(ctx, sQuery, sArgs...); err != nil { // tx!
		return appError.NewAppError("EXEC_ERROR", "could not execute customer update", err)
	}

	uQuery, uArgs, err := psql.
		Update("users").
		Set("updated_at", profile.UpdatedAt).
		Where(sq.Eq{"id": profile.ID}).
		ToSql()
	if err != nil {
		r.logger.WithError(err).Error("failed to build user update query")
		return appError.NewAppError("SQL_BUILD_ERROR", "could not build user update query", err)
	}

	if _, err = tx.Exec(ctx, uQuery, uArgs...); err != nil { // tx!
		return appError.NewAppError("EXEC_ERROR", "could not execute user update", err)
	}

	r.logger.WithField("user_id", profile.ID).Info("seller profile updated successfully")
	return nil
}

func (r *sellerRepository) GetByUsername(ctx context.Context, username string) (*entity.SellerProfile, error) {
	return r.getByField(ctx, "username", username)
}

func (r *sellerRepository) GetByEmail(ctx context.Context, email string) (*entity.SellerProfile, error) {
	return r.getByField(ctx, "email", email)
}

func (r *sellerRepository) getByField(ctx context.Context, field, value string) (*entity.SellerProfile, error) {
	query, args, err := psql.
		Select(
			"u.id", "u.username", "u.password_hash", "u.email",
			"u.updated_at", "u.created_at", "s.company_name", "s.rating",
		).
		From("users u").
		Join("sellers s ON u.id = s.user_id").
		Where(sq.Eq{fmt.Sprintf("u.%s", field): value}).
		ToSql()
	if err != nil {
		r.logger.WithError(err).Error("failed to build getByField query")
		return nil, appError.NewAppError("SQL_BUILD_ERROR", "could not build getByField query", err)
	}

	var s entity.SellerProfile
	row := r.pool.QueryRow(ctx, query, args...)
	if err := row.Scan(
		&s.ID, &s.Username, &s.PasswordHash, &s.Email,
		&s.UpdatedAt, &s.CreatedAt, &s.CompanyName, &s.Rating,
	); err != nil {
		r.logger.WithError(err).Warn("seller not found")
		return nil, appError.NewAppError("NOT_FOUND", "seller not found", appError.ErrNotFound)
	}

	r.logger.WithField("user_id", s.ID).Info("seller profile retrieved")
	return &s, nil
}
