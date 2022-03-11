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
	tableMedicine = "medicine"
)

type Medicine struct {
	db *sqlx.DB
}

func NewMedicine(db *sqlx.DB) Medicine {
	return Medicine{
		db: db,
	}
}

func (ms Medicine) GetAll(ctx context.Context) ([]models.Medicine, error) {
	getAllMedicinesSQL := fmt.Sprintf(`
	SELECT id, name, price, location, created_at
	FROM %s
	WHERE deleted_at IS NULL
	ORDER BY price asc
	`, tableMedicine)

	rows, err := ms.db.QueryContext(ctx, getAllMedicinesSQL)
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
			id        int64
			name      string
			price     float64
			location  sql.NullString
			createdAt sql.NullTime
		)
		if err := rows.Scan(&id, &name, &price, &location, &createdAt); err != nil {
			return nil, fmt.Errorf("error getting medicines: %w", err)
		}
		medicines = append(
			medicines,
			models.Medicine{
				ID:        id,
				Name:      name,
				Price:     price,
				Location:  location.String,
				CreatedAt: createdAt.Time,
			},
		)
	}

	return medicines, nil
}

func (ms Medicine) GetMedicineByID(ctx context.Context, medicineID int64) (*models.Medicine, error) {
	getMedicineSQL := fmt.Sprintf(`
	SELECT id, name, price, location, created_at
	FROM %s
	WHERE id = $1 AND deleted_at IS NULL
	`, tableMedicine)

	row := ms.db.QueryRowContext(ctx, getMedicineSQL, medicineID)

	var (
		id        int64
		name      string
		price     float64
		location  sql.NullString
		createdAt sql.NullTime
	)
	if err := row.Scan(
		&id,
		&name,
		&price,
		&location,
		&createdAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, models.ErrNotFound
		}

		return nil, fmt.Errorf("error reading medicine row: %w", err)
	}

	return &models.Medicine{
		ID:        id,
		Name:      name,
		Price:     price,
		Location:  location.String,
		CreatedAt: createdAt.Time,
	}, nil
}

func (ms Medicine) GetMedicinesByIDs(ctx context.Context, medicineIDs []int64) ([]models.Medicine, error) {
	getMedicineByIDsSQL := fmt.Sprintf(`
	SELECT id, name, price, location, created_at
	FROM %s
	WHERE id IN (?) AND deleted_at IS NULL
	`, tableMedicine)

	query, args, err := sqlx.In(getMedicineByIDsSQL, medicineIDs)
	if err != nil {
		return nil, fmt.Errorf("error building IN query: %w", err)
	}
	query = ms.db.Rebind(query)

	rows, err := ms.db.QueryContext(ctx, query, args...)
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
			id        int64
			name      string
			price     float64
			location  sql.NullString
			createdAt sql.NullTime
		)
		if err := rows.Scan(&id, &name, &price, &location, &createdAt); err != nil {
			return nil, fmt.Errorf("error getting medicines: %w", err)
		}
		medicines = append(
			medicines,
			models.Medicine{
				ID:        id,
				Name:      name,
				Price:     price,
				Location:  location.String,
				CreatedAt: createdAt.Time,
			},
		)
	}

	return medicines, nil
}

func (ms Medicine) CreateMedicine(ctx context.Context, medicineRequest models.MedicineCreationRequest) (*models.Medicine, error) {
	createMedicineSQL := fmt.Sprintf(`
	INSERT INTO %s (name, price, location, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5) RETURNING id;
	`, tableMedicine)

	now := time.Now().UTC()
	var medicineID int64
	err := ms.db.QueryRowContext(ctx, createMedicineSQL, medicineRequest.Name, medicineRequest.Price, medicineRequest.Location, now, now).Scan(&medicineID)
	if err != nil {
		return nil, fmt.Errorf("could not create medicine record within db: %w", err)
	}

	return &models.Medicine{
		ID:       medicineID,
		Name:     medicineRequest.Name,
		Price:    medicineRequest.Price,
		Location: medicineRequest.Location,
	}, nil
}
