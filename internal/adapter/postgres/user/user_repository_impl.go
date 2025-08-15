package user

import (
	"context"
	"errors"
	"fmt"
	"marketplace/internal/entity"
	appError "marketplace/pkg/errors"
	"strings"

	"github.com/jackc/pgx/v5"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

type userRepository struct {
	pool   *pgxpool.Pool
	logger *logrus.Logger
}

func NewUserRepository(pool *pgxpool.Pool, logger *logrus.Logger) *userRepository {
	return &userRepository{
		pool:   pool,
		logger: logger,
	}
}

func (r *userRepository) Create(ctx context.Context, user *entity.User) (err error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		r.logger.WithError(err).Error("failed to begin transaction")
		return appError.NewAppError("TX_BEGIN_FAIL", "could not start DB transaction", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
			r.logger.WithError(err).Warn("transaction rolled back")
		}
	}()

	query, args, err := psql.
		Insert("users").
		Columns("id", "user_type", "username", "password_hash", "email", "created_at", "updated_at").
		Values(user.ID, user.UserType, user.Username, user.PasswordHash, user.Email, user.CreatedAt, user.UpdatedAt).
		ToSql()
	if err != nil {
		r.logger.WithError(err).Error("failed to build insert query for users")
		return appError.NewAppError("SQL_BUILD_ERROR", "could not build insert query for users", err)
	}

	res, err := tx.Exec(ctx, query, args...)
	if err != nil {
		r.logger.WithError(err).Error("failed to execute insert query for users")
		return appError.NewAppError("EXEC_ERROR", "could not execute insert query for users", err)
	}
	if res.RowsAffected() == 0 {
		r.logger.Warn("insert users affected 0 rows")
		return appError.NewAppError("NOT_CREATED", "user insert returned 0 affected rows", appError.ErrNotFound)
	}

	var q2 string
	var a2 []interface{}
	switch strings.ToLower(user.UserType) {
	case "customer":
		q2, a2, err = psql.Insert("customers").Columns("user_id").Values(user.ID).ToSql()
	case "seller":
		q2, a2, err = psql.Insert("sellers").Columns("user_id").Values(user.ID).ToSql()
	default:
		msg := fmt.Sprintf("unsupported user_type: %s", user.UserType)
		r.logger.Warn(msg)
		return appError.NewAppError("INVALID_TYPE", msg, nil)
	}

	if err != nil {
		r.logger.WithError(err).Error("failed to build insert query for user subtype")
		return appError.NewAppError("SQL_BUILD_ERROR", "could not build subtype insert query", err)
	}

	res2, err := tx.Exec(ctx, q2, a2...)
	if err != nil {
		r.logger.WithError(err).Error("failed to execute subtype insert")
		return appError.NewAppError("EXEC_ERROR", "could not execute subtype insert", err)
	}
	if res2.RowsAffected() == 0 {
		r.logger.Warn("insert user subtype affected 0 rows")
		return appError.NewAppError("NOT_CREATED", "subtype insert returned 0 affected rows", appError.ErrNotFound)
	}

	if err = tx.Commit(ctx); err != nil {
		r.logger.WithError(err).Error("failed to commit transaction")
		return appError.NewAppError("TX_COMMIT_FAIL", "could not commit transaction", err)
	}

	r.logger.WithFields(logrus.Fields{
		"user_id": user.ID,
		"type":    user.UserType,
	}).Info("user created successfully")

	return nil
}

func (r *userRepository) GetByID(ctx context.Context, userID string) (*entity.User, error) {
	query, args, err := psql.
		Select("id", "user_type", "username", "password_hash", "email", "created_at", "updated_at").
		From("users").
		Where(sq.Eq{"id": userID}).
		ToSql()
	if err != nil {
		r.logger.WithError(err).Error("failed to build select query for user by id")
		return nil, appError.NewAppError("SQL_BUILD_ERROR", "could not build select query for user by id", err)
	}

	var u entity.User
	err = r.pool.QueryRow(ctx, query, args...).Scan(
		&u.ID,
		&u.UserType,
		&u.Username,
		&u.PasswordHash,
		&u.Email,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.logger.WithField("user_id", userID).Warn("user not found by id")
			return nil, appError.NewAppError("NOT_FOUND", "user not found", appError.ErrNotFound)
		}
		r.logger.WithError(err).Error("failed to execute select query for user by id")
		return nil, appError.NewAppError("EXEC_ERROR", "could not execute select query for user by id", err)
	}

	return &u, nil
}

func (r *userRepository) UpdateAuth(ctx context.Context, id string, username, email, password string) error {

	query, args, err := psql.
		Update("users").
		Set("username", username).
		Set("password_hash", password).
		Set("email", email).
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		r.logger.WithError(err).Error("failed to build update query")
		return appError.NewAppError("SQL_BUILD_ERROR", "could not build update query", err)
	}

	res, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		r.logger.WithError(err).Error("failed to execute update query")
		return appError.NewAppError("EXEC_ERROR", "could not execute update query", err)
	}

	if res.RowsAffected() == 0 {
		r.logger.Warn("update affected 0 rows")
		return appError.NewAppError("NOT_UPDATED", "update returned 0 affected rows", appError.ErrNotFound)
	}

	r.logger.WithField("user_id", id).Info("user auth updated successfully")
	return nil
}

func (r *userRepository) Delete(ctx context.Context, id string) (err error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		r.logger.WithError(err).Error("failed to begin delete transaction")
		return appError.NewAppError("TX_BEGIN_FAIL", "could not begin delete transaction", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
			r.logger.WithError(err).Warn("delete transaction rolled back")
		}
	}()

	query, args, err := psql.Delete("users").Where(sq.Eq{"id": id}).ToSql()
	if err != nil {
		r.logger.WithError(err).Error("failed to build delete query")
		return appError.NewAppError("SQL_BUILD_ERROR", "could not build delete query", err)
	}

	res, err := tx.Exec(ctx, query, args...)
	if err != nil {
		r.logger.WithError(err).Error("failed to execute delete query")
		return appError.NewAppError("EXEC_ERROR", "could not execute delete query", err)
	}
	if res.RowsAffected() == 0 {
		r.logger.Warn("delete affected 0 rows")
		return appError.NewAppError("NOT_DELETED", "delete returned 0 affected rows", appError.ErrNotFound)
	}

	if err = tx.Commit(ctx); err != nil {
		r.logger.WithError(err).Error("failed to commit delete transaction")
		return appError.NewAppError("TX_COMMIT_FAIL", "could not commit delete transaction", err)
	}

	r.logger.WithField("user_id", id).Info("user deleted successfully")
	return nil
}
