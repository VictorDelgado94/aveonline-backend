package usecase

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/VictorDelgado94/aveonline-backend/models"
	"github.com/VictorDelgado94/aveonline-backend/store"
)

const (
	layoutDate = "2006-01-02T15:04:05Z"
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

	promotion, medicines, err := b.getEntities(ctx, billingRequest)
	if err != nil {
		return nil, err
	}

	billing := b.buildBilling(ctx, promotion, medicines)

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
	if len(medicines) != len(billingReq.Medicines) {
		return promotion, nil, models.CustomError{
			Err:      fmt.Errorf("createBilling: not all medicines could be found, the transaction cannot be processed"),
			HTTPCode: http.StatusInternalServerError,
			Code:     "b78e227b-dbd2-4efc-a78e-a5e229488b3a",
		}
	}

	return promotion, medicines, nil
}

func (b Billings) buildBilling(ctx context.Context, promotion models.Promotion, medicines []models.Medicine) models.BillingDetail {
	billing := models.BillingDetail{}

	total := 0.0
	for _, medicine := range medicines {
		total += medicine.Price
	}
	if promotion.ID > 0 {
		total = total - (total * (promotion.Percentage / 100))
	}

	billing.Promotion = promotion
	billing.Medicines = medicines
	billing.Total = total

	return billing
}