package store

import (
	"errors"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/jmoiron/sqlx"
)

type Store struct {
	databaseURL string
	db          *sqlx.DB
}

func NewStore(databaseURL string) (Store, error) {
	database, err := sqlx.Connect("postgres", databaseURL)
	if err != nil {
		return Store{}, err
	}

	return Store{
		db:          database,
		databaseURL: databaseURL,
	}, nil
}

func (s *Store) Migrate(targetDBSchemaVersion uint) error {
	migrator, err := migrate.New("file://migrations/", s.databaseURL)
	if err != nil {
		return err
	}
	currentDBSchemaVersion, _, err := migrator.Version()
	if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
		return err
	}

	steps := int(targetDBSchemaVersion) - int(currentDBSchemaVersion)
	if steps > 0 {
		err = migrator.Steps(steps)
		if err != nil {
			return fmt.Errorf("failed to migrate up to DB schema version '%v' - %w", targetDBSchemaVersion, err)
		}
		log.Printf("Data migration applied to the tables in postgres")
	} else if steps < 0 {
		log.Printf(
			"The current DB schema version is '%v'. Assuming compatibility with target version '%v'. Please verify",
			currentDBSchemaVersion, targetDBSchemaVersion,
		)
	}

	return nil
}

func (s *Store) GetDB() *sqlx.DB {
	return s.db
}
