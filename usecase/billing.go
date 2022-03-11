package usecase

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/VictorDelgado94/aveonline-backend/models"
	"github.com/VictorDelgado94/aveonline-backend/store"
)

const (
	layoutDate     = "2006-01-02T15:04:05Z"
	layoutDateOnly = "2006-01-02"
)

type BillingStore interface {
	CreateBilling(ctx context.Context, billing models.BillingDetail) (*models.BillingDetail, error)
	GetBillingsByDates(ctx context.Context, startDate, endDate time.Time) ([]models.Billing, error)
	GetBillingByID(ctx context.Context, billingID int64) (*models.BillingDetail, error)
}

type Billings struct {
	Store          BillingStore
	PromotionStore PromotionStore
	MedicineStore  MedicineStore
}

func NewBillings(bs store.Billing, ps store.Promotions, ms store.Medicine) Billings {
	return Billings{
		Store:          bs,
		PromotionStore: ps,
		MedicineStore:  ms,
	}
}

func (b Billings) Get(ctx context.Context, startDate, endDate string) ([]models.Billing, error) {
	startDateTime, err := time.Parse(layoutDate, startDate)
	if err != nil {
		return nil, models.CustomError{
			Err:      fmt.Errorf("invalid start date received received: %w", err),
			HTTPCode: http.StatusBadRequest,
			Code:     "daa86901-58f7-4cb6-83be-d63ea06c0c65",
		}
	}
	endDateTime, err := time.Parse(layoutDate, endDate)
	if err != nil {
		return nil, models.CustomError{
			Err:      fmt.Errorf("invalid end date received received: %w", err),
			HTTPCode: http.StatusBadRequest,
			Code:     "da579a05-8b72-4cb5-ac7a-c2df9d0cc423",
		}
	}

	billings, err := b.Store.GetBillingsByDates(ctx, startDateTime, endDateTime)
	if err != nil {
		return nil, models.CustomError{
			Err:      fmt.Errorf("getting billing from db: %w", err),
			HTTPCode: http.StatusInternalServerError,
			Code:     "b24b5484-63c9-46ab-bf2c-24651a0fbb3e",
		}
	}

	return billings, nil
}

func (b Billings) GetByID(ctx context.Context, billingIDParam string) (*models.BillingDetail, error) {
	billingID, err := strconv.ParseInt(billingIDParam, 10, 64)
	if err != nil {
		return nil, models.CustomError{
			Err:      fmt.Errorf("invalid billingID received: %w", err),
			HTTPCode: http.StatusBadRequest,
			Code:     "2dbdcf53-2ca1-4721-8926-58eaefe84a30",
		}
	}

	billing, err := b.Store.GetBillingByID(ctx, billingID)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return nil, models.CustomError{
				Err:      fmt.Errorf("billing not found in database: %w", err),
				HTTPCode: http.StatusNotFound,
				Code:     "4a7ee994-1af6-4a35-8498-62b01007e180",
			}
		}

		return nil, models.CustomError{
			Err:      fmt.Errorf("getting billing from the database: %w", err),
			HTTPCode: http.StatusInternalServerError,
			Code:     "35efed97-6ce5-4a82-873f-d54c4af971c4",
		}
	}

	return billing, nil
}

func (b Billings) Create(ctx context.Context, billingRequest models.BillingCreationRequest) (*models.BillingDetail, error) {
	if err := billingRequest.ValidateBillingRequest(); err != nil {
		return nil, models.CustomError{
			Err:      fmt.Errorf("createBilling: request data is invalid: %w", err),
			HTTPCode: http.StatusBadRequest,
			Code:     "afbe7ecc-8cb3-4ddb-bfa9-36517ff30849",
		}
	}

	// remove repeated medicines and created a map with quantity
	uniqueMedicinesIDs, quantityMedicines := uniqueMEdicinesIDsAndSetQuantities(billingRequest.Medicines)

	promotion, medicines, err := b.getEntities(ctx, billingRequest)
	if err != nil {
		return nil, err
	}
	if len(medicines) != len(uniqueMedicinesIDs) {
		return nil, models.CustomError{
			Err:      fmt.Errorf("createBilling: not all medicines could be found, the transaction cannot be processed"),
			HTTPCode: http.StatusInternalServerError,
			Code:     "b78e227b-dbd2-4efc-a78e-a5e229488b3a",
		}
	}

	billing := b.buildBilling(ctx, promotion, medicines, quantityMedicines)
	billing.CreatedAt = billingRequest.CreatedDate

	createdBilling, err := b.Store.CreateBilling(ctx, billing)
	if err != nil {
		return nil, models.CustomError{
			Err:      fmt.Errorf("creating billing within the database: %w", err),
			HTTPCode: http.StatusInternalServerError,
			Code:     "2a68d18d-4f61-45bf-a78b-ce3ae046ddf2",
		}
	}

	return createdBilling, nil
}

func (b Billings) getEntities(ctx context.Context, billingReq models.BillingCreationRequest) (models.Promotion, []models.Medicine, error) {
	var promotion models.Promotion
	var err error
	// if promotion exists in the request, verify that it exists in the database
	if billingReq.PromotionID > 0 {
		promotion, err = b.PromotionStore.GetPromoByID(ctx, billingReq.PromotionID)
		if err != nil {
			if errors.Is(err, models.ErrNotFound) {
				return promotion, nil, models.CustomError{
					Err:      fmt.Errorf("createBilling: promotion not found in database: %w", err),
					HTTPCode: http.StatusNotFound,
					Code:     "fd1ed496-d23e-4c7b-9293-4922cc7a3b4e",
				}
			}

			return promotion, nil, models.CustomError{
				Err:      fmt.Errorf("createBilling: checking promotion in the database: %w", err),
				HTTPCode: http.StatusInternalServerError,
				Code:     "72dd1c6b-09c9-4be1-af85-2bb6950c1d12",
			}
		}
	}

	medicines, err := b.MedicineStore.GetMedicinesByIDs(ctx, billingReq.Medicines)
	if err != nil {
		return promotion, nil, models.CustomError{
			Err:      fmt.Errorf("createBilling: getting medicines from the database: %w", err),
			HTTPCode: http.StatusInternalServerError,
			Code:     "6c18183a-b5f5-43a1-95d6-9ae6f80e57e7",
		}
	}

	return promotion, medicines, nil
}

func (b Billings) buildBilling(ctx context.Context, promotion models.Promotion, medicines []models.Medicine, quantities map[int64]int) models.BillingDetail {
	billing := models.BillingDetail{}

	total := 0.0
	for _, medicine := range medicines {
		total += medicine.Price * float64(quantities[medicine.ID])
	}
	if promotion.ID > 0 {
		total = total - (total * (promotion.Percentage / 100))
	}

	billing.Promotion = promotion
	billing.Medicines = medicines
	billing.Total = total

	return billing
}

func (b Billings) Simulator(ctx context.Context, date, medicinesIDsParam string) (*models.SimulatorResponse, error) {
	dateTime, err := time.Parse(layoutDateOnly, date)
	if err != nil {
		return nil, models.CustomError{
			Err:      fmt.Errorf("invalid date received: %w", err),
			HTTPCode: http.StatusBadRequest,
			Code:     "0bd363ff-f16a-4566-9a60-edd8bd40d860",
		}
	}
	endDay := dateTime.Add(24 * time.Hour)

	if endDay.Before(time.Now().UTC()) {
		return nil, models.CustomError{
			Err:      fmt.Errorf("invalid date received: date must be after current date"),
			HTTPCode: http.StatusBadRequest,
			Code:     "efc4df93-748a-44b9-a501-602ecf6eb8d5",
		}
	}

	medicinesIDs := strings.Split(medicinesIDsParam, ",")
	mIDs := make([]int64, 0)
	for _, id := range medicinesIDs {
		medicineID, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			return nil, models.CustomError{
				Err:      fmt.Errorf("invalid medicineID received: %w", err),
				HTTPCode: http.StatusBadRequest,
				Code:     "beb548c5-f97d-4953-b850-30fcc5435346",
			}
		}
		mIDs = append(mIDs, medicineID)
	}

	uniqueMedicinesIDs, quantityMedicines := uniqueMEdicinesIDsAndSetQuantities(mIDs)

	medicines, err := b.MedicineStore.GetMedicinesByIDs(ctx, mIDs)
	if err != nil {
		return nil, models.CustomError{
			Err:      fmt.Errorf("simulator: getting medicines from the database: %w", err),
			HTTPCode: http.StatusInternalServerError,
			Code:     "c1eb357a-efc9-455b-807a-603ac01abb96",
		}
	}
	if len(medicines) != len(uniqueMedicinesIDs) {
		return nil, models.CustomError{
			Err:      fmt.Errorf("simulator: not all medicines could be found, the transaction cannot be processed"),
			HTTPCode: http.StatusInternalServerError,
			Code:     "ef8f592b-2c92-47f7-9a7f-875966a3ee07",
		}
	}

	promotion, err := b.PromotionStore.GetByDate(ctx, dateTime, endDay)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, models.CustomError{
			Err:      fmt.Errorf("simulator: getting promotion for the specific date: %w", err),
			HTTPCode: http.StatusInternalServerError,
			Code:     "d65d410c-37a3-4825-8f2c-8795f94ab413",
		}
	}
	billing := b.buildBilling(ctx, promotion, medicines, quantityMedicines)

	return &models.SimulatorResponse{
		Total: billing.Total,
	}, nil
}

func uniqueMEdicinesIDsAndSetQuantities(medicinesIDs []int64) ([]int64, map[int64]int) {
	// remove repeated medicines and created a map with quantity
	quantityMedicines := make(map[int64]int, 0)
	uniqueMedicinesIDs := make([]int64, 0)
	for _, medicineID := range medicinesIDs {
		quantityMedicines[medicineID] += 1
	}
	for medicineID := range quantityMedicines {
		uniqueMedicinesIDs = append(uniqueMedicinesIDs, medicineID)
	}

	return uniqueMedicinesIDs, quantityMedicines
}
