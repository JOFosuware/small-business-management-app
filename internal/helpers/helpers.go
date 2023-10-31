package helpers

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	"math"
	"mime/multipart"
	"net/http"
	"runtime/debug"

	"github.com/jofosuware/small-business-management-app/internal/config"
	"github.com/nfnt/resize"
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

func ProcessImage(file multipart.File) ([]byte, error) {
	//Decode the file into an image.Image type
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	//Resize image
	resizedImg := resize.Resize(100, 100, img, resize.Lanczos3)

	buf := new(bytes.Buffer)
	err = png.Encode(buf, resizedImg)
	if err != nil {
		return nil, err
	}
	imgData := buf.Bytes()

	return imgData, nil
}

func ToDecimalPlace(x float64, precision int) float64 {
	multipler := math.Pow(10, float64(precision))
	rounded := math.Round(x*multipler) / multipler
	return rounded
}
