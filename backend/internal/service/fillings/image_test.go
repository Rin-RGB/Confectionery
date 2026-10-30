package fillings

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"SPOproject/internal/core/domain"
	coreerrors "SPOproject/internal/core/errors"
	"github.com/google/uuid"
)

type imageRepositoryStub struct {
	fillingRepository
	filling domain.Filling
	fail    error
}

func (r *imageRepositoryStub) GetFillingByID(context.Context, string) (domain.Filling, error) {
	return r.filling, nil
}

func (r *imageRepositoryStub) UpdateFilling(_ context.Context, _ string, patch domain.FillingPatch) (domain.Filling, error) {
	if r.fail != nil {
		return domain.Filling{}, r.fail
	}
	if patch.ImageName != nil {
		r.filling.ImageName = *patch.ImageName
	}
	return r.filling, nil
}

func TestReplaceFillingImage(t *testing.T) {
	for _, fail := range []bool{false, true} {
		t.Run(map[bool]string{false: "success", true: "database error"}[fail], func(t *testing.T) {
			dir := t.TempDir()
			oldPath := filepath.Join(dir, "old.png")
			if err := os.WriteFile(oldPath, []byte("old"), 0o644); err != nil {
				t.Fatal(err)
			}
			repo := &imageRepositoryStub{filling: domain.Filling{ID: uuid.NewString(), ImageName: "old.png"}}
			if fail {
				repo.fail = coreerrors.ErrFillingNotFound
			}
			service := NewService(repo, nil, dir)
			updated, err := service.UpdateFilling(context.Background(), repo.filling.ID, domain.FillingPatch{}, []byte("new"))
			if fail {
				if !errors.Is(err, repo.fail) {
					t.Fatalf("expected repository error, got %v", err)
				}
				entries, err := os.ReadDir(dir)
				if err != nil || len(entries) != 1 {
					t.Fatal("failed update left replacement file")
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				if updated.ImageName == "old.png" {
					t.Fatal("image name did not change")
				}
				data, err := os.ReadFile(filepath.Join(dir, updated.ImageName))
				if err != nil || string(data) != "new" {
					t.Fatal("new image was not saved")
				}
			}
			data, err := os.ReadFile(oldPath)
			if err != nil || string(data) != "old" {
				t.Fatal("old image was overwritten")
			}
		})
	}
}
