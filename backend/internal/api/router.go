package api

import (
	"github.com/gin-gonic/gin"
	"liquiswiss/internal/adapter/db_adapter"
	"liquiswiss/internal/adapter/email_adapter"
	"liquiswiss/internal/api/handlers"
	"liquiswiss/internal/mcp"
	"liquiswiss/internal/middleware"
	"liquiswiss/internal/oauth"
	"liquiswiss/internal/service/api_service"
)

type API struct {
	Router          *gin.Engine
	DBService       db_adapter.IDatabaseAdapter
	APIService      api_service.IAPIService
	EmailService email_adapter.IEmailAdapter
}

func NewAPI(
	dbService db_adapter.IDatabaseAdapter,
	apiService api_service.IAPIService,
	emailService email_adapter.IEmailAdapter,
) *API {
	gin.DebugPrintRouteFunc = func(httpMethod, absolutePath, handlerName string, nuHandlers int) {
		// Suppress listing all available routes for less log spamming
	}
	api := &API{
		Router:          gin.Default(),
		DBService:       dbService,
		APIService:      apiService,
		EmailService: emailService,
	}
	api.setupRouter()
	return api
}

func (api *API) setupRouter() {
	oauthHandler := oauth.NewHandler(api.DBService)

	// OAuth discovery endpoints must live at the root path per RFC 8414
	api.Router.GET("/.well-known/oauth-protected-resource", oauthHandler.ProtectedResourceMetadata)
	api.Router.GET("/.well-known/oauth-authorization-server", oauthHandler.AuthorizationServerMetadata)

	group := api.Router.Group("/api")
	{
		// OAuth 2.1 authorization server (public endpoints)
		oauthGroup := group.Group("/oauth")
		{
			oauthGroup.POST("/register", oauthHandler.Register)
			oauthGroup.GET("/authorize", oauthHandler.Authorize)
			oauthGroup.POST("/token", oauthHandler.Token)
			oauthGroup.POST("/revoke", oauthHandler.Revoke)
		}

		// MCP endpoint protected by OAuth bearer tokens
		mcpGroup := group.Group("/mcp")
		mcpGroup.Use(middleware.OAuthBearerMiddleware)
		mcpGroup.Any("", mcp.GinHandler(api.APIService, api.DBService))

		// Health check endpoint for monitoring and CI
		group.GET("/health", func(ctx *gin.Context) {
			ctx.JSON(200, gin.H{"status": "ok"})
		})

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

			// Invitations (public)
			public.GET("/invitation/check", func(ctx *gin.Context) {
				handlers.CheckInvitation(api.APIService, ctx)
			})
			public.POST("/invitation/accept", func(ctx *gin.Context) {
				handlers.AcceptInvitation(api.APIService, ctx)
			})
		}

		protected := group.Group("/")
		protected.Use(middleware.AuthMiddleware)
		// editorRoutes: mutations on organisation-scoped business data (editor+)
		editorRoutes := protected.Group("/")
		editorRoutes.Use(middleware.RequireMinRole(middleware.RoleEditor))
		// adminRoutes: organisation + member/invitation management (admin+)
		adminRoutes := protected.Group("/")
		adminRoutes.Use(middleware.RequireMinRole(middleware.RoleAdmin))
		{
			// OAuth consent + connection management (cookie session)
			protected.POST("/oauth/approve", oauthHandler.Approve)
			protected.GET("/oauth/connections", oauthHandler.ListConnections)
			protected.DELETE("/oauth/connections/:clientId", oauthHandler.RevokeConnection)

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
			adminRoutes.PATCH("/organisations/:organisationID", func(ctx *gin.Context) {
				handlers.UpdateOrganisation(api.APIService, ctx)
			})
			// TODO: Find a way to delete organisations by offering reassigning or transferring data

			// Organisation Members
			protected.GET("/organisations/:organisationID/members", func(ctx *gin.Context) {
				handlers.ListOrganisationMembers(api.APIService, ctx)
			})
			adminRoutes.PATCH("/organisations/:organisationID/members/:memberUserID", func(ctx *gin.Context) {
				handlers.UpdateOrganisationMember(api.APIService, ctx)
			})
			adminRoutes.DELETE("/organisations/:organisationID/members/:memberUserID", func(ctx *gin.Context) {
				handlers.RemoveOrganisationMember(api.APIService, ctx)
			})

			// Pending invitations for the current user (across any organisation)
			protected.GET("/me/invitations", func(ctx *gin.Context) {
				handlers.ListMyPendingInvitations(api.APIService, ctx)
			})
			protected.DELETE("/me/invitations/:invitationID", func(ctx *gin.Context) {
				handlers.DeclineMyInvitation(api.APIService, ctx)
			})

			// Organisation Invitations (admin+ only)
			adminRoutes.GET("/organisations/:organisationID/invitations", func(ctx *gin.Context) {
				handlers.ListOrganisationInvitations(api.APIService, ctx)
			})
			adminRoutes.POST("/organisations/:organisationID/invitations", func(ctx *gin.Context) {
				handlers.CreateOrganisationInvitation(api.APIService, ctx)
			})
			adminRoutes.DELETE("/organisations/:organisationID/invitations/:invitationID", func(ctx *gin.Context) {
				handlers.DeleteOrganisationInvitation(api.APIService, ctx)
			})
			adminRoutes.POST("/organisations/:organisationID/invitations/:invitationID/resend", func(ctx *gin.Context) {
				handlers.ResendOrganisationInvitation(api.APIService, ctx)
			})

			// Transactions
			protected.GET("/transactions", func(ctx *gin.Context) {
				handlers.ListTransactions(api.APIService, ctx)
			})
			protected.GET("/transactions/:transactionID", func(ctx *gin.Context) {
				handlers.GetTransaction(api.APIService, ctx)
			})
			editorRoutes.POST("/transactions", func(ctx *gin.Context) {
				handlers.CreateTransaction(api.APIService, ctx)
			})
			editorRoutes.PATCH("/transactions/:transactionID", func(ctx *gin.Context) {
				handlers.UpdateTransaction(api.APIService, ctx)
			})
			editorRoutes.DELETE("/transactions/:transactionID", func(ctx *gin.Context) {
				handlers.DeleteTransaction(api.APIService, ctx)
			})

			// Employees
			protected.GET("/employees", func(ctx *gin.Context) {
				handlers.ListEmployees(api.APIService, ctx)
			})
			protected.GET("/employees/:employeeID", func(ctx *gin.Context) {
				handlers.GetEmployee(api.APIService, ctx)
			})
			editorRoutes.POST("/employees", func(ctx *gin.Context) {
				handlers.CreateEmployee(api.APIService, ctx)
			})
			editorRoutes.PATCH("/employees/:employeeID", func(ctx *gin.Context) {
				handlers.UpdateEmployee(api.APIService, ctx)
			})
			editorRoutes.DELETE("/employees/:employeeID", func(ctx *gin.Context) {
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
			editorRoutes.POST("/employees/:employeeID/salary", func(ctx *gin.Context) {
				handlers.CreateSalary(api.APIService, ctx)
			})
			editorRoutes.PATCH("/employees/salary/:salaryID", func(ctx *gin.Context) {
				handlers.UpdateSalary(api.APIService, ctx)
			})
			editorRoutes.DELETE("/employees/salary/:salaryID", func(ctx *gin.Context) {
				handlers.DeleteSalary(api.APIService, ctx)
			})

			// Employee Salary Costs
			protected.GET("/employees/salary/:salaryID/costs", func(ctx *gin.Context) {
				handlers.ListSalaryCosts(api.APIService, ctx)
			})
			protected.GET("/employees/salary/costs/:salaryCostID", func(ctx *gin.Context) {
				handlers.GetSalaryCost(api.APIService, ctx)
			})
			editorRoutes.POST("/employees/salary/:salaryID/costs", func(ctx *gin.Context) {
				handlers.CreateSalaryCost(api.APIService, ctx)
			})
			editorRoutes.POST("/employees/salary/:salaryID/costs/copy", func(ctx *gin.Context) {
				handlers.CopySalaryCosts(api.APIService, ctx)
			})
			editorRoutes.PATCH("/employees/salary/costs/:salaryCostID", func(ctx *gin.Context) {
				handlers.UpdateSalaryCost(api.APIService, ctx)
			})
			editorRoutes.DELETE("/employees/salary/costs/:salaryCostID", func(ctx *gin.Context) {
				handlers.DeleteSalaryCost(api.APIService, ctx)
			})

			// Employee Salary Cost Labels
			protected.GET("/employees/salary/costs/labels", func(ctx *gin.Context) {
				handlers.ListSalaryCostLabels(api.APIService, ctx)
			})
			protected.GET("/employees/salary/costs/labels/:salaryCostLabelID", func(ctx *gin.Context) {
				handlers.GetSalaryCostLabel(api.APIService, ctx)
			})
			editorRoutes.POST("/employees/salary/costs/labels", func(ctx *gin.Context) {
				handlers.CreateSalaryCostLabel(api.APIService, ctx)
			})
			editorRoutes.PATCH("/employees/salary/costs/labels/:salaryCostLabelID", func(ctx *gin.Context) {
				handlers.UpdateSalaryCostLabel(api.APIService, ctx)
			})
			editorRoutes.DELETE("/employees/salary/costs/labels/:salaryCostLabelID", func(ctx *gin.Context) {
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
			editorRoutes.POST("/forecasts/exclude", func(ctx *gin.Context) {
				handlers.CreateForecastExclusion(api.APIService, ctx)
			})
			editorRoutes.PUT("/forecasts/exclude", func(ctx *gin.Context) {
				handlers.UpdateForecastExclusions(api.APIService, ctx)
			})
			editorRoutes.DELETE("/forecasts/exclude", func(ctx *gin.Context) {
				handlers.DeleteForecastExclusion(api.APIService, ctx)
			})

			// Bank Accounts
			protected.GET("/bank-accounts", func(ctx *gin.Context) {
				handlers.ListBankAccounts(api.APIService, ctx)
			})
			protected.GET("/bank-accounts/:bankAccountID", func(ctx *gin.Context) {
				handlers.GetBankAccount(api.APIService, ctx)
			})
			editorRoutes.POST("/bank-accounts", func(ctx *gin.Context) {
				handlers.CreateBankAccount(api.APIService, ctx)
			})
			editorRoutes.PATCH("/bank-accounts/:bankAccountID", func(ctx *gin.Context) {
				handlers.UpdateBankAccount(api.APIService, ctx)
			})
			editorRoutes.DELETE("/bank-accounts/:bankAccountID", func(ctx *gin.Context) {
				handlers.DeleteBankAccount(api.APIService, ctx)
			})

			// Vats
			protected.GET("/vats", func(ctx *gin.Context) {
				handlers.ListVats(api.APIService, ctx)
			})
			protected.GET("/vats/:vatID", func(ctx *gin.Context) {
				handlers.GetVat(api.APIService, ctx)
			})
			editorRoutes.POST("/vats", func(ctx *gin.Context) {
				handlers.CreateVat(api.APIService, ctx)
			})
			editorRoutes.PATCH("/vats/:vatID", func(ctx *gin.Context) {
				handlers.UpdateVat(api.APIService, ctx)
			})
			editorRoutes.DELETE("/vats/:vatID", func(ctx *gin.Context) {
				handlers.DeleteVat(api.APIService, ctx)
			})

			// VAT Settings
			protected.GET("/vat-settings", func(ctx *gin.Context) {
				handlers.GetVatSetting(api.APIService, ctx)
			})
			editorRoutes.POST("/vat-settings", func(ctx *gin.Context) {
				handlers.CreateVatSetting(api.APIService, ctx)
			})
			editorRoutes.PATCH("/vat-settings", func(ctx *gin.Context) {
				handlers.UpdateVatSetting(api.APIService, ctx)
			})
			editorRoutes.DELETE("/vat-settings", func(ctx *gin.Context) {
				handlers.DeleteVatSetting(api.APIService, ctx)
			})

			// User Settings (global)
			protected.GET("/user-settings", func(ctx *gin.Context) {
				handlers.GetUserSetting(api.APIService, ctx)
			})
			protected.PATCH("/user-settings", func(ctx *gin.Context) {
				handlers.UpdateUserSetting(api.APIService, ctx)
			})

			// User Organisation Settings (per-organisation)
			protected.GET("/user-organisation-settings", func(ctx *gin.Context) {
				handlers.GetUserOrganisationSetting(api.APIService, ctx)
			})
			protected.PATCH("/user-organisation-settings", func(ctx *gin.Context) {
				handlers.UpdateUserOrganisationSetting(api.APIService, ctx)
			})

			// Categories
			protected.GET("/categories", func(ctx *gin.Context) {
				handlers.ListCategories(api.APIService, ctx)
			})
			protected.GET("/categories/:id", func(ctx *gin.Context) {
				handlers.GetCategory(api.APIService, ctx)
			})
			editorRoutes.POST("/categories", func(ctx *gin.Context) {
				handlers.CreateCategory(api.APIService, ctx)
			})
			editorRoutes.PATCH("/categories/:id", func(ctx *gin.Context) {
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
