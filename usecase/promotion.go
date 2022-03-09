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

type PromotionStore interface {
	CreatePromotion(ctx context.Context, promoRequest models.PromotionCreationRequest) (*models.Promotion, error)
	GetPromoByID(ctx context.Context, promoID int64) (models.Promotion, error)
	GetAll(ctx context.Context) ([]models.Promotion, error)
	CountPromosBetweenDates(ctx context.Context, startDate, endDate time.Time) (int, error)
}

type Promotions struct {
	Store PromotionStore
}

func NewPromotions(ps store.Promotions) Promotions {
	return Promotions{
		Store: ps,
	}
}

func (p Promotions) Create(
	ctx context.Context, promoRequest models.PromotionCreationRequest) (*models.PromotionCreationResponse, error,
) {
	// validate request data
	if err := promoRequest.ValidatePromotionRequest(); err != nil {
		return nil, models.CustomError{
			Err:      fmt.Errorf("createPromo: request data is invalid: %w", err),
			HTTPCode: http.StatusBadRequest,
			Code:     "c7c007be-a97d-42df-aa28-15ef80b869b1",
		}
	}

	promosInDates, err := p.Store.CountPromosBetweenDates(ctx, promoRequest.StartDate, promoRequest.EndDate)
	if err != nil {
		return nil, models.CustomError{
			Err:      fmt.Errorf("createPromo: verifying if there are promos in the specified date range: %w", err),
			HTTPCode: http.StatusInternalServerError,
			Code:     "c4350752-49db-4546-aea9-fd3a6ed0a938",
		}
	}
	if promosInDates > 0 {
		return nil, models.CustomError{
			Err:      fmt.Errorf("createPromo: promotions already exist in the specified date range"),
			HTTPCode: http.StatusBadRequest,
			Code:     "80ed3ce8-2880-49ee-9b98-af7e91c49a38",
		}
	}

	createdPromo, err := p.Store.CreatePromotion(ctx, promoRequest)
	if err != nil {
		return nil, models.CustomError{
			Err:      fmt.Errorf("creating promotion with in the database: %w", err),
			HTTPCode: http.StatusInternalServerError,
			Code:     "c4350752-49db-4546-aea9-fd3a6ed0a938",
		}
	}

	return &models.PromotionCreationResponse{
		ID: createdPromo.ID,
	}, nil
}

func (p Promotions) Get(ctx context.Context) ([]models.Promotion, error) {
	promos, err := p.Store.GetAll(ctx)
	if err != nil {
		return nil, models.CustomError{
			Err:      fmt.Errorf("getting promotions from the database: %w", err),
			HTTPCode: http.StatusInternalServerError,
			Code:     "4dcdf08e-88be-4a2c-bbbc-ac5061c21b2b",
		}
	}

	return promos, nil
}

func (p Promotions) GetByID(ctx context.Context, promoIDParam string) (models.Promotion, error) {
	promoID, err := strconv.ParseInt(promoIDParam, 10, 64)
	if err != nil {
		return models.Promotion{}, models.CustomError{
			Err:      fmt.Errorf("invalid promotionID received: %w", err),
			HTTPCode: http.StatusBadRequest,
			Code:     "96eba417-49db-4b60-b582-e69c06aec866",
		}
	}

	promo, err := p.Store.GetPromoByID(ctx, promoID)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return models.Promotion{}, models.CustomError{
				Err:      fmt.Errorf("promotion not found in database: %w", err),
				HTTPCode: http.StatusNotFound,
				Code:     "53822d3a-1d19-453b-989e-35dc9986de4d",
			}
		}

		return models.Promotion{}, models.CustomError{
			Err:      fmt.Errorf("getting promotion from the database: %w", err),
			HTTPCode: http.StatusInternalServerError,
			Code:     "0fbea0f6-ba8a-45b1-bc24-181740f40858",
		}
	}

	return promo, nil
}
