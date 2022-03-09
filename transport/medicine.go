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
	medicineIDParam = "medicineID"
)

type MedicinesUsecase interface {
	Create(ctx context.Context, medicineRequest models.MedicineCreationRequest) (*models.MedicineCreationResponse, error)
	Get(ctx context.Context) ([]models.Medicine, error)
	GetByID(ctx context.Context, medicineID string) (*models.Medicine, error)
}

type Medicines struct {
	Usecase MedicinesUsecase
}

func NewMedicines(muc usecase.Medicines) Medicines {
	return Medicines{
		Usecase: muc,
	}
}

func (m Medicines) Get(e echo.Context) error {
	ctx := e.Request().Context()

	medicines, err := m.Usecase.Get(ctx)
	if err != nil {
		return parseErrorResponse(e, err)
	}

	return e.JSON(http.StatusOK, medicines)
}

func (m Medicines) GetByID(e echo.Context) error {
	ctx := e.Request().Context()

	medicineID := e.Param(medicineIDParam)

	medicine, err := m.Usecase.GetByID(ctx, medicineID)
	if err != nil {
		return parseErrorResponse(e, err)
	}

	return e.JSON(http.StatusOK, medicine)
}

func (m Medicines) Create(e echo.Context) error {
	ctx := e.Request().Context()

	var requestedMedicine models.MedicineCreationRequest
	if err := e.Bind(&requestedMedicine); err != nil {
		return parseErrorResponse(e, models.CustomError{
			Err:      fmt.Errorf("createMedicine: invalid medicine request body :%v", err),
			HTTPCode: http.StatusBadRequest,
			Code:     "c4b0cf4d-15a5-4724-8a24-ba89466218dd",
		})
	}

	createdMedicine, err := m.Usecase.Create(ctx, requestedMedicine)
	if err != nil {
		return parseErrorResponse(e, err)
	}

	return e.JSON(http.StatusCreated, createdMedicine)
}
