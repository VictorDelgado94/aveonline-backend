package models

import (
	"fmt"
	"time"
)

type Promotion struct {
	ID          int64     `json:"id"`
	Description string    `json:"description"`
	Percentage  float64   `json:"percentage"`
	StartDate   time.Time `json:"startDate"`
	EndtDate    time.Time `json:"endDate"`
}

// ----------------------------------------------------------------------------
//                            VIEW MODELS
// ----------------------------------------------------------------------------

type PromotionCreationRequest struct {
	Description string    `json:"description"`
	Percentage  float64   `json:"percentage"`
	StartDate   time.Time `json:"startDate"`
	EndDate     time.Time `json:"endDate"`
}

type PromotionCreationResponse struct {
	ID int64 `json:"id"`
}

// ----------------------------------------------------------------------------
//                           VALIDATIONS
// ----------------------------------------------------------------------------

func (promoReq PromotionCreationRequest) ValidatePromotionRequest() error {
	if promoReq.Description == "" {
		return fmt.Errorf("createPromotion: promotion description is empty")
	}
	if promoReq.Percentage > float64(70) {
		return fmt.Errorf("createPromotion: invalid promotion percentage, this must be less than 70")
	}
	if promoReq.StartDate.Before(time.Now().UTC()) {
		return fmt.Errorf("createPromotion: invalid start time, this must be greater than current date")
	}
	if promoReq.EndDate.Before(time.Now().UTC()) {
		return fmt.Errorf("createPromotion: invalid end time, this must be greater than current date")
	}
	if promoReq.StartDate.After(promoReq.EndDate) {
		return fmt.Errorf("createPromotion: invalid Promotion times, Start date must be before end date")
	}

	return nil
}
