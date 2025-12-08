package handlers_test

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/pressly/goose/v3"

	"liquiswiss/internal/adapter/db_adapter"
	"liquiswiss/internal/db"
	"liquiswiss/internal/service/api_service"
	"liquiswiss/pkg/logger"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"
)

func TestMain(m *testing.M) {
	utils.InitValidator()
	gin.SetMode(gin.TestMode)

	logger.NewZapLogger(false)

	code := m.Run()
	os.Exit(code)
}

// SetupTestEnvironment setups the base and `simulatedTime` can be optionally set to define the database date and time
func SetupTestEnvironment(t *testing.T) *sql.DB {
	testingEnvironment := os.Getenv("TESTING_ENVIRONMENT")
	dotEnvPath := "../../../.env.local.testing"
	if testingEnvironment == "github" {
		dotEnvPath = "../../../.env.github.testing"
	}
	// Load environment variables
	t.Logf("Loading environment variables from %s", dotEnvPath)
	err := godotenv.Load(dotEnvPath)
	if err != nil {
		t.Fatalf("Error loading .env file: %v", err)
	}

	// Connect to the test database
	conn, err := db.Connect()
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// Perform migrations
	migrateDatabase(t, conn)

	return conn
}

// CreateUserWithOrganisation is a helper method to quickly create a user with an organisation attached
func CreateUserWithOrganisation(apiService api_service.IAPIService, dbService db_adapter.IDatabaseAdapter, email, password, organisationName string) (*models.User, *models.Organisation, error) {
	userID, err := dbService.CreateUser(email, password)
	if err != nil {
		return nil, nil, err
	}

	organisation, err := apiService.CreateOrganisation(models.CreateOrganisation{
		Name: organisationName,
	}, userID, true)
	if err != nil {
		return nil, nil, err
	}

	userName := "John Doe"
	user, err := apiService.UpdateProfile(models.UpdateUser{
		Name: &userName,
	}, userID)
	if err != nil {
		return nil, nil, err
	}

	return user, organisation, nil
}

func CreateEmployee(apiService api_service.IAPIService, userID int64, name string) (*models.Employee, error) {
	employee, err := apiService.CreateEmployee(models.CreateEmployee{
		Name: name,
	}, userID)
	if err != nil {
		return nil, err
	}

	return employee, nil
}

func CreateCurrency(apiService api_service.IAPIService, code, description, localeCode string) (*models.Currency, error) {
	currency, err := apiService.CreateCurrency(models.CreateCurrency{
		Code:        code,
		Description: description,
		LocaleCode:  localeCode,
	})
	if err != nil {
		return nil, err
	}

	return currency, nil
}

func CreateSalaryCostLabel(apiService api_service.IAPIService, userID int64, name string) (*models.SalaryCostLabel, error) {
	salaryCostLabel, err := apiService.CreateSalaryCostLabel(models.CreateSalaryCostLabel{
		Name: name,
	}, userID)
	if err != nil {
		return nil, err
	}

	return salaryCostLabel, nil
}

func SetDatabaseTime(conn *sql.DB, simulatedTime string) error {
	query := fmt.Sprintf("SET TIMESTAMP = UNIX_TIMESTAMP('%s');", simulatedTime)
	_, err := conn.Exec(query)
	if err != nil {
		return err
	}
	return nil
}

func migrateDatabase(t *testing.T, conn *sql.DB) {
	migrationDirs := []string{
		"../../../internal/db/migrations/static",
		"../../../internal/db/migrations/dynamic",
	}

	// Configure Goose
	goose.SetBaseFS(nil)
	goose.SetLogger(goose.NopLogger())
	if err := goose.SetDialect("mysql"); err != nil {
		t.Fatalf("Failed to set Goose dialect: %v", err)
	}

	for _, dir := range migrationDirs {
		// Apply migrations
		if err := goose.DownTo(conn, dir, 0, goose.WithNoVersioning()); err != nil {
			if !isMissingTableError(err) {
				t.Fatalf("Failed to roll back migrations: %v", err)
			}
		}
		if err := goose.Up(conn, dir, goose.WithNoVersioning()); err != nil {
			t.Fatalf("Failed to apply migrations: %v", err)
		}
	}
}

func isMissingTableError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "doesn't exist") || strings.Contains(msg, "Unknown table")
}
