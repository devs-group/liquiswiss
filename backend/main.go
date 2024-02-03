package main

import (
	"liquiswiss/internal/api"
	"liquiswiss/internal/db"
	"liquiswiss/internal/middleware"
	"liquiswiss/internal/service"
	"liquiswiss/pkg/logger"
	"liquiswiss/pkg/utils"
	"net/http"
	"os"
)

func main() {
	log := logger.NewZapLogger()

	// Global Validator
	utils.InitValidator()

	db, err := db.Connect()
	if err != nil {
		log.Error("Failed to connect to database:", err)
		os.Exit(1)
	}
	defer db.Close()

	dbService := service.NewDatabaseService(db, log)
	middleware.InjectUserService(dbService)
	api := api.NewAPI(dbService)

	err = http.ListenAndServe(":8080", api.Router)
	if err != nil {
		log.Error("Failed to start api:", err)
		os.Exit(1)
	}
}
