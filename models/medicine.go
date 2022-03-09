package models

import (
	"fmt"
	"time"
)

type Medicine struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Price     float64   `json:"price"`
	Location  string    `json:"location"`
	CreatedAt time.Time `json:"createdAt"`
}

// ----------------------------------------------------------------------------
//                            VIEW MODELS
// ----------------------------------------------------------------------------

type MedicineCreationRequest struct {
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	Location string  `json:"location"`
}

type MedicineCreationResponse struct {
	ID int64 `json:"id"`
}

// ----------------------------------------------------------------------------
//                           VALIDATIONS
// ----------------------------------------------------------------------------

func (medicineReq MedicineCreationRequest) ValidateMedicineRequest() error {
	if medicineReq.Name == "" {
		return fmt.Errorf("createMedicine: medicine name is empty")
	}
	if medicineReq.Price <= 0 {
		return fmt.Errorf("createMedicine: invalid medicine price, this must be greater than 0")
	}

	return nil
}
