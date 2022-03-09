package usecase

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/VictorDelgado94/aveonline-backend/models"
	"github.com/VictorDelgado94/aveonline-backend/store"
)

type MedicineStore interface {
	GetAll(ctx context.Context) ([]models.Medicine, error)
	GetMedicinesByIDs(ctx context.Context, medicineIDs []int64) ([]models.Medicine, error)
	GetMedicineByID(ctx context.Context, medicineID int64) (*models.Medicine, error)
	CreateMedicine(ctx context.Context, medicineRequest models.MedicineCreationRequest) (*models.Medicine, error)
}

type Medicines struct {
	Store MedicineStore
}

func NewMedicines(ms store.Medicine) Medicines {
	return Medicines{
		Store: ms,
	}
}

func (m Medicines) Get(ctx context.Context) ([]models.Medicine, error) {
	medicines, err := m.Store.GetAll(ctx)
	if err != nil {
		return nil, models.CustomError{
			Err:      fmt.Errorf("getting medicines from the database: %w", err),
			HTTPCode: http.StatusInternalServerError,
			Code:     "d451f6e2-ef92-44d2-9fe4-630c5272b2c7",
		}
	}

	return medicines, nil
}

func (m Medicines) GetByID(ctx context.Context, medicineIDParam string) (*models.Medicine, error) {
	medicineID, err := strconv.ParseInt(medicineIDParam, 10, 64)
	if err != nil {
		return nil, models.CustomError{
			Err:      fmt.Errorf("invalid medicineID received: %w", err),
			HTTPCode: http.StatusBadRequest,
			Code:     "0b8f7008-b19d-4360-bb58-8a3c299b962d",
		}
	}

	medicine, err := m.Store.GetMedicineByID(ctx, medicineID)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return nil, models.CustomError{
				Err:      fmt.Errorf("medicine not found in database: %w", err),
				HTTPCode: http.StatusNotFound,
				Code:     "a18e42b2-99f8-40bd-bc01-f9edd75113f7",
			}
		}

		return nil, models.CustomError{
			Err:      fmt.Errorf("getting medicine from the database: %w", err),
			HTTPCode: http.StatusInternalServerError,
			Code:     "49cb77bc-d8d4-40c4-b69a-36e59075b27c",
		}
	}

	return medicine, nil
}

func (m Medicines) Create(ctx context.Context, medicineRequest models.MedicineCreationRequest) (*models.MedicineCreationResponse, error) {
	if err := medicineRequest.ValidateMedicineRequest(); err != nil {
		return nil, models.CustomError{
			Err:      fmt.Errorf("createMedicine: request data is invalid: %w", err),
			HTTPCode: http.StatusBadRequest,
			Code:     "8c9e3533-d5d6-4e98-9402-32e71f844bce",
		}
	}

	createdMedicine, err := m.Store.CreateMedicine(ctx, medicineRequest)
	if err != nil {
		return nil, models.CustomError{
			Err:      fmt.Errorf("creating medicine with in the database: %w", err),
			HTTPCode: http.StatusInternalServerError,
			Code:     "ca6095ee-b3b5-467d-b4c9-e3822bf1ff13",
		}
	}

	return &models.MedicineCreationResponse{
		ID: createdMedicine.ID,
	}, nil
}
