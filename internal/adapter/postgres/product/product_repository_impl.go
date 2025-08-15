package product

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
	tableProducts = "products"

	errCodeBuildQuery = "BUILD_QUERY"
	errCodeExecQuery  = "EXEC_QUERY"
	errCodeScanErr    = "SCAN_ERR"
	errCodeAcquire    = "ACQUIRE_CONN"
	errCodeBeginTx    = "BEGIN_TX"
	errCodeCommitTx   = "COMMIT_TX"
	errCodeRollbackTx = "ROLLBACK_TX"
)

var productColumns = []string{
	"id",
	"seller_id",
	"title",
	"description",
	"price",
	"created_at",
	"updated_at",
	"category_id",
	"is_active",
}

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

type productRepository struct {
	pool   *pgxpool.Pool
	logger *logrus.Logger
}

func NewProductRepository(pool *pgxpool.Pool, logger *logrus.Logger) *productRepository {
	return &productRepository{
		pool:   pool,
		logger: logger,
	}
}

func (s *productRepository) Create(ctx context.Context, product *entity.Product) error {
	return s.withTx(ctx, func(tx pgx.Tx) error {
		query, args, err := psql.
			Insert(tableProducts).
			Columns(productColumns...).
			Values(
				product.ID,
				product.SellerID,
				product.Title,
				product.Description,
				product.Price,
				product.CreatedAt,
				product.UpdatedAt,
				product.CategoryID,
				product.IsActive,
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
				"product_id": product.ID,
				"query":      query,
				"args":       args,
			}).Warn("No rows affected during create")
		}

		return nil
	})
}

func (s *productRepository) GetByID(ctx context.Context, id string) (*entity.Product, error) {
	return s.getBy(ctx, "id", id)
}

func (s *productRepository) GetByTitle(ctx context.Context, title string) (*entity.Product, error) {
	return s.getBy(ctx, "title", title)
}

func (s *productRepository) Update(ctx context.Context, product *entity.Product) error {
	return s.withTx(ctx, func(tx pgx.Tx) error {
		query, args, err := psql.
			Update(tableProducts).
			Set("title", product.Title).
			Set("description", product.Description).
			Set("price", product.Price).
			Set("updated_at", product.UpdatedAt).
			Set("category_id", product.CategoryID).
			Set("is_active", product.IsActive).
			Where(sq.Eq{"id": product.ID}).
			ToSql()
		if err != nil {
			return errors.NewAppError(errCodeBuildQuery, "failed build query", err)
		}

		tag, err := tx.Exec(ctx, query, args...)
		if err != nil {
			return errors.NewAppError(errCodeExecQuery, "failed execute update query", err)
		}
		if tag.RowsAffected() == 0 {
			s.logger.WithFields(logrus.Fields{
				"operation":  "update",
				"product_id": product.ID,
				"query":      query,
				"args":       args,
			}).Warn("No rows affected during update")
		}

		return nil
	})
}

func (s *productRepository) Delete(ctx context.Context, id string) error {
	return s.withTx(ctx, func(tx pgx.Tx) error {
		query, args, err := psql.
			Delete(tableProducts).
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

func (s *productRepository) List(ctx context.Context, categoryID string, limit, offset int) ([]entity.Product, error) {
	builder := psql.
		Select(productColumns...).
		From(tableProducts).
		Limit(uint64(limit)).
		Offset(uint64(offset))

	if categoryID != "" {
		builder = builder.Where(sq.Eq{"category_id": categoryID})
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, errors.NewAppError(errCodeBuildQuery, "failed build query", err)
	}

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"operation":   "list",
			"category_id": categoryID,
			"limit":       limit,
			"offset":      offset,
			"query":       query,
			"args":        args,
			"error":       err,
		}).Error("Failed to execute list query")
		return nil, errors.NewAppError(errCodeExecQuery, "failed execute list query", err)
	}
	defer rows.Close()

	var products []entity.Product
	for rows.Next() {
		var p entity.Product
		if err := rows.Scan(
			&p.ID,
			&p.SellerID,
			&p.Title,
			&p.Description,
			&p.Price,
			&p.CreatedAt,
			&p.UpdatedAt,
			&p.CategoryID,
			&p.IsActive,
		); err != nil {
			s.logger.WithFields(logrus.Fields{
				"operation": "list",
				"error":     err,
			}).Error("Failed to scan query row")
			return nil, errors.NewAppError(errCodeScanErr, "failed scan query row", err)
		}
		products = append(products, p)
	}

	if err := rows.Err(); err != nil {
		s.logger.WithFields(logrus.Fields{
			"operation": "list",
			"error":     err,
		}).Error("Error after scanning rows")
		return nil, errors.NewAppError(errCodeScanErr, "error after scanning rows", err)
	}

	return products, nil
}

func (s *productRepository) withTx(ctx context.Context, fn func(tx pgx.Tx) error) error {
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

func (s *productRepository) getBy(ctx context.Context, field string, value any) (*entity.Product, error) {
	query, args, err := psql.
		Select(productColumns...).
		From(tableProducts).
		Where(sq.Eq{field: value}).
		Limit(1).
		ToSql()
	if err != nil {
		return nil, errors.NewAppError(errCodeBuildQuery, "failed build query", err)
	}

	var p entity.Product
	err = s.pool.QueryRow(ctx, query, args...).Scan(
		&p.ID,
		&p.SellerID,
		&p.Title,
		&p.Description,
		&p.Price,
		&p.CreatedAt,
		&p.UpdatedAt,
		&p.CategoryID,
		&p.IsActive,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		s.logger.WithFields(logrus.Fields{
			"operation": "get_by",
			"field":     field,
			"value":     value,
			"query":     query,
			"args":      args,
			"error":     err,
		}).Error("Failed to scan query row")
		return nil, errors.NewAppError(errCodeScanErr, "failed scan query row", err)
	}

	return &p, nil
}
