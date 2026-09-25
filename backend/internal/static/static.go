package static

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed files
var embeddedFiles embed.FS

// Handler возвращает встроенные статические файлы.
// @Summary Получение изображения начинки
// @Tags static
// @Produce png,jpeg
// @Param fileName path string true "Имя файла из image_name"
// @Success 200 {file} file
// @Failure 404
// @Router /static/fillings/{fileName} [get]
func Handler() http.Handler {
	files, err := fs.Sub(embeddedFiles, "files")
	if err != nil {
		panic("create static files filesystem: " + err.Error())
	}

	return http.FileServer(http.FS(files))
}
