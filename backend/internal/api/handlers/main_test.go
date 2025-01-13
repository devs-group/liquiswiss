package handlers_test

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/pressly/goose/v3"
	"liquiswiss/internal/db"
	"liquiswiss/internal/service/db_service"
	"liquiswiss/pkg/logger"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"
	"os"
	"testing"
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
func CreateUserWithOrganisation(dbService db_service.IDatabaseService, email, password, organisationName string) (*models.User, *models.Organisation, error) {
	userID, err := dbService.CreateUser(email, password)
	if err != nil {
		return nil, nil, err
	}

	organisationID, err := dbService.CreateOrganisation(organisationName)
	if err != nil {
		return nil, nil, err
	}

	err = dbService.AssignUserToOrganisation(userID, organisationID, "owner", true)
	if err != nil {
		return nil, nil, err
	}

	err = dbService.SetUserCurrentOrganisation(userID, organisationID)
	if err != nil {
		return nil, nil, err
	}

	userName := "John Doe"
	err = dbService.UpdateProfile(models.UpdateUser{
		Name: &userName,
	}, userID)
	if err != nil {
		return nil, nil, err
	}

	user, err := dbService.GetProfile(userID)
	if err != nil {
		return nil, nil, err
	}

	organisation, err := dbService.GetOrganisation(userID, organisationID)

	return user, organisation, nil
}

func CreateEmployee(dbService db_service.IDatabaseService, userID int64, name string) (*models.Employee, error) {
	employeeID, err := dbService.CreateEmployee(models.CreateEmployee{
		Name: name,
	}, userID)
	if err != nil {
		return nil, err
	}

	employee, err := dbService.GetEmployee(userID, employeeID)
	if err != nil {
		return nil, err
	}

	return employee, nil
}

func CreateCurrency(dbService db_service.IDatabaseService, code, description, localeCode string) (*models.Currency, error) {
	currencyID, err := dbService.CreateCurrency(models.CreateCurrency{
		Code:        code,
		Description: description,
		LocaleCode:  localeCode,
	})
	if err != nil {
		return nil, err
	}

	currency, err := dbService.GetCurrency(currencyID)
	if err != nil {
		return nil, err
	}

	return currency, nil
}

func CreateEmployeeHistoryCostLabel(dbService db_service.IDatabaseService, userID int64, name string) (*models.EmployeeHistoryCostLabel, error) {
	historyCostLabelID, err := dbService.CreateEmployeeHistoryCostLabel(models.CreateEmployeeHistoryCostLabel{
		Name: name,
	}, userID)
	if err != nil {
		return nil, err
	}

	historyCostLabel, err := dbService.GetEmployeeHistoryCostLabel(userID, historyCostLabelID)
	if err != nil {
		return nil, err
	}

	return historyCostLabel, nil
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
			t.Fatalf("Failed to roll back migrations: %v", err)
		}
		if err := goose.Up(conn, dir, goose.WithNoVersioning()); err != nil {
			t.Fatalf("Failed to apply migrations: %v", err)
		}
	}
}
