package main

import (
	"embed"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"github.com/pressly/goose/v3"
	"github.com/robfig/cron/v3"
	"liquiswiss/config"
	"liquiswiss/internal/api"
	"liquiswiss/internal/db"
	"liquiswiss/internal/middleware"
	"liquiswiss/internal/service/db_service"
	"liquiswiss/internal/service/fixer_io_service"
	"liquiswiss/internal/service/forecast_service"
	"liquiswiss/internal/service/sendgrid_service"
	"liquiswiss/internal/service/user_service"
	"liquiswiss/pkg/logger"
	"liquiswiss/pkg/utils"
	"net/http"
	"os"
)

//go:embed internal/db/migrations/static/*.sql
var staticMigrations embed.FS

//go:embed internal/db/migrations/dynamic/*.sql
var dynamicMigrations embed.FS

func main() {
	// Init global logger
	logger.NewZapLogger(utils.IsProduction())

	// Environment for DEV
	if !utils.IsProduction() {
		err := godotenv.Load()
		if err != nil {
			logger.Logger.Errorf("Error loading .env file: %v", err)
			os.Exit(1)
		}
	}

	runApp()
}

func runApp() {
	// Global Validator
	utils.InitValidator()

	conn, err := db.Connect()
	if err != nil {
		logger.Logger.Error(err)
		os.Exit(1)
	}
	defer conn.Close()

	// Run auto migrations
	err = runStaticMigrations()
	if err != nil {
		logger.Logger.Error(err)
		os.Exit(1)
	}

	err = runDynamicMigrations()
	if err != nil {
		logger.Logger.Error(err)
		os.Exit(1)
	}

	cfg := config.GetConfig()
	dbService := db_service.NewDatabaseService(conn)
	fixerIOService := fixer_io_service.NewFixerIOService(&dbService)
	userService := user_service.NewUserServiceService(&dbService)
	sendgridService := sendgrid_service.NewSendgridService(cfg.SendgridToken)
	forecastService := forecast_service.NewForecastService(&dbService, &userService)
	middleware.InjectUserService(dbService)
	apiHandler := api.NewAPI(dbService, sendgridService, forecastService, userService)

	// Cronjob
	c := cron.New()
	_, err = c.AddFunc("@every 12h", fixerIOService.FetchFiatRates)
	if err != nil {
		logger.Logger.Errorf("Failed to set fixer.io cronjob: %v", err)
		return
	}
	c.Start()

	go func() {
		requiresInitialFetch, err := fixerIOService.RequiresInitialFetch()
		if err != nil {
			logger.Logger.Error("Error checking if initial fetch is required", err)
			return
		}
		if requiresInitialFetch {
			logger.Logger.Info("Count of fiat rate currencies doesn't match currencies, fetching from fixer.io")
			fixerIOService.FetchFiatRates()
		} else {
			logger.Logger.Info("No initial fetch required for fiat rates")
		}
	}()

	err = http.ListenAndServe(":8080", apiHandler.Router)
	if err != nil {
		c.Stop()
		logger.Logger.Error("Failed to start api:", err)
		os.Exit(1)
	}
}

func runStaticMigrations() error {
	logger.Logger.Info("Running static migrations...")

	goose.SetBaseFS(staticMigrations)

	if err := goose.SetDialect(string(goose.DialectMySQL)); err != nil {
		return errors.Wrapf(err, `failed to set goose dialect to "%s"`, goose.DialectMySQL)
	}

	gooseConn, err := db.Connect()
	if err != nil {
		return errors.Wrapf(err, "failed to connect to database as SQL for Goose")
	}

	if err := goose.Up(gooseConn, "internal/db/migrations/static"); err != nil {
		return errors.Wrapf(err, "failed to apply static migrations")
	}

	err = gooseConn.Close()
	if err != nil {
		return errors.Wrapf(err, "failed to close temporary Goose DB connection")
	}
	return nil
}

func runDynamicMigrations() error {
	logger.Logger.Info("Running dynamic migrations...")

	goose.SetBaseFS(dynamicMigrations)

	if err := goose.SetDialect(string(goose.DialectMySQL)); err != nil {
		return errors.Wrapf(err, `failed to set goose dialect to "%s"`, goose.DialectMySQL)
	}

	gooseConn, err := db.Connect()
	if err != nil {
		return errors.Wrapf(err, "failed to connect to database as SQL for Goose")
	}

	goose.SetLogger(goose.NopLogger())
	if err := goose.DownTo(gooseConn, "internal/db/migrations/dynamic", 0, goose.WithNoVersioning()); err != nil {
		return errors.Wrapf(err, "failed to revert dynamic migrations")
	}

	goose.SetLogger(logger.StdLogger{})
	if err := goose.Up(gooseConn, "internal/db/migrations/dynamic", goose.WithNoVersioning()); err != nil {
		return errors.Wrapf(err, "failed to apply dynmic migrations")
	}

	err = gooseConn.Close()
	if err != nil {
		return errors.Wrapf(err, "failed to close temporary Goose DB connection")
	}
	return nil
}
