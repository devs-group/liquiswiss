package api

import (
	"github.com/gin-gonic/gin"
	"liquiswiss/internal/adapter/db_adapter"
	"liquiswiss/internal/adapter/sendgrid_adapter"
	"liquiswiss/internal/api/handlers"
	"liquiswiss/internal/middleware"
	"liquiswiss/internal/service/api_service"
)

type API struct {
	Router          *gin.Engine
	DBService       db_adapter.IDatabaseAdapter
	APIService      api_service.IAPIService
	SendgridService sendgrid_adapter.ISendgridAdapter
}

func NewAPI(
	dbService db_adapter.IDatabaseAdapter,
	apiService api_service.IAPIService,
	sendgridService sendgrid_adapter.ISendgridAdapter,
) *API {
	gin.DebugPrintRouteFunc = func(httpMethod, absolutePath, handlerName string, nuHandlers int) {
		// Suppress listing all available routes for less log spamming
	}
	api := &API{
		Router:          gin.Default(),
		DBService:       dbService,
		APIService:      apiService,
		SendgridService: sendgridService,
	}
	api.setupRouter()
	return api
}

func (api *API) setupRouter() {
	group := api.Router.Group("/api")
	{
		public := group.Group("/auth")
		{
			public.POST("/login", func(ctx *gin.Context) {
				handlers.Login(api.APIService, ctx)
			})
			public.GET("/logout", func(ctx *gin.Context) {
				handlers.Logout(api.APIService, ctx)
			})
			public.POST("/forgot-password", func(ctx *gin.Context) {
				handlers.ForgotPassword(api.APIService, ctx)
			})
			public.POST("/reset-password", func(ctx *gin.Context) {
				handlers.ResetPassword(api.APIService, ctx)
			})
			public.POST("/reset-password-check-code", func(ctx *gin.Context) {
				handlers.CheckResetPasswordCode(api.APIService, ctx)
			})

			// Registration
			public.POST("/registration/create", func(ctx *gin.Context) {
				handlers.CreateRegistration(api.APIService, ctx)
			})
			public.POST("/registration/check-code", func(ctx *gin.Context) {
				handlers.CheckRegistrationCode(api.APIService, ctx)
			})
			public.POST("/registration/finish", func(ctx *gin.Context) {
				handlers.FinishRegistration(api.APIService, ctx)
			})
		}

		protected := group.Group("/")
		protected.Use(middleware.AuthMiddleware)
		{
			// Profile & Auth
			protected.GET("/profile", func(ctx *gin.Context) {
				handlers.GetProfile(api.APIService, ctx)
			})
			protected.PATCH("/profile", func(ctx *gin.Context) {
				handlers.UpdateProfile(api.APIService, ctx)
			})
			protected.POST("/profile/password", func(ctx *gin.Context) {
				handlers.UpdatePassword(api.APIService, ctx)
			})
			protected.PATCH("/profile/organisation", func(ctx *gin.Context) {
				handlers.SetUserCurrentOrganisation(api.APIService, ctx)
			})
			protected.GET("/profile/organisation", func(ctx *gin.Context) {
				handlers.GetUserCurrentOrganisation(api.APIService, ctx)
			})
			protected.GET("/access-token", func(ctx *gin.Context) {
				handlers.GetAccessToken(ctx)
			})

			// Organisations
			protected.GET("/organisations", func(ctx *gin.Context) {
				handlers.ListOrganisations(api.APIService, ctx)
			})
			protected.GET("/organisations/:organisationID", func(ctx *gin.Context) {
				handlers.GetOrganisation(api.APIService, ctx)
			})
			protected.POST("/organisations", func(ctx *gin.Context) {
				handlers.CreateOrganisation(api.APIService, ctx)
			})
			protected.PATCH("/organisations/:organisationID", func(ctx *gin.Context) {
				handlers.UpdateOrganisation(api.APIService, ctx)
			})
			// TODO: Find a way to delete organisations by offering reassigning or transferring data

			// Transactions
			protected.GET("/transactions", func(ctx *gin.Context) {
				handlers.ListTransactions(api.APIService, ctx)
			})
			protected.GET("/transactions/:transactionID", func(ctx *gin.Context) {
				handlers.GetTransaction(api.APIService, ctx)
			})
			protected.POST("/transactions", func(ctx *gin.Context) {
				handlers.CreateTransaction(api.APIService, ctx)
			})
			protected.PATCH("/transactions/:transactionID", func(ctx *gin.Context) {
				handlers.UpdateTransaction(api.APIService, ctx)
			})
			protected.DELETE("/transactions/:transactionID", func(ctx *gin.Context) {
				handlers.DeleteTransaction(api.APIService, ctx)
			})

			// Employees
			protected.GET("/employees", func(ctx *gin.Context) {
				handlers.ListEmployees(api.APIService, ctx)
			})
			protected.GET("/employees/:employeeID", func(ctx *gin.Context) {
				handlers.GetEmployee(api.APIService, ctx)
			})
			protected.POST("/employees", func(ctx *gin.Context) {
				handlers.CreateEmployee(api.APIService, ctx)
			})
			protected.PATCH("/employees/:employeeID", func(ctx *gin.Context) {
				handlers.UpdateEmployee(api.APIService, ctx)
			})
			protected.DELETE("/employees/:employeeID", func(ctx *gin.Context) {
				handlers.DeleteEmployee(api.APIService, ctx)
			})
			protected.GET("/employees/pagination", func(ctx *gin.Context) {
				handlers.GetEmployeesPagination(api.APIService, ctx)
			})

			// Employee Salaries
			protected.GET("/employees/:employeeID/salary", func(ctx *gin.Context) {
				handlers.ListSalaries(api.APIService, ctx)
			})
			protected.GET("/employees/salary/:salaryID", func(ctx *gin.Context) {
				handlers.GetSalary(api.APIService, ctx)
			})
			protected.POST("/employees/:employeeID/salary", func(ctx *gin.Context) {
				handlers.CreateSalary(api.APIService, ctx)
			})
			protected.PATCH("/employees/salary/:salaryID", func(ctx *gin.Context) {
				handlers.UpdateSalary(api.APIService, ctx)
			})
			protected.DELETE("/employees/salary/:salaryID", func(ctx *gin.Context) {
				handlers.DeleteSalary(api.APIService, ctx)
			})

			// Employee Salary Costs
			protected.GET("/employees/salary/:salaryID/costs", func(ctx *gin.Context) {
				handlers.ListSalaryCosts(api.APIService, ctx)
			})
			protected.GET("/employees/salary/costs/:salaryCostID", func(ctx *gin.Context) {
				handlers.GetSalaryCost(api.APIService, ctx)
			})
			protected.POST("/employees/salary/:salaryID/costs", func(ctx *gin.Context) {
				handlers.CreateSalaryCost(api.APIService, ctx)
			})
			protected.POST("/employees/salary/:salaryID/costs/copy", func(ctx *gin.Context) {
				handlers.CopySalaryCosts(api.APIService, ctx)
			})
			protected.PATCH("/employees/salary/costs/:salaryCostID", func(ctx *gin.Context) {
				handlers.UpdateSalaryCost(api.APIService, ctx)
			})
			protected.DELETE("/employees/salary/costs/:salaryCostID", func(ctx *gin.Context) {
				handlers.DeleteSalaryCost(api.APIService, ctx)
			})

			// Employee Salary Cost Labels
			protected.GET("/employees/salary/costs/labels", func(ctx *gin.Context) {
				handlers.ListSalaryCostLabels(api.APIService, ctx)
			})
			protected.GET("/employees/salary/costs/labels/:salaryCostLabelID", func(ctx *gin.Context) {
				handlers.GetSalaryCostLabel(api.APIService, ctx)
			})
			protected.POST("/employees/salary/costs/labels", func(ctx *gin.Context) {
				handlers.CreateSalaryCostLabel(api.APIService, ctx)
			})
			protected.PATCH("/employees/salary/costs/labels/:salaryCostLabelID", func(ctx *gin.Context) {
				handlers.UpdateSalaryCostLabel(api.APIService, ctx)
			})
			protected.DELETE("/employees/salary/costs/labels/:salaryCostLabelID", func(ctx *gin.Context) {
				handlers.DeleteSalaryCostLabel(api.APIService, ctx)
			})

			// Forecasts
			protected.GET("/forecasts", func(ctx *gin.Context) {
				handlers.ListForecasts(api.APIService, ctx)
			})
			protected.GET("/forecasts/details", func(ctx *gin.Context) {
				handlers.ListForecastDetails(api.APIService, ctx)
			})
			protected.GET("/forecasts/calculate", func(ctx *gin.Context) {
				handlers.CalculateForecasts(api.APIService, ctx)
			})
			protected.GET("/forecasts/exclude", func(ctx *gin.Context) {
				handlers.ListForecastExclusions(api.APIService, ctx)
			})
			protected.POST("/forecasts/exclude", func(ctx *gin.Context) {
				handlers.CreateForecastExclusion(api.APIService, ctx)
			})
			protected.PUT("/forecasts/exclude", func(ctx *gin.Context) {
				handlers.UpdateForecastExclusions(api.APIService, ctx)
			})
			protected.DELETE("/forecasts/exclude", func(ctx *gin.Context) {
				handlers.DeleteForecastExclusion(api.APIService, ctx)
			})

			// Bank Accounts
			protected.GET("/bank-accounts", func(ctx *gin.Context) {
				handlers.ListBankAccounts(api.APIService, ctx)
			})
			protected.GET("/bank-accounts/:bankAccountID", func(ctx *gin.Context) {
				handlers.GetBankAccount(api.APIService, ctx)
			})
			protected.POST("/bank-accounts", func(ctx *gin.Context) {
				handlers.CreateBankAccount(api.APIService, ctx)
			})
			protected.PATCH("/bank-accounts/:bankAccountID", func(ctx *gin.Context) {
				handlers.UpdateBankAccount(api.APIService, ctx)
			})
			protected.DELETE("/bank-accounts/:bankAccountID", func(ctx *gin.Context) {
				handlers.DeleteBankAccount(api.APIService, ctx)
			})

			// Vats
			protected.GET("/vats", func(ctx *gin.Context) {
				handlers.ListVats(api.APIService, ctx)
			})
			protected.GET("/vats/:vatID", func(ctx *gin.Context) {
				handlers.GetVat(api.APIService, ctx)
			})
			protected.POST("/vats", func(ctx *gin.Context) {
				handlers.CreateVat(api.APIService, ctx)
			})
			protected.PATCH("/vats/:vatID", func(ctx *gin.Context) {
				handlers.UpdateVat(api.APIService, ctx)
			})
			protected.DELETE("/vats/:vatID", func(ctx *gin.Context) {
				handlers.DeleteVat(api.APIService, ctx)
			})

			// VAT Settings
			protected.GET("/vat-settings", func(ctx *gin.Context) {
				handlers.GetVatSetting(api.APIService, ctx)
			})
			protected.POST("/vat-settings", func(ctx *gin.Context) {
				handlers.CreateVatSetting(api.APIService, ctx)
			})
			protected.PATCH("/vat-settings", func(ctx *gin.Context) {
				handlers.UpdateVatSetting(api.APIService, ctx)
			})
			protected.DELETE("/vat-settings", func(ctx *gin.Context) {
				handlers.DeleteVatSetting(api.APIService, ctx)
			})

			// Categories
			protected.GET("/categories", func(ctx *gin.Context) {
				handlers.ListCategories(api.APIService, ctx)
			})
			protected.GET("/categories/:id", func(ctx *gin.Context) {
				handlers.GetCategory(api.APIService, ctx)
			})
			protected.POST("/categories", func(ctx *gin.Context) {
				handlers.CreateCategory(api.APIService, ctx)
			})
			protected.PATCH("/categories/:id", func(ctx *gin.Context) {
				handlers.UpdateCategory(api.APIService, ctx)
			})

			// Currencies
			protected.GET("/currencies", func(ctx *gin.Context) {
				handlers.ListCurrencies(api.APIService, ctx)
			})
			protected.GET("/currencies/:currencyID", func(ctx *gin.Context) {
				handlers.GetCurrency(api.APIService, ctx)
			})
			protected.POST("/currencies", func(ctx *gin.Context) {
				handlers.CreateCurrency(api.APIService, ctx)
			})
			protected.PATCH("/currencies/:currencyID", func(ctx *gin.Context) {
				handlers.UpdateCurrency(api.APIService, ctx)
			})

			// Fiat Rates
			protected.GET("/fiat-rates/:base", func(ctx *gin.Context) {
				handlers.ListFiatRates(api.APIService, ctx)
			})
			protected.GET("/fiat-rates/:base/:target", func(ctx *gin.Context) {
				handlers.GetFiatRate(api.APIService, ctx)
			})
		}
	}
}
