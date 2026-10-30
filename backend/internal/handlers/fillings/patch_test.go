package fillings

import (
	"bytes"
	"errors"
	"image"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	coreerrors "SPOproject/internal/core/errors"
	"github.com/go-playground/validator/v10"
)

func TestParseFillingPatch(t *testing.T) {
	var validPNG bytes.Buffer
	if err := png.Encode(&validPNG, image.NewRGBA(image.Rect(0, 0, 2, 2))); err != nil {
		t.Fatal(err)
	}
	handler := NewHandler(nil, nil, validator.New())
	for _, tc := range []struct {
		name    string
		data    []byte
		invalid bool
	}{
		{"PNG upload", validPNG.Bytes(), false},
		{"invalid file", []byte("not an image"), true},
		{"empty file", []byte{}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var body bytes.Buffer
			form := multipart.NewWriter(&body)
			if err := form.WriteField("description", ""); err != nil {
				t.Fatal(err)
			}
			if err := form.WriteField("is_active", "false"); err != nil {
				t.Fatal(err)
			}
			file, err := form.CreateFormFile("image", "replacement.png")
			if err != nil {
				t.Fatal(err)
			}
			if _, err := file.Write(tc.data); err != nil {
				t.Fatal(err)
			}
			if err := form.Close(); err != nil {
				t.Fatal(err)
			}
			request := httptest.NewRequest(http.MethodPatch, "/", &body)
			request.Header.Set("Content-Type", form.FormDataContentType())
			patch, image, err := handler.parseUpdateFillingRequest(request)
			if tc.invalid {
				if !errors.Is(err, coreerrors.ErrInvalidRequest) {
					t.Fatalf("expected invalid request, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(image, tc.data) || patch.Description == nil || *patch.Description != "" || patch.IsActive == nil || *patch.IsActive || patch.Name != nil || patch.Price != nil {
				t.Fatal("multipart patch lost uploaded bytes or optional field semantics")
			}
		})
	}
	request := httptest.NewRequest(http.MethodPatch, "/", bytes.NewBufferString(`{"name":"Updated"}`))
	request.Header.Set("Content-Type", "application/json")
	patch, uploaded, err := handler.parseUpdateFillingRequest(request)
	if err != nil || uploaded != nil || patch.Name == nil || *patch.Name != "Updated" {
		t.Fatalf("JSON compatibility failed: %v", err)
	}
}
