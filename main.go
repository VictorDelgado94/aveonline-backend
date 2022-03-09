package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/VictorDelgado94/aveonline-backend/config"
	"github.com/VictorDelgado94/aveonline-backend/store"
	"github.com/VictorDelgado94/aveonline-backend/transport"
	"github.com/VictorDelgado94/aveonline-backend/usecase"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/labstack/echo/middleware"
	_ "github.com/lib/pq"
)

const (
	defaultTimeoutSeconds      = 10
	targetDBSchemaVersion uint = 1
)

func main() {
	configValues, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("error init config application: %v", err)
	}

	storeAdapter, err := store.NewStore(configValues.DatabaseURL)
	if err != nil {
		log.Fatalf("error initializing database: %v", err)
	}

	promotionsStore := store.NewPromotions(storeAdapter.GetDB())
	promotionsUsecase := usecase.NewPromotions(promotionsStore)
	promotionsTransport := transport.NewPromotions(promotionsUsecase)

	medicineStore := store.NewMedicine(storeAdapter.GetDB())
	medicinesUsecase := usecase.NewMedicines(medicineStore)
	medicinesTransport := transport.NewMedicines(medicinesUsecase)

	billingStore := store.NewBilling(storeAdapter.GetDB())
	billingUsecase := usecase.NewBillings(billingStore, promotionsStore, medicineStore)
	billingTransport := transport.NewBillings(billingUsecase)

	echoHandler := transport.NewRouter(promotionsTransport, medicinesTransport, billingTransport)

	echoHandler.Pre(middleware.RemoveTrailingSlash())
	echoHandler.Use(middleware.CORS())
	echoHandler.Use(middleware.Logger())
	echoHandler.Use(middleware.Recover())

	// Start server
	go func() {
		if err := echoHandler.Start(fmt.Sprintf(":%s", configValues.HTTPPort)); err != nil {
			echoHandler.Logger.Info("shutting down the server")
		}
	}()

	if err := storeAdapter.Migrate(targetDBSchemaVersion); err != nil {
		log.Fatalf("error in migration:  %s", err)
	}

	// Handle graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeoutSeconds*time.Second)
	defer cancel()

	if err := echoHandler.Shutdown(ctx); err != nil {
		echoHandler.Logger.Fatal(err)
	}
}
