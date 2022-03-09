package models

import (
	"fmt"
	"time"
)

type BillingDetail struct {
	ID        int64      `json:"id"`
	Promotion Promotion  `json:"promotion"`
	Medicines []Medicine `json:"medicines"`
	Total     float64    `json:"total"`
	CreatedAt time.Time  `json:"createdAt"`
}

type Billing struct {
	ID        int64     `json:"id"`
	Total     float64   `json:"total"`
	CreatedAt time.Time `json:"createdAt"`
}

// ----------------------------------------------------------------------------
//                            VIEW MODELS
// ----------------------------------------------------------------------------

type BillingCreationRequest struct {
	PromotionID int64   `json:"promotionID"`
	Medicines   []int64 `json:"medicines"`
}

// ----------------------------------------------------------------------------
//                           VALIDATIONS
// ----------------------------------------------------------------------------

func (billingReq BillingCreationRequest) ValidateBillingRequest() error {
	if billingReq.PromotionID < 0 {
		return fmt.Errorf("invalid promotionID received: [%d]", billingReq.PromotionID)
	}

	for _, medicineID := range billingReq.Medicines {
		if medicineID <= 0 {
			return fmt.Errorf("invalid medicineID received: [%d]", medicineID)
		}
	}

	return nil
}
