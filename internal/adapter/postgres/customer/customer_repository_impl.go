package customer

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

type customerRepository struct {
	pool   *pgxpool.Pool
	logger *logrus.Logger
}

func NewCustomerRepository(pool *pgxpool.Pool, logger *logrus.Logger) *customerRepository {
	return &customerRepository{pool: pool, logger: logger}
}

func (r *customerRepository) UpdateProfile(ctx context.Context, profile *entity.CustomerProfile) (err error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		r.logger.WithError(err).Error("failed to begin transaction")
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

	cQuery, cArgs, err := psql.
		Update("customers").
		Set("first_name", profile.FirstName).
		Set("last_name", profile.LastName).
		Set("phone", profile.Phone).
		Set("date_birth", profile.DateBirth).
		Set("address", profile.Address).
		Where(sq.Eq{"user_id": profile.ID}).
		ToSql()
	if err != nil {
		r.logger.WithError(err).Error("failed to build customer update query")
		return appError.NewAppError("SQL_BUILD_ERROR", "could not build customer update query", err)
	}

	if _, err = tx.Exec(ctx, cQuery, cArgs...); err != nil { // tx!
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

	r.logger.WithField("user_id", profile.ID).Info("customer profile updated successfully")
	return nil
}

func (r *customerRepository) GetByUsername(ctx context.Context, username string) (*entity.CustomerProfile, error) {
	return r.getByField(ctx, "username", username)
}

func (r *customerRepository) GetByEmail(ctx context.Context, email string) (*entity.CustomerProfile, error) {
	return r.getByField(ctx, "email", email)
}

func (r *customerRepository) getByField(ctx context.Context, field, value string) (*entity.CustomerProfile, error) {
	query, args, err := psql.
		Select(
			"u.id", "u.username", "u.password_hash", "u.email",
			"u.updated_at", "u.created_at",
			"c.first_name", "c.last_name", "c.phone", "c.date_birth", "c.address",
		).
		From("users u").
		Join("customers c ON u.id = c.user_id").
		Where(sq.Eq{fmt.Sprintf("u.%s", field): value}).
		ToSql()
	if err != nil {
		r.logger.WithError(err).Error("failed to build getByField query")
		return nil, appError.NewAppError("SQL_BUILD_ERROR", "could not build getByField query", err)
	}

	var c entity.CustomerProfile
	row := r.pool.QueryRow(ctx, query, args...)
	if err := row.Scan(
		&c.ID, &c.Username, &c.PasswordHash, &c.Email,
		&c.UpdatedAt, &c.CreatedAt,
		&c.FirstName, &c.LastName, &c.Phone, &c.DateBirth, &c.Address,
	); err != nil {
		r.logger.WithError(err).Warn("customer not found")
		return nil, appError.NewAppError("NOT_FOUND", "customer not found", appError.ErrNotFound)
	}

	r.logger.WithField("user_id", c.ID).Info("customer profile retrieved")
	return &c, nil
}
