package fillings

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"SPOproject/internal/core/domain"
	coreerrors "SPOproject/internal/core/errors"

	"github.com/google/uuid"
)

type Service struct {
	fillingsRepo fillingRepository
}

func NewService(fillingsRepo fillingRepository) *Service {
	return &Service{fillingsRepo: fillingsRepo}
}

func (s *Service) GetFillings(ctx context.Context) ([]domain.Filling, error) {
	fillings, err := s.fillingsRepo.GetFillings(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("list active fillings: %w", err)
	}

	return fillings, nil
}

func (s *Service) GetFillingByID(ctx context.Context, fillingID string) (domain.Filling, error) {
	if _, err := uuid.Parse(fillingID); err != nil {
		return domain.Filling{}, fmt.Errorf("invalid filling id: %w", coreerrors.ErrInvalidRequest)
	}

	filling, err := s.fillingsRepo.GetFillingByID(ctx, fillingID)
	if err != nil {
		return domain.Filling{}, fmt.Errorf("get filling: %w", err)
	}
	if !filling.IsActive {
		return domain.Filling{}, coreerrors.ErrFillingNotFound
	}

	return filling, nil
}

func (s *Service) CreateFilling(ctx context.Context, filling domain.Filling) (domain.Filling, error) {
	filling.Name = strings.TrimSpace(filling.Name)
	filling.Description = strings.TrimSpace(filling.Description)
	filling.ImageName = strings.TrimSpace(filling.ImageName)
	if filling.Name == "" || utf8.RuneCountInString(filling.Name) > 150 || filling.Price <= 0 {
		return domain.Filling{}, coreerrors.ErrInvalidRequest
	}

	createdFilling, err := s.fillingsRepo.CreateFilling(ctx, filling)
	if err != nil {
		return domain.Filling{}, fmt.Errorf("create filling: %w", err)
	}

	return createdFilling, nil
}

func (s *Service) UpdateFilling(ctx context.Context, fillingID string, patch domain.FillingPatch) (domain.Filling, error) {
	if _, err := uuid.Parse(fillingID); err != nil {
		return domain.Filling{}, fmt.Errorf("invalid filling id: %w", coreerrors.ErrInvalidRequest)
	}
	if patch.Name != nil {
		name := strings.TrimSpace(*patch.Name)
		if name == "" || utf8.RuneCountInString(name) > 150 {
			return domain.Filling{}, coreerrors.ErrInvalidRequest
		}
		patch.Name = &name
	}
	if patch.Description != nil {
		description := strings.TrimSpace(*patch.Description)
		patch.Description = &description
	}
	if patch.Price != nil && *patch.Price <= 0 {
		return domain.Filling{}, coreerrors.ErrInvalidRequest
	}
	if patch.ImageName != nil {
		imageName := strings.TrimSpace(*patch.ImageName)
		patch.ImageName = &imageName
	}

	updatedFilling, err := s.fillingsRepo.UpdateFilling(ctx, fillingID, patch)
	if err != nil {
		return domain.Filling{}, fmt.Errorf("update filling: %w", err)
	}

	return updatedFilling, nil
}

func (s *Service) HideFilling(ctx context.Context, fillingID string) error {
	if _, err := uuid.Parse(fillingID); err != nil {
		return fmt.Errorf("invalid filling id: %w", coreerrors.ErrInvalidRequest)
	}

	if err := s.fillingsRepo.HideFilling(ctx, fillingID); err != nil {
		return fmt.Errorf("hide filling: %w", err)
	}

	return nil
}
