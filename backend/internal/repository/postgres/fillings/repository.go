package repository_fillings

import (
	"context"
	"errors"
	"fmt"

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

func (r *Repository) GetFillings(ctx context.Context, onlyActive bool) ([]domain.Filling, error) {
	queryContext, cancel := context.WithTimeout(ctx, r.txManager.GetTimeout())
	defer cancel()

	const query = `
		SELECT id, name, COALESCE(description, ''), price_per_kg,
			COALESCE(image_name, ''), is_active
		FROM fillings
		WHERE NOT $1 OR is_active = TRUE
		ORDER BY created_at DESC
	`
	rows, err := r.txManager.GetExecutor(queryContext).Query(queryContext, query, onlyActive)
	if err != nil {
		return nil, fmt.Errorf("select fillings: %w", err)
	}
	defer rows.Close()

	fillings := make([]domain.Filling, 0)
	for rows.Next() {
		var filling domain.Filling
		if err = rows.Scan(
			&filling.ID,
			&filling.Name,
			&filling.Description,
			&filling.Price,
			&filling.ImageName,
			&filling.IsActive,
		); err != nil {
			return nil, fmt.Errorf("scan filling: %w", err)
		}
		fillings = append(fillings, filling)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate fillings: %w", err)
	}

	return fillings, nil
}

func (r *Repository) GetFillingByID(ctx context.Context, fillingID string) (domain.Filling, error) {
	queryContext, cancel := context.WithTimeout(ctx, r.txManager.GetTimeout())
	defer cancel()

	const query = `
		SELECT id, name, COALESCE(description, ''), price_per_kg,
		COALESCE(image_name, ''), is_active
		FROM fillings
		WHERE id = $1
	`
	var filling domain.Filling
	if err := r.txManager.GetExecutor(queryContext).QueryRow(
		queryContext,
		query,
		fillingID,
	).Scan(
		&filling.ID,
		&filling.Name,
		&filling.Description,
		&filling.Price,
		&filling.ImageName,
		&filling.IsActive,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Filling{}, coreerrors.ErrFillingNotFound
		}
		return domain.Filling{}, fmt.Errorf("select filling by id: %w", err)
	}

	return filling, nil
}

func (r *Repository) CreateFilling(ctx context.Context, filling domain.Filling) (domain.Filling, error) {
	queryContext, cancel := context.WithTimeout(ctx, r.txManager.GetTimeout())
	defer cancel()

	const query = `
		INSERT INTO fillings (name, description, price_per_kg, image_name)
		VALUES ($1, NULLIF($2, ''), $3, NULLIF($4, ''))
		RETURNING id, name, COALESCE(description, ''), price_per_kg,
			COALESCE(image_name, ''), is_active
	`
	if err := r.txManager.GetExecutor(queryContext).QueryRow(
		queryContext,
		query,
		filling.Name,
		filling.Description,
		filling.Price,
		filling.ImageName,
	).Scan(
		&filling.ID,
		&filling.Name,
		&filling.Description,
		&filling.Price,
		&filling.ImageName,
		&filling.IsActive,
	); err != nil {
		return domain.Filling{}, fmt.Errorf("insert filling: %w", err)
	}

	return filling, nil
}

func (r *Repository) UpdateFilling(ctx context.Context, fillingID string, patch domain.FillingPatch) (domain.Filling, error) {
	queryContext, cancel := context.WithTimeout(ctx, r.txManager.GetTimeout())
	defer cancel()

	const query = `
		UPDATE fillings
		SET name = COALESCE($2::VARCHAR, name),
			description = CASE
				WHEN $3::TEXT IS NULL THEN description
				ELSE NULLIF($3::TEXT, '')
			END,
			price_per_kg = COALESCE($4::NUMERIC, price_per_kg),
			image_name = CASE
				WHEN $5::TEXT IS NULL THEN image_name
				ELSE NULLIF($5::TEXT, '')
			END,
			is_active = COALESCE($6::BOOLEAN, is_active)
		WHERE id = $1
		RETURNING id, name, COALESCE(description, ''), price_per_kg,
			COALESCE(image_name, ''), is_active
	`
	var filling domain.Filling
	if err := r.txManager.GetExecutor(queryContext).QueryRow(
		queryContext,
		query,
		fillingID,
		patch.Name,
		patch.Description,
		patch.Price,
		patch.ImageName,
		patch.IsActive,
	).Scan(
		&filling.ID,
		&filling.Name,
		&filling.Description,
		&filling.Price,
		&filling.ImageName,
		&filling.IsActive,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Filling{}, coreerrors.ErrFillingNotFound
		}
		return domain.Filling{}, fmt.Errorf("update filling: %w", err)
	}

	return filling, nil
}

func (r *Repository) HideFilling(ctx context.Context, fillingID string) error {
	queryContext, cancel := context.WithTimeout(ctx, r.txManager.GetTimeout())
	defer cancel()

	const query = `
		UPDATE fillings
		SET is_active = FALSE
		WHERE id = $1
			AND is_active = TRUE
	`
	commandTag, err := r.txManager.GetExecutor(queryContext).Exec(queryContext, query, fillingID)
	if err != nil {
		return fmt.Errorf("hide filling: %w", err)
	}
	if commandTag.RowsAffected() != 1 {
		return coreerrors.ErrFillingNotFound
	}

	return nil
}
