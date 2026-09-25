package repository_orders

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"SPOproject/internal/core/domain"
	coreerrors "SPOproject/internal/core/errors"
	"SPOproject/internal/repository"

	"github.com/jackc/pgx/v5"
)

type Repository struct {
	txManager repository.TxManager
}

func NewRepository(txManager repository.TxManager) *Repository {
	return &Repository{txManager: txManager}
}

func (r *Repository) CreateOrder(ctx context.Context, order domain.Order) (domain.Order, error) {
	queryContext, cancel := context.WithTimeout(ctx, r.txManager.GetTimeout())
	defer cancel()

	const query = `
		INSERT INTO orders (
			user_id, idempotency_key, weight_grams, decoration_wishes,
			delivery_address, total_price, status
		)
		VALUES ($1, $2, $3, NULLIF($4, ''), $5, $6, $7)
		ON CONFLICT (user_id, idempotency_key) DO UPDATE
		SET idempotency_key = EXCLUDED.idempotency_key
		RETURNING id, user_id, weight_grams, COALESCE(decoration_wishes, ''),
			delivery_address, total_price, status, created_at, updated_at
	`
	if err := r.txManager.GetExecutor(queryContext).QueryRow(
		queryContext,
		query,
		order.UserID,
		order.IdempotencyKey,
		order.WeightGrams,
		order.DecorationWishes,
		order.DeliveryAddress,
		order.TotalPrice,
		order.Status,
	).Scan(
		&order.ID,
		&order.UserID,
		&order.WeightGrams,
		&order.DecorationWishes,
		&order.DeliveryAddress,
		&order.TotalPrice,
		&order.Status,
		&order.CreatedAt,
		&order.UpdatedAt,
	); err != nil {
		return domain.Order{}, fmt.Errorf("insert order: %w", err)
	}

	return order, nil
}

func (r *Repository) CreateOrderFillings(
	ctx context.Context,
	orderID string,
	fillings []domain.OrderFilling,
) error {
	queryContext, cancel := context.WithTimeout(ctx, r.txManager.GetTimeout())
	defer cancel()

	fillingIDs := make([]string, len(fillings))
	fillingNames := make([]string, len(fillings))
	pricesPerKG := make([]string, len(fillings))
	weightsGrams := make([]int32, len(fillings))
	for i := range fillings {
		fillingIDs[i] = fillings[i].ID
		fillingNames[i] = fillings[i].Name
		pricesPerKG[i] = strconv.FormatFloat(fillings[i].PricePerKG, 'f', 2, 64)
		weightsGrams[i] = int32(fillings[i].WeightGrams)
	}

	const query = `
		INSERT INTO order_fillings (
			order_id, filling_id, filling_name, filling_price_per_kg, weight_grams
		)
		SELECT
			$1::UUID,
			filling_id::UUID,
			filling_name,
			price_per_kg::NUMERIC,
			weight_grams
		FROM unnest(
			$2::TEXT[],
			$3::TEXT[],
			$4::TEXT[],
			$5::INTEGER[]
		) AS filling(filling_id, filling_name, price_per_kg, weight_grams)
	`
	commandTag, err := r.txManager.GetExecutor(queryContext).Exec(
		queryContext,
		query,
		orderID,
		fillingIDs,
		fillingNames,
		pricesPerKG,
		weightsGrams,
	)
	if err != nil {
		return fmt.Errorf("insert order fillings: %w", err)
	}
	if commandTag.RowsAffected() != int64(len(fillings)) {
		return fmt.Errorf(
			"inserted %d of %d order fillings",
			commandTag.RowsAffected(),
			len(fillings),
		)
	}

	return nil
}

func (r *Repository) GetOrderByID(ctx context.Context, orderID string) (domain.Order, error) {
	queryContext, cancel := context.WithTimeout(ctx, r.txManager.GetTimeout())
	defer cancel()

	const orderQuery = `
		SELECT id, user_id, weight_grams, COALESCE(decoration_wishes, ''),
			delivery_address, total_price, status, created_at, updated_at
		FROM orders
		WHERE id = $1
	`
	var order domain.Order
	if err := r.txManager.GetExecutor(queryContext).QueryRow(
		queryContext,
		orderQuery,
		orderID,
	).Scan(
		&order.ID,
		&order.UserID,
		&order.WeightGrams,
		&order.DecorationWishes,
		&order.DeliveryAddress,
		&order.TotalPrice,
		&order.Status,
		&order.CreatedAt,
		&order.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Order{}, coreerrors.ErrOrderNotFound
		}
		return domain.Order{}, fmt.Errorf("select order by id: %w", err)
	}

	return order, nil
}

func (r *Repository) GetOrderFillings(ctx context.Context, orderID string) ([]domain.OrderFilling, error) {
	queryContext, cancel := context.WithTimeout(ctx, r.txManager.GetTimeout())
	defer cancel()

	const query = `
		SELECT filling_id, filling_name, filling_price_per_kg, weight_grams
		FROM order_fillings
		WHERE order_id = $1
		ORDER BY filling_name
	`
	rows, err := r.txManager.GetExecutor(queryContext).Query(queryContext, query, orderID)
	if err != nil {
		return nil, fmt.Errorf("select order fillings: %w", err)
	}
	defer rows.Close()

	fillings := make([]domain.OrderFilling, 0)
	for rows.Next() {
		var filling domain.OrderFilling
		if err = rows.Scan(
			&filling.ID,
			&filling.Name,
			&filling.PricePerKG,
			&filling.WeightGrams,
		); err != nil {
			return nil, fmt.Errorf("scan order filling: %w", err)
		}
		fillings = append(fillings, filling)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate order fillings: %w", err)
	}

	return fillings, nil
}

func (r *Repository) GetOrdersByUser(ctx context.Context, userID string) ([]domain.Order, error) {
	queryContext, cancel := context.WithTimeout(ctx, r.txManager.GetTimeout())
	defer cancel()

	const query = `
		SELECT id, total_price, status
		FROM orders
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.txManager.GetExecutor(queryContext).Query(queryContext, query, userID)
	if err != nil {
		return nil, fmt.Errorf("select user orders: %w", err)
	}
	defer rows.Close()

	orders := make([]domain.Order, 0)
	for rows.Next() {
		var order domain.Order
		if err = rows.Scan(&order.ID, &order.TotalPrice, &order.Status); err != nil {
			return nil, fmt.Errorf("scan user order: %w", err)
		}
		orders = append(orders, order)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate user orders: %w", err)
	}

	return orders, nil
}

func (r *Repository) GetOrders(ctx context.Context, status *domain.OrderStatus) ([]domain.Order, error) {
	queryContext, cancel := context.WithTimeout(ctx, r.txManager.GetTimeout())
	defer cancel()

	const query = `
		SELECT id, total_price, status
		FROM orders
		WHERE $1::TEXT IS NULL OR status = $1
		ORDER BY created_at DESC
	`
	var statusValue any
	if status != nil {
		statusValue = string(*status)
	}

	rows, err := r.txManager.GetExecutor(queryContext).Query(queryContext, query, statusValue)
	if err != nil {
		return nil, fmt.Errorf("select orders: %w", err)
	}
	defer rows.Close()

	orders := make([]domain.Order, 0)
	for rows.Next() {
		var order domain.Order
		if err = rows.Scan(&order.ID, &order.TotalPrice, &order.Status); err != nil {
			return nil, fmt.Errorf("scan order: %w", err)
		}
		orders = append(orders, order)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate orders: %w", err)
	}

	return orders, nil
}

func (r *Repository) UpdateStatus(ctx context.Context, orderID string, status domain.OrderStatus) error {
	queryContext, cancel := context.WithTimeout(ctx, r.txManager.GetTimeout())
	defer cancel()

	const query = `
		UPDATE orders
		SET status = $2,
			updated_at = now()
		WHERE id = $1
	`
	commandTag, err := r.txManager.GetExecutor(queryContext).Exec(queryContext, query, orderID, status)
	if err != nil {
		return fmt.Errorf("update order status: %w", err)
	}
	if commandTag.RowsAffected() != 1 {
		return coreerrors.ErrOrderNotFound
	}

	return nil
}
