package category

import (
	"context"
	"marketplace/internal/entity"
	"marketplace/pkg/errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

const (
	tableCategories = "categories"

	errCodeBuildQuery = "BUILD_QUERY"
	errCodeExecQuery  = "EXEC_QUERY"
	errCodeScanErr    = "SCAN_ERR"
	errCodeAcquire    = "ACQUIRE_CONN"
	errCodeBeginTx    = "BEGIN_TX"
	errCodeCommitTx   = "COMMIT_TX"
	errCodeRollbackTx = "ROLLBACK_TX"
)

var categoryColums = []string{
	"id",
	"name",
	"created_at",
	"updated_at",
}

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

type categoryRepository struct {
	pool   *pgxpool.Pool
	logger *logrus.Logger
}

func NewCategoryRepository(pool *pgxpool.Pool, logger *logrus.Logger) *categoryRepository {
	return &categoryRepository{
		pool:   pool,
		logger: logger,
	}
}

func (s *categoryRepository) Create(ctx context.Context, category *entity.Category) error {
	return s.withTx(ctx, func(tx pgx.Tx) error {
		query, args, err := psql.
			Insert(tableCategories).
			Columns(categoryColums...).
			Values(
				category.ID,
				category.Name,
				category.CreatedAt,
				category.UpdatedAt,
			).
			ToSql()
		if err != nil {
			return errors.NewAppError(errCodeBuildQuery, "failed build query", err)
		}

		tag, err := tx.Exec(ctx, query, args...)
		if err != nil {
			return errors.NewAppError(errCodeExecQuery, "failed execute create query", err)
		}
		if tag.RowsAffected() == 0 {
			s.logger.WithFields(logrus.Fields{
				"operation":   "create",
				"caregory_id": category.ID,
				"query":       query,
				"args":        args,
			}).Warn("No rows affected during create")
		}
		return nil
	})
}

func (s *categoryRepository) Update(ctx context.Context, category *entity.Category) error {
	return s.withTx(ctx, func(tx pgx.Tx) error {
		query, args, err := psql.
			Update(tableCategories).
			Set("name", category.Name).
			Set("updated_at", category.UpdatedAt).
			Where(sq.Eq{"id": category.ID}).
			ToSql()
		if err != nil {
			return errors.NewAppError(errCodeBuildQuery, "failed to build query", err)
		}

		tag, err := tx.Exec(ctx, query, args...)
		if err != nil {
			return errors.NewAppError(errCodeExecQuery, "failed execute update query", err)
		}
		if tag.RowsAffected() == 0 {
			s.logger.WithFields(logrus.Fields{
				"operation":   "update",
				"category_id": category.ID,
				"query":       query,
				"args":        args,
			}).Warn("No rows affected during update")
		}

		return nil
	})
}

func (s *categoryRepository) Delete(ctx context.Context, id string) error {
	return s.withTx(ctx, func(tx pgx.Tx) error {
		query, args, err := psql.
			Delete(tableCategories).
			Where(sq.Eq{"id": id}).
			ToSql()
		if err != nil {
			return errors.NewAppError(errCodeBuildQuery, "failed build query", err)
		}

		tag, err := tx.Exec(ctx, query, args...)
		if err != nil {
			return errors.NewAppError(errCodeExecQuery, "failed execute delete query", err)
		}
		if tag.RowsAffected() == 0 {
			s.logger.WithFields(logrus.Fields{
				"operation": "delete",
				"id":        id,
				"query":     query,
				"args":      args,
			}).Warn("No rows affected during delete")
		}
		return nil
	})
}

func (s *categoryRepository) List(ctx context.Context, limit int, offset int) ([]entity.Category, error) {
	builder := psql.Select(categoryColums...).From(tableCategories).Limit(uint64(limit)).Offset(uint64(offset))

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, errors.NewAppError(errCodeBuildQuery, "failed build query", err)
	}

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"operation": "list",
			"limit":     limit,
			"offset":    offset,
			"query":     query,
			"args":      args,
			"error":     err,
		}).Error("Failed to execute list query")
		return nil, errors.NewAppError(errCodeExecQuery, "failed execute list query", err)
	}

	defer rows.Close()

	var categories []entity.Category
	for rows.Next() {
		var c entity.Category
		if err := rows.Scan(
			&c.ID,
			&c.Name,
			&c.CreatedAt,
			&c.UpdatedAt,
		); err != nil {
			s.logger.WithFields(logrus.Fields{
				"operation": "list",
				"error":     err,
			}).Error("Failed to scan query row")
			return nil, errors.NewAppError(errCodeScanErr, "failed scan query row", err)
		}
		categories = append(categories, c)
	}

	if err := rows.Err(); err != nil {
		s.logger.WithFields(logrus.Fields{
			"operation": "list",
			"error":     err,
		}).Error("Error after scanning rows")
		return nil, errors.NewAppError(errCodeScanErr, "error after scanning rows", err)
	}

	return categories, nil
}

func (s *categoryRepository) GetByID(ctx context.Context, id string) (*entity.Category, error) {
	query, args, err := psql.
		Select(categoryColums...).
		From(tableCategories).
		Where(sq.Eq{"id": id}).
		Limit(1).
		ToSql()
	if err != nil {
		return nil, errors.NewAppError(errCodeBuildQuery, "failed build query", err)
	}

	var c entity.Category
	err = s.pool.QueryRow(ctx, query, args...).Scan(
		&c.ID,
		&c.Name,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		s.logger.WithFields(logrus.Fields{
			"operation": "get_by",
			"id":        id,
			"query":     query,
			"args":      args,
			"error":     err,
		}).Error("Failed to scan query row")
		return nil, errors.NewAppError(errCodeScanErr, "failed scan query row", err)
	}

	return &c, nil
}

func (s *categoryRepository) withTx(ctx context.Context, fn func(tx pgx.Tx) error) error {
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return errors.NewAppError(errCodeAcquire, "failed to acquire connection", err)
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return errors.NewAppError(errCodeBeginTx, "failed to begin transaction", err)
	}

	if err = fn(tx); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return errors.NewAppError(errCodeRollbackTx, "failed to rollback transaction", rbErr)
		}
		return err
	}

	if cmErr := tx.Commit(ctx); cmErr != nil {
		return errors.NewAppError(errCodeCommitTx, "failed to commit transaction", cmErr)
	}

	return nil
}
