package transport

import (
	"errors"
	"net/http"

	"github.com/VictorDelgado94/aveonline-backend/models"
	"github.com/labstack/echo"
)

const (
	genericErrResponse = `{"error":"something went wrong","code":"959a1908-62f0-4cad-afc0-d9b4300085db"}`
)

// NewRouter returns a new echo.Echo struct
func NewRouter(promotionsT Promotions, medicinesT Medicines, billingsT Billings) *echo.Echo {

	e := echo.New()
	baseURL := e.Group("/aveonline/pharmacy")

	promotions := baseURL.Group("/promotions")
	promotions.GET("", promotionsT.Get)
	promotions.GET("/:promoID", promotionsT.GetByID)
	promotions.POST("", promotionsT.Create)

	medicines := baseURL.Group("/medicines")
	medicines.GET("", medicinesT.Get)
	medicines.GET("/:medicineID", medicinesT.GetByID)
	medicines.POST("", medicinesT.Create)

	billings := baseURL.Group("/billing")
	billings.GET("", billingsT.Get)
	billings.GET("/:billingID", billingsT.GetByID)
	billings.POST("", billingsT.Create)

	return e
}

func parseErrorResponse(e echo.Context, err error) error {
	var ce models.CustomError
	if errors.As(err, &ce) {
		return e.JSON(ce.HTTPCode, ce.ToResponseError())
	}

	return e.JSON(http.StatusInternalServerError, genericErrResponse)
}
