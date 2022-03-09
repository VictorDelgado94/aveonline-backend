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
	promoIDParam = "promoID"
)

type PromotionsUsecase interface {
	Create(ctx context.Context, promotionRequest models.PromotionCreationRequest) (*models.PromotionCreationResponse, error)
	GetByID(ctx context.Context, promoID string) (models.Promotion, error)
	Get(ctx context.Context) ([]models.Promotion, error)
}

type Promotions struct {
	Usecase PromotionsUsecase
}

func NewPromotions(puc usecase.Promotions) Promotions {
	return Promotions{
		Usecase: puc,
	}
}

func (p Promotions) Create(e echo.Context) error {
	ctx := e.Request().Context()

	var requestedPromotion models.PromotionCreationRequest
	if err := e.Bind(&requestedPromotion); err != nil {
		return parseErrorResponse(e, models.CustomError{
			Err:      fmt.Errorf("createPromotion: invalid promotion request body :%v", err),
			HTTPCode: http.StatusBadRequest,
			Code:     "fd17f12a-7453-4a99-8cac-6f4b7b61d7f0",
		})
	}

	createdPromotion, err := p.Usecase.Create(ctx, requestedPromotion)
	if err != nil {
		return parseErrorResponse(e, err)
	}

	return e.JSON(http.StatusCreated, createdPromotion)
}

func (p Promotions) Get(e echo.Context) error {
	ctx := e.Request().Context()

	promotions, err := p.Usecase.Get(ctx)
	if err != nil {
		return parseErrorResponse(e, err)
	}

	return e.JSON(http.StatusOK, promotions)
}

func (p Promotions) GetByID(e echo.Context) error {
	ctx := e.Request().Context()

	promoID := e.Param(promoIDParam)

	promo, err := p.Usecase.GetByID(ctx, promoID)
	if err != nil {
		return parseErrorResponse(e, err)
	}

	return e.JSON(http.StatusOK, promo)
}
