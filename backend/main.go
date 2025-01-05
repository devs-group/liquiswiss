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
	"liquiswiss/pkg/logger"
	"liquiswiss/pkg/utils"
	"net/http"
	"os"
	"strconv"
)

//go:embed internal/db/migrations/*.sql
var embedMigrations embed.FS

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

	args := os.Args
	if len(args) > 1 {
		switch args[1] {
		case "goose-down-to-and-up":
			if len(args) != 3 {
				logger.Logger.Info("Usage: liquiswiss goose-down-to-and-up [version]")
				os.Exit(1)
			}
			version, err := strconv.Atoi(args[2])
			if err != nil {
				logger.Logger.Errorf("Invalid version: %v\n", err)
				os.Exit(1)
			}
			err = runGooseDownToAndUp(version)
			if err != nil {
				logger.Logger.Error(err)
				os.Exit(1)
			}
			os.Exit(0)
		default:
			logger.Logger.Errorf("Unknown command: %s\n", args[1])
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
		logger.Logger.Error("Failed to connect to database:", err)
		os.Exit(1)
	}
	defer conn.Close()

	// Do automigrations
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("mysql"); err != nil {
		logger.Logger.Errorf("Failed to setup validator: %v", err)
		panic(err)
	}

	gooseConn, err := db.Connect()
	if err != nil {
		logger.Logger.Errorf("Failed to connect to database as SQL for Goose: %v", err)
		panic(err)
	}

	if err := goose.Up(gooseConn, "internal/db/migrations"); err != nil {
		logger.Logger.Errorf("Failed to apply migrations: %v", err.Error())
		panic(err)
	}
	err = gooseConn.Close()
	if err != nil {
		logger.Logger.Errorf("Failed to close temporary Goose DB connection: %v", err)
		panic(err)
	}

	cfg := config.GetConfig()
	dbService := db_service.NewDatabaseService(conn)
	fixerIOService := fixer_io_service.NewFixerIOService(&dbService)
	sendgridService := sendgrid_service.NewSendgridService(cfg.SendgridToken)
	forecastService := forecast_service.NewForecastService(&dbService)
	middleware.InjectUserService(dbService)
	apiHandler := api.NewAPI(dbService, sendgridService, forecastService)

	// Cronjob
	c := cron.New()
	_, err = c.AddFunc("@every 30m", fixerIOService.FetchFiatRates)
	if err != nil {
		logger.Logger.Errorf("Failed to set fixer.io cronjob: %v", err)
		return
	}
	c.Start()

	go fixerIOService.FetchFiatRates()

	err = http.ListenAndServe(":8080", apiHandler.Router)
	if err != nil {
		c.Stop()
		logger.Logger.Error("Failed to start api:", err)
		os.Exit(1)
	}
}

func runGooseDownToAndUp(version int) error {
	conn, err := db.Connect()
	if err != nil {
		return errors.Wrapf(err, "Failed to connect to database: %v", err)
	}
	defer conn.Close()

	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("mysql"); err != nil {
		return errors.Wrapf(err, "Failed to set Goose dialect: %v", err)
	}

	// Run Goose Down to the specified version
	if err := goose.DownTo(conn, "internal/db/migrations", int64(version)); err != nil {
		return errors.Wrapf(err, "Failed to run goose down-to: %v", err)
	}

	// Run Goose Up to apply all migrations
	if err := goose.Up(conn, "internal/db/migrations"); err != nil {
		return errors.Wrapf(err, "Failed to run goose up: %v", err)
	}

	logger.Logger.Info("Goose migrations down-to and up completed successfully")
	return nil
}
