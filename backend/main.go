package main

import (
	"embed"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/pressly/goose/v3"
	"github.com/robfig/cron/v3"
	"liquiswiss/internal/api"
	"liquiswiss/internal/db"
	"liquiswiss/internal/middleware"
	"liquiswiss/internal/service"
	"liquiswiss/pkg/logger"
	"liquiswiss/pkg/utils"
	"net/http"
	"os"
)

//go:embed internal/db/migrations/*.sql
var embedMigrations embed.FS

func main() {
	// Init global logger
	logger.NewZapLogger()

	// Environment for DEV
	if !utils.IsProduction() {
		err := godotenv.Load()
		if err != nil {
			panic(fmt.Sprintf("Error loading .env file: %v", err))
		}
	}

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

	dbService := service.NewDatabaseService(conn)
	fixerIOService := service.NewFixerIOService(dbService)
	middleware.InjectUserService(dbService)
	apiHandler := api.NewAPI(dbService)

	// Cronjob
	c := cron.New()
	_, err = c.AddFunc("@every 15m", fixerIOService.FetchFiatRates)
	if err != nil {
		logger.Logger.Errorf("Failed to set fixer.io cronjob: %v", err)
		return
	}
	c.Start()

	if utils.IsProduction() {
		go fixerIOService.FetchFiatRates()
	} else {
		logger.Logger.Debug("Skipping Fiat Rates because we are not on Production")
	}

	// TODO: REMOVE ME BEFORE RELEASE
	err = dbService.ApplyMocks()
	if err != nil {
		logger.Logger.Warnf("Failed to apply mocks: %v", err)
	}

	err = http.ListenAndServe(":8080", apiHandler.Router)
	if err != nil {
		c.Stop()
		logger.Logger.Error("Failed to start api:", err)
		os.Exit(1)
	}
}
