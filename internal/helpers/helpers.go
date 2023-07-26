package helpers

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"runtime/debug"

	"github.com/jofosuware/small-business-management-app/internal/config"
)

var app *config.AppConfig

// NewHelpers sets up app config for helpers
func NewHandlers(a *config.AppConfig) {
	app = a
}

func ClientError(w http.ResponseWriter, status int) {
	app.InfoLog.Println("Client error with status of", status)
	http.Error(w, http.StatusText(status), status)
}

func ServerError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	app.ErrorLog.Println(trace)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func IsAuthenticated(r *http.Request) bool {
	exists := app.Session.Exists(r.Context(), "user_id")
	return exists
}

func FileSave(f multipart.File, h *multipart.FileHeader, fileName string) (string, error) {
	defer f.Close()
	path := filepath.Join(".", "files")
	_ = os.Mkdir(path, os.ModePerm)
	fullPath := path + "/" + fileName
	file, err := os.OpenFile(fullPath, os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		return "", err
	}

	defer file.Close()

	// Copy the file to the destination path
	_, err = io.Copy(file, f)
	if err != nil {
		return "", err
	}

	return fileName + filepath.Ext(h.Filename), nil
}
