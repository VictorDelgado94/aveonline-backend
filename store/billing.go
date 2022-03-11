package store

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/VictorDelgado94/aveonline-backend/models"
	"github.com/jmoiron/sqlx"
)

const (
	tableBilling       = "billing"
	tableBillingDetail = "billing_detail"
)

type Billing struct {
	db *sqlx.DB
}

func NewBilling(db *sqlx.DB) Billing {
	return Billing{
		db: db,
	}
}

func (b Billing) GetBillingsByDates(ctx context.Context, startDate, endDate time.Time) ([]models.Billing, error) {
	getBillingsBetweenDatesSQL := fmt.Sprintf(`
	SELECT id, total, created_at
	FROM %s
	WHERE created_at BETWEEN $1 AND $2 AND deleted_at IS NULL
	`, tableBilling)

	rows, err := b.db.QueryContext(ctx, getBillingsBetweenDatesSQL, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("error while building query: %w", err)
	}
	defer func() {
		errClose := rows.Close()
		errRows := rows.Err()
		if errClose != nil || errRows != nil {
			log.Printf("something went wrong while closing rows: %v, %v", errClose, errRows)
		}
	}()
	billings := make([]models.Billing, 0)
	for rows.Next() {
		var (
			id        int64
			total     float64
			createdAt sql.NullTime
		)
		if err := rows.Scan(&id, &total, &createdAt); err != nil {
			return nil, fmt.Errorf("error getting billings between dates: %w", err)
		}
		billings = append(
			billings,
			models.Billing{
				ID:        id,
				Total:     total,
				CreatedAt: createdAt.Time,
			},
		)
	}

	return billings, nil
}

func (b Billing) GetBillingByID(ctx context.Context, billingID int64) (*models.BillingDetail, error) {
	getBillingSQL := fmt.Sprintf(`
	SELECT promotion_id, total, created_at
	FROM %s
	WHERE id = $1 AND deleted_at IS NULL
	`, tableBilling)

	row := b.db.QueryRowContext(ctx, getBillingSQL, billingID)
	var (
		promotionID sql.NullInt64
		total       float64
		createdAt   sql.NullTime
	)
	if err := row.Scan(
		&promotionID,
		&total,
		&createdAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, models.ErrNotFound
		}
		return nil, fmt.Errorf("error reading billing row: %w", err)
	}

	billing := &models.BillingDetail{
		ID:        billingID,
		Total:     total,
		CreatedAt: createdAt.Time,
	}

	var err error
	billing.Promotion, err = b.getPromotion(ctx, promotionID.Int64)
	if err != nil {
		return nil, fmt.Errorf("error reading billing's promotion: %w", err)
	}

	billing.Medicines, err = b.getBillingDetail(ctx, billingID)
	if err != nil {
		return nil, fmt.Errorf("error reading billing's medicines: %w", err)
	}

	return billing, nil
}

func (b Billing) getPromotion(ctx context.Context, promoID int64) (models.Promotion, error) {
	if promoID <= 0 {
		return models.Promotion{}, nil
	}

	getPromoSQL := fmt.Sprintf(`
	SELECT id, description, percentage, start_date, end_date
	FROM %s
	WHERE id = $1 AND deleted_at IS NULL
	`, tablePromotions)

	row := b.db.QueryRowContext(ctx, getPromoSQL, promoID)
	var (
		id          int64
		description sql.NullString
		percentage  sql.NullFloat64
		startDate   time.Time
		endDate     time.Time
	)
	if err := row.Scan(
		&id,
		&description,
		&percentage,
		&startDate,
		&endDate,
	); err != nil {
		if err == sql.ErrNoRows {
			return models.Promotion{}, models.ErrNotFound
		}
		return models.Promotion{}, fmt.Errorf("error reading promotion row: %w", err)
	}

	return models.Promotion{
		ID:          id,
		Description: description.String,
		Percentage:  percentage.Float64,
		StartDate:   startDate,
		EndtDate:    endDate,
	}, nil
}

func (b Billing) getBillingDetail(ctx context.Context, billingID int64) ([]models.Medicine, error) {
	getBillingDetailSQL := fmt.Sprintf(`
	SELECT medicine_id, medicine_name, medicine_price
	FROM %s
	WHERE billing_id = $1 AND deleted_at IS NULL
	`, tableBillingDetail)

	rows, err := b.db.QueryContext(ctx, getBillingDetailSQL, billingID)
	if err != nil {
		return nil, fmt.Errorf("error while building query: %w", err)
	}
	defer func() {
		errClose := rows.Close()
		errRows := rows.Err()
		if errClose != nil || errRows != nil {
			log.Printf("something went wrong while closing rows: %v, %v", errClose, errRows)
		}
	}()
	medicines := make([]models.Medicine, 0)
	for rows.Next() {
		var (
			medicineID    int64
			medicineName  string
			medicinePrice float64
		)
		if err := rows.Scan(&medicineID, &medicineName, &medicinePrice); err != nil {
			return nil, fmt.Errorf("error getting billings medicines: %w", err)
		}
		medicines = append(
			medicines,
			models.Medicine{
				ID:    medicineID,
				Name:  medicineName,
				Price: medicinePrice,
			},
		)
	}

	return medicines, nil
}

func (b Billing) CreateBilling(ctx context.Context, billing models.BillingDetail) (*models.BillingDetail, error) {
	createBillingSQL := fmt.Sprintf(`
	INSERT INTO %s (promotion_id, total, created_at, updated_at)
	VALUES ($1, $2, $3, $4) RETURNING id;
	`, tableBilling)

	createBillingDetailSQL := fmt.Sprintf(`
	INSERT INTO %s (billing_id, medicine_id, medicine_name, medicine_price, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6);
	`, tableBillingDetail)

	tx, err := b.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("createBilling: could not begin transaction")
	}

	var billingID int64
	var promoID sql.NullInt64
	fmt.Println("billing.Promotion", billing.Promotion)
	if billing.Promotion.ID > 0 {
		promoID.Int64 = billing.Promotion.ID
		promoID.Valid = true
	}
	now := time.Now().UTC()
	err = tx.QueryRowContext(ctx, createBillingSQL, promoID, billing.Total, billing.CreatedAt, now).Scan(&billingID)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			return nil, fmt.Errorf("createBilling: could not rollback transaction: %w", err)
		}
		return nil, fmt.Errorf("createBilling: could not create billing within db: %w", err)
	}

	for _, m := range billing.Medicines {
		_, err := tx.ExecContext(ctx, createBillingDetailSQL, billingID, m.ID, m.Name, m.Price, now, now)
		if err != nil {
			if err := tx.Rollback(); err != nil {
				return nil, fmt.Errorf("createBilling: could not rollback transaction: %w", err)
			}
			return nil, fmt.Errorf("createBilling: could not create billing detail within db: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		if err := tx.Rollback(); err != nil {
			return nil, fmt.Errorf("createBilling: could not rollback transaction: %w", err)
		}
		return nil, fmt.Errorf("createBilling: could not commit transaction: %w", err)
	}

	return &billing, nil
}
