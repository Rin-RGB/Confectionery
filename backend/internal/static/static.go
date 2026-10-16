package static

import (
	"net/http"
)

// Handler возвращает статические файлы из указанной директории.
// @Summary Получение изображения начинки
// @Tags static
// @Produce png,jpeg
// @Param fileName path string true "Имя файла из image_name"
// @Success 200 {file} file
// @Failure 404
// @Router /static/fillings/{fileName} [get]
func Handler(filesFolder string) http.Handler {
	return http.FileServer(http.Dir(filesFolder))
}
