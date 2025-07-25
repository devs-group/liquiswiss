package api

import (
	"liquiswiss/internal/api/handlers"
	"liquiswiss/internal/middleware"
	"liquiswiss/internal/service/db_service"
	"liquiswiss/internal/service/forecast_service"
	"liquiswiss/internal/service/sendgrid_service"
	"liquiswiss/internal/service/user_service"

	"github.com/gin-gonic/gin"
)

type API struct {
	Router          *gin.Engine
	DBService       db_service.IDatabaseService
	SendgridService sendgrid_service.ISendgridService
	ForecastService forecast_service.IForecastService
	UserService     user_service.IUserService
}

func NewAPI(
	dbService db_service.IDatabaseService,
	sendgridService sendgrid_service.ISendgridService,
	forecastService forecast_service.IForecastService,
	userService user_service.IUserService,
) *API {
	gin.DebugPrintRouteFunc = func(httpMethod, absolutePath, handlerName string, nuHandlers int) {
		// Suppress listing all available routes for less log spamming
	}
	api := &API{
		Router:          gin.Default(),
		DBService:       dbService,
		SendgridService: sendgridService,
		ForecastService: forecastService,
		UserService:     userService,
	}
	api.setupRouter()
	return api
}

func (api *API) setupRouter() {
	group := api.Router.Group("/api")
	{
		public := group.Group("/auth")
		{
			public.POST("/registration/create", func(ctx *gin.Context) {
				handlers.CreateRegistration(api.DBService, api.SendgridService, ctx)
			})
			public.POST("/registration/check-code", func(ctx *gin.Context) {
				handlers.CheckRegistrationCode(api.DBService, ctx)
			})
			public.POST("/registration/finish", func(ctx *gin.Context) {
				handlers.FinishRegistration(api.DBService, ctx)
			})
			public.POST("/login", func(ctx *gin.Context) {
				handlers.Login(api.DBService, ctx)
			})
			public.POST("/forgot-password", func(ctx *gin.Context) {
				handlers.ForgotPassword(api.DBService, api.SendgridService, ctx)
			})
			public.POST("/reset-password", func(ctx *gin.Context) {
				handlers.ResetPassword(api.DBService, ctx)
			})
			public.POST("/reset-password-check-code", func(ctx *gin.Context) {
				handlers.CheckResetPasswordCode(api.DBService, ctx)
			})
			public.GET("/logout", func(ctx *gin.Context) {
				handlers.Logout(api.DBService, ctx)
			})
		}

		protected := group.Group("/")
		protected.Use(middleware.AuthMiddleware)
		{
			// Profile & Auth
			protected.GET("/profile", func(ctx *gin.Context) {
				handlers.GetProfile(api.DBService, ctx)
			})
			protected.PATCH("/profile", func(ctx *gin.Context) {
				handlers.UpdateProfile(api.DBService, ctx)
			})
			protected.POST("/profile/password", func(ctx *gin.Context) {
				handlers.UpdatePassword(api.DBService, ctx)
			})
			protected.PATCH("/profile/organisation", func(ctx *gin.Context) {
				handlers.SetUserCurrentOrganisation(api.DBService, ctx)
			})
			protected.GET("/profile/organisation", func(ctx *gin.Context) {
				handlers.GetUserCurrentOrganisation(api.UserService, ctx)
			})
			protected.GET("/access-token", func(ctx *gin.Context) {
				handlers.GetAccessToken(api.DBService, ctx)
			})

			// Organisations
			protected.GET("/organisations", func(ctx *gin.Context) {
				handlers.ListOrganisations(api.DBService, ctx)
			})
			protected.GET("/organisations/:organisationID", func(ctx *gin.Context) {
				handlers.GetOrganisation(api.DBService, ctx)
			})
			protected.POST("/organisations", func(ctx *gin.Context) {
				handlers.CreateOrganisation(api.DBService, ctx)
			})
			protected.PATCH("/organisations/:organisationID", func(ctx *gin.Context) {
				handlers.UpdateOrganisation(api.DBService, ctx)
			})
			// TODO: Find a way to delete organisations by offering reassigning or transferring data

			// Transactions
			protected.GET("/transactions", func(ctx *gin.Context) {
				handlers.ListTransactions(api.DBService, ctx)
			})
			protected.GET("/transactions/:transactionID", func(ctx *gin.Context) {
				handlers.GetTransaction(api.DBService, ctx)
			})
			protected.POST("/transactions", func(ctx *gin.Context) {
				handlers.CreateTransaction(api.DBService, api.ForecastService, ctx)
			})
			protected.PATCH("/transactions/:transactionID", func(ctx *gin.Context) {
				handlers.UpdateTransaction(api.DBService, api.ForecastService, ctx)
			})
			protected.DELETE("/transactions/:transactionID", func(ctx *gin.Context) {
				handlers.DeleteTransaction(api.DBService, api.ForecastService, ctx)
			})

			// Employees
			protected.GET("/employees", func(ctx *gin.Context) {
				handlers.ListEmployees(api.DBService, ctx)
			})
			protected.GET("/employees/:employeeID", func(ctx *gin.Context) {
				handlers.GetEmployee(api.DBService, ctx)
			})
			protected.POST("/employees", func(ctx *gin.Context) {
				handlers.CreateEmployee(api.DBService, ctx)
			})
			protected.PATCH("/employees/:employeeID", func(ctx *gin.Context) {
				handlers.UpdateEmployee(api.DBService, ctx)
			})
			protected.DELETE("/employees/:employeeID", func(ctx *gin.Context) {
				handlers.DeleteEmployee(api.DBService, ctx)
			})
			protected.GET("/employees/pagination", func(ctx *gin.Context) {
				handlers.GetEmployeesPagination(api.DBService, ctx)
			})

			// Employee Salaries
			protected.GET("/employees/:employeeID/salary", func(ctx *gin.Context) {
				handlers.ListSalaries(api.DBService, ctx)
			})
			protected.GET("/employees/salary/:salaryID", func(ctx *gin.Context) {
				handlers.GetSalary(api.DBService, ctx)
			})
			protected.POST("/employees/:employeeID/salary", func(ctx *gin.Context) {
				handlers.CreateSalary(api.DBService, api.ForecastService, ctx)
			})
			protected.PATCH("/employees/salary/:salaryID", func(ctx *gin.Context) {
				handlers.UpdateSalary(api.DBService, api.ForecastService, ctx)
			})
			protected.DELETE("/employees/salary/:salaryID", func(ctx *gin.Context) {
				handlers.DeleteSalary(api.DBService, api.ForecastService, ctx)
			})

			// Employee Salary Costs
			protected.GET("/employees/salary/:salaryID/costs", func(ctx *gin.Context) {
				handlers.ListSalaryCosts(api.DBService, ctx)
			})
			protected.GET("/employees/salary/costs/:salaryCostID", func(ctx *gin.Context) {
				handlers.GetSalaryCost(api.DBService, ctx)
			})
			protected.POST("/employees/salary/:salaryID/costs", func(ctx *gin.Context) {
				handlers.CreateSalaryCost(api.DBService, api.ForecastService, ctx)
			})
			protected.POST("/employees/salary/:salaryID/costs/copy", func(ctx *gin.Context) {
				handlers.CopySalaryCosts(api.DBService, api.ForecastService, ctx)
			})
			protected.PATCH("/employees/salary/costs/:salaryCostID", func(ctx *gin.Context) {
				handlers.UpdateSalaryCost(api.DBService, api.ForecastService, ctx)
			})
			protected.DELETE("/employees/salary/costs/:salaryCostID", func(ctx *gin.Context) {
				handlers.DeleteSalaryCost(api.DBService, api.ForecastService, ctx)
			})

			// Employee Salary Cost Labels
			protected.GET("/employees/salary/costs/labels", func(ctx *gin.Context) {
				handlers.ListSalaryCostLabels(api.DBService, ctx)
			})
			protected.GET("/employees/salary/costs/labels/:salaryCostLabelID", func(ctx *gin.Context) {
				handlers.GetSalaryCostLabel(api.DBService, ctx)
			})
			protected.POST("/employees/salary/costs/labels", func(ctx *gin.Context) {
				handlers.CreateSalaryCostLabel(api.DBService, ctx)
			})
			protected.PATCH("/employees/salary/costs/labels/:salaryCostLabelID", func(ctx *gin.Context) {
				handlers.UpdateSalaryCostLabel(api.DBService, ctx)
			})
			protected.DELETE("/employees/salary/costs/labels/:salaryCostLabelID", func(ctx *gin.Context) {
				handlers.DeleteSalaryCostLabel(api.DBService, ctx)
			})

			// Forecasts
			protected.GET("/forecasts", func(ctx *gin.Context) {
				handlers.ListForecasts(api.DBService, ctx)
			})
			protected.GET("/forecasts/details", func(ctx *gin.Context) {
				handlers.ListForecastDetails(api.DBService, ctx)
			})
			protected.GET("/forecasts/calculate", func(ctx *gin.Context) {
				handlers.CalculateForecasts(api.ForecastService, ctx)
			})
			protected.GET("/forecasts/exclude", func(ctx *gin.Context) {
				handlers.ListForecastExclusions(api.DBService, ctx)
			})
			protected.POST("/forecasts/exclude", func(ctx *gin.Context) {
				handlers.CreateForecastExclusion(api.DBService, ctx)
			})
			protected.DELETE("/forecasts/exclude", func(ctx *gin.Context) {
				handlers.DeleteForecastExclusion(api.DBService, ctx)
			})

			// Bank Accounts
			protected.GET("/bank-accounts", func(ctx *gin.Context) {
				handlers.ListBankAccounts(api.DBService, ctx)
			})
			protected.GET("/bank-accounts/:bankAccountID", func(ctx *gin.Context) {
				handlers.GetBankAccount(api.DBService, ctx)
			})
			protected.POST("/bank-accounts", func(ctx *gin.Context) {
				handlers.CreateBankAccount(api.DBService, ctx)
			})
			protected.PATCH("/bank-accounts/:bankAccountID", func(ctx *gin.Context) {
				handlers.UpdateBankAccount(api.DBService, ctx)
			})
			protected.DELETE("/bank-accounts/:bankAccountID", func(ctx *gin.Context) {
				handlers.DeleteBankAccount(api.DBService, ctx)
			})

			// Vats
			protected.GET("/vats", func(ctx *gin.Context) {
				handlers.ListVats(api.DBService, ctx)
			})
			protected.GET("/vats/:vatID", func(ctx *gin.Context) {
				handlers.GetVat(api.DBService, ctx)
			})
			protected.POST("/vats", func(ctx *gin.Context) {
				handlers.CreateVat(api.DBService, ctx)
			})
			protected.PATCH("/vats/:vatID", func(ctx *gin.Context) {
				handlers.UpdateVat(api.DBService, ctx)
			})
			protected.DELETE("/vats/:vatID", func(ctx *gin.Context) {
				handlers.DeleteVat(api.DBService, ctx)
			})

			// Categories
			protected.GET("/categories", func(ctx *gin.Context) {
				handlers.ListCategories(api.DBService, ctx)
			})
			protected.GET("/categories/:id", func(ctx *gin.Context) {
				handlers.GetCategory(api.DBService, ctx)
			})
			protected.POST("/categories", func(ctx *gin.Context) {
				handlers.CreateCategory(api.DBService, ctx)
			})
			protected.PATCH("/categories/:id", func(ctx *gin.Context) {
				handlers.UpdateCategory(api.DBService, ctx)
			})

			// Currencies
			protected.GET("/currencies", func(ctx *gin.Context) {
				handlers.ListCurrencies(api.DBService, ctx)
			})
			protected.GET("/currencies/:currencyID", func(ctx *gin.Context) {
				handlers.GetCurrency(api.DBService, ctx)
			})
			protected.POST("/currencies", func(ctx *gin.Context) {
				handlers.CreateCurrency(api.DBService, ctx)
			})
			protected.PATCH("/currencies/:currencyID", func(ctx *gin.Context) {
				handlers.UpdateCurrency(api.DBService, ctx)
			})

			// Fiat Rates
			protected.GET("/fiat-rates/:base", func(ctx *gin.Context) {
				handlers.ListFiatRates(api.DBService, ctx)
			})
			protected.GET("/fiat-rates/:base/:target", func(ctx *gin.Context) {
				handlers.GetFiatRate(api.DBService, ctx)
			})
		}
	}
}
