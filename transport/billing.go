package transport

import (
	"context"
	"fmt"
	"net/http"

	"github.com/VictorDelgado94/aveonline-backend/models"
	"github.com/VictorDelgado94/aveonline-backend/usecase"
	"github.com/labstack/echo"
)

const (
	billingIDParam = "billingID"

	startDateQueryParam = "startDate"
	endDateQueryParam   = "endDate"
)

type BillingsUsecase interface {
	Create(ctx context.Context, billlingRequest models.BillingCreationRequest) (*models.BillingDetail, error)
	Get(ctx context.Context, startDate, endDate string) ([]models.Billing, error)
	GetByID(ctx context.Context, billingID string) (*models.BillingDetail, error)
}

type Billings struct {
	Usecase BillingsUsecase
}

func NewBillings(buc usecase.Billings) Billings {
	return Billings{
		Usecase: buc,
	}
}

func (m Billings) Create(e echo.Context) error {
	ctx := e.Request().Context()

	var requestedBilling models.BillingCreationRequest
	if err := e.Bind(&requestedBilling); err != nil {
		return parseErrorResponse(e, models.CustomError{
			Err:      fmt.Errorf("createBilling: invalid billing request body :%v", err),
			HTTPCode: http.StatusBadRequest,
			Code:     "3124d155-d315-4d0d-9732-bd2f03e9a359",
		})
	}

	createdBilling, err := m.Usecase.Create(ctx, requestedBilling)
	if err != nil {
		return parseErrorResponse(e, err)
	}

	return e.JSON(http.StatusCreated, createdBilling)
}

func (b Billings) Get(e echo.Context) error {
	ctx := e.Request().Context()

	startDate := e.QueryParam(startDateQueryParam)
	endDate := e.QueryParam(endDateQueryParam)

	billings, err := b.Usecase.Get(ctx, startDate, endDate)
	if err != nil {
		return parseErrorResponse(e, err)
	}

	return e.JSON(http.StatusOK, billings)
}

func (b Billings) GetByID(e echo.Context) error {
	ctx := e.Request().Context()

	billingID := e.Param(billingIDParam)

	billing, err := b.Usecase.GetByID(ctx, billingID)
	if err != nil {
		return parseErrorResponse(e, err)
	}

	return e.JSON(http.StatusOK, billing)
}
