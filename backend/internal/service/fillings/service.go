package fillings

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"SPOproject/internal/core/domain"
	coreerrors "SPOproject/internal/core/errors"

	"github.com/google/uuid"
)

type Service struct {
	fillingsRepo    fillingRepository
	txManager       txManager
	imagesDirectory string
}

func NewService(fillingsRepo fillingRepository, txManager txManager, imagesDirectory string) *Service {
	return &Service{
		fillingsRepo:    fillingsRepo,
		txManager:       txManager,
		imagesDirectory: imagesDirectory,
	}
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

func (s *Service) CreateFilling(
	ctx context.Context,
	filling domain.Filling,
	image []byte,
) (domain.Filling, error) {
	filling.Name = strings.TrimSpace(filling.Name)
	filling.Description = strings.TrimSpace(filling.Description)
	if filling.Name == "" ||
		len(filling.Name) > 150 ||
		filling.Price <= 0 ||
		len(image) == 0 {
		return domain.Filling{}, coreerrors.ErrInvalidRequest
	}

	filling.ID = uuid.NewString()
	filling.ImageName = filling.ID + ".png"
	imagePath := filepath.Join(s.imagesDirectory, filling.ImageName)
	imageCreated := false

	var createdFilling domain.Filling
	err := s.txManager.WithinTx(ctx, func(txCtx context.Context) error {
		var err error
		createdFilling, err = s.fillingsRepo.CreateFilling(txCtx, filling)
		if err != nil {
			return fmt.Errorf("create filling in database: %w", err)
		}

		if err = os.MkdirAll(s.imagesDirectory, 0o755); err != nil {
			return fmt.Errorf("create fillings image directory: %w", err)
		}

		imageFile, err := os.OpenFile(imagePath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if err != nil {
			return fmt.Errorf("create filling image: %w", err)
		}
		imageCreated = true

		if _, err = imageFile.Write(image); err != nil {
			closeErr := imageFile.Close()
			return errors.Join(fmt.Errorf("write filling image: %w", err), closeErr)
		}
		if err = imageFile.Close(); err != nil {
			return fmt.Errorf("close filling image: %w", err)
		}

		return nil
	})
	if err != nil {
		if imageCreated {
			err = errors.Join(err, removeImage(imagePath))
		}
		return domain.Filling{}, fmt.Errorf("create filling: %w", err)
	}

	return createdFilling, nil
}

func removeImage(imagePath string) error {
	if err := os.Remove(imagePath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("remove filling image after rollback: %w", err)
	}
	return nil
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
