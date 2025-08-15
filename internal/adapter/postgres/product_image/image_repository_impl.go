package productimage

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
	tableProductImages = "product_images"

	errCodeBuildQuery = "BUILD_QUERY"
	errCodeExecQuery  = "EXEC_QUERY"
	errCodeScanErr    = "SCAN_ERR"
	errCodeAcquire    = "ACQUIRE_CONN"
	errCodeBeginTx    = "BEGIN_TX"
	errCodeCommitTx   = "COMMIT_TX"
	errCodeRollbackTx = "ROLLBACK_TX"
)

var productImageColums = []string{
	"id",
	"product_id",
	"url",
	"created_at",
}

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

type productImageRepository struct {
	pool   *pgxpool.Pool
	logger *logrus.Logger
}

func NewProductImageRepository(pool *pgxpool.Pool, logger *logrus.Logger) *productImageRepository {
	return &productImageRepository{
		pool:   pool,
		logger: logger,
	}
}

func (s *productImageRepository) Create(ctx context.Context, image *entity.ProductImage) error {
	return s.withTx(ctx, func(tx pgx.Tx) error {
		query, args, err := psql.
			Insert(tableProductImages).
			Columns(productImageColums...).
			Values(
				image.ID,
				image.ProductID,
				image.URL,
				image.CreatedAt,
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
				"operation":  "create",
				"image_id":   image.ID,
				"product_at": image.ProductID,
				"query":      query,
				"args":       args,
			}).Warn("No rows affected during create")
		}

		return nil
	})
}

func (s *productImageRepository) GetByID(ctx context.Context, id string) (*entity.ProductImage, error) {
	query, args, err := psql.
		Select(productImageColums...).
		From(tableProductImages).
		Where(sq.Eq{"id": id}).
		Limit(1).
		ToSql()
	if err != nil {
		return nil, errors.NewAppError(errCodeBuildQuery, "failed build query", err)
	}

	var i entity.ProductImage
	err = s.pool.QueryRow(ctx, query, args...).Scan(
		&i.ID,
		&i.ProductID,
		&i.URL,
		&i.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		s.logger.WithFields(logrus.Fields{
			"operation": "get_by_id",
			"id":        id,
			"query":     query,
			"args":      args,
			"error":     err,
		}).Error("Failed to scan query row")
		return nil, errors.NewAppError(errCodeScanErr, "failed scan query row", err)
	}

	return &i, nil
}

func (s *productImageRepository) Delete(ctx context.Context, id string) error {
	return s.withTx(ctx, func(tx pgx.Tx) error {
		query, args, err := psql.
			Delete(tableProductImages).
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

func (s *productImageRepository) ListByProductID(ctx context.Context, productID string, limit, offset int) ([]entity.ProductImage, error) {
	builder := psql.Select(productImageColums...).From(tableProductImages).Where(sq.Eq{"product_id": productID})

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, errors.NewAppError(errCodeBuildQuery, "failed build query", err)
	}

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"operation":  "list",
			"product_id": productID,
			"query":      query,
			"args":       args,
			"error":      err,
		}).Error("Failed to execute query")
		return nil, errors.NewAppError(errCodeExecQuery, "failed execute query list", err)
	}
	defer rows.Close()

	var images []entity.ProductImage
	for rows.Next() {
		var i entity.ProductImage
		if err := rows.Scan(
			&i.ID,
			&i.ProductID,
			&i.URL,
			&i.CreatedAt,
		); err != nil {
			s.logger.WithFields(logrus.Fields{
				"operation": "list",
				"error":     err,
			}).Error("Failed to scan query row")
			return nil, errors.NewAppError(errCodeScanErr, "failed scan query row", err)
		}
		images = append(images, i)
	}

	if err := rows.Err(); err != nil {
		s.logger.WithFields(logrus.Fields{
			"operation": "list",
			"error":     err,
		}).Error("Failed after scanning rows")
		return nil, errors.NewAppError(errCodeScanErr, "erroe after scanning rows", err)
	}

	return images, nil
}

func (s *productImageRepository) withTx(ctx context.Context, fn func(tx pgx.Tx) error) error {
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
