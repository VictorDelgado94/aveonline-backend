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
	tablePromotions = "promotion"
)

type Promotions struct {
	db *sqlx.DB
}

func NewPromotions(db *sqlx.DB) Promotions {
	return Promotions{
		db: db,
	}
}

func (ps Promotions) GetAll(ctx context.Context) ([]models.Promotion, error) {
	getAllPromoSQL := fmt.Sprintf(`
	SELECT id, description, percentage, start_date, end_date
	FROM %s
	WHERE deleted_at IS NULL
	ORDER BY start_date asc
	`, tablePromotions)

	rows, err := ps.db.QueryContext(ctx, getAllPromoSQL)
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
	promotions := make([]models.Promotion, 0)
	for rows.Next() {
		var (
			id          int64
			description sql.NullString
			percentage  sql.NullFloat64
			startDate   time.Time
			endDate     time.Time
		)
		if err := rows.Scan(&id, &description, &percentage, &startDate, &endDate); err != nil {
			return nil, fmt.Errorf("error getting promotions: %w", err)
		}
		promotions = append(
			promotions,
			models.Promotion{
				ID:          id,
				Description: description.String,
				Percentage:  percentage.Float64,
				StartDate:   startDate,
				EndtDate:    endDate,
			},
		)
	}

	return promotions, nil
}

func (ps Promotions) GetPromoByID(ctx context.Context, promoID int64) (models.Promotion, error) {
	getPromoSQL := fmt.Sprintf(`
	SELECT id, description, percentage, start_date, end_date
	FROM %s
	WHERE id = $1 AND deleted_at IS NULL
	`, tablePromotions)

	row := ps.db.QueryRowContext(ctx, getPromoSQL, promoID)

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

func (ps Promotions) CreatePromotion(ctx context.Context, promoRequest models.PromotionCreationRequest) (*models.Promotion, error) {
	createPromotionSQL := fmt.Sprintf(`
	INSERT INTO %s (description, percentage, start_date, end_date, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6) RETURNING id;
	`, tablePromotions)

	now := time.Now().UTC()
	var promoID int64
	err := ps.db.QueryRowContext(ctx, createPromotionSQL, promoRequest.Description, promoRequest.Percentage, promoRequest.StartDate, promoRequest.EndDate, now, now).Scan(&promoID)
	if err != nil {
		return nil, fmt.Errorf("could not create promotion within db: %w", err)
	}

	return &models.Promotion{
		ID:          promoID,
		Description: promoRequest.Description,
		Percentage:  promoRequest.Percentage,
		StartDate:   promoRequest.StartDate,
		EndtDate:    promoRequest.EndDate,
	}, nil
}

func (ps Promotions) CountPromosBetweenDates(ctx context.Context, startDate, endDate time.Time) (int, error) {
	countPromoBetweenDatesSQL := fmt.Sprintf(`
	SELECT COUNT(*)
	FROM %s
	WHERE start_date BETWEEN $1 AND $2
	OR end_date BETWEEN $3 AND $4
	AND deleted_at IS NULL
	`, tablePromotions)

	totalPromos := 0
	err := ps.db.QueryRowContext(ctx, countPromoBetweenDatesSQL, startDate, endDate, startDate, endDate).Scan(&totalPromos)
	if err != nil {
		return 0, fmt.Errorf("could not count promotions between dates db: %w", err)
	}

	return totalPromos, nil
}

func (ps Promotions) GetByDate(ctx context.Context, date, endDay time.Time) (models.Promotion, error) {
	getPromoByDateSQL := fmt.Sprintf(`
	SELECT id, description, percentage, start_date, end_date
	FROM %s
	WHERE start_date BETWEEN $1 AND $2
	OR end_date BETWEEN $3 AND $4
	AND deleted_at IS NULL
	`, tablePromotions)

	row := ps.db.QueryRowContext(ctx, getPromoByDateSQL, date, endDay, date, endDay)
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
