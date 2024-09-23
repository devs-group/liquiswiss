package api

import (
	"liquiswiss/internal/api/handlers"
	"liquiswiss/internal/middleware"
	"liquiswiss/internal/service"

	"github.com/gin-gonic/gin"
)

type API struct {
	Router    *gin.Engine
	DBService service.IDatabaseService
}

func NewAPI(dbService service.IDatabaseService) *API {
	api := &API{
		Router:    gin.Default(),
		DBService: dbService,
	}
	api.setupRouter()
	return api
}

func (api *API) setupRouter() {
	group := api.Router.Group("/api")
	{
		public := group.Group("/auth")
		{
			public.POST("/register", func(ctx *gin.Context) {
				handlers.Register(api.DBService, ctx)
			})
			public.POST("/login", func(ctx *gin.Context) {
				handlers.Login(api.DBService, ctx)
			})
			public.GET("/logout", func(ctx *gin.Context) {
				handlers.Logout(api.DBService, ctx)
			})
		}

		protected := group.Group("/")
		protected.Use(middleware.AuthMiddleware)
		{
			protected.GET("/profile", func(ctx *gin.Context) {
				handlers.GetProfile(api.DBService, ctx)
			})
			protected.GET("/access-token", func(ctx *gin.Context) {
				handlers.GetAccessToken(api.DBService, ctx)
			})

			protected.GET("/organisations", func(ctx *gin.Context) {
				handlers.ListOrganisations(api.DBService, ctx)
			})
			protected.GET("/organisations/:id", func(ctx *gin.Context) {
				handlers.GetOrganisation(api.DBService, ctx)
			})
			protected.POST("/organisations", func(ctx *gin.Context) {
				handlers.CreateOrganisation(api.DBService, ctx)
			})
			protected.PATCH("/organisations/:id", func(ctx *gin.Context) {
				handlers.UpdateOrganisation(api.DBService, ctx)
			})

			protected.GET("/transactions", func(ctx *gin.Context) {
				handlers.ListTransactions(api.DBService, ctx)
			})
			protected.GET("/transactions/:id", func(ctx *gin.Context) {
				handlers.GetTransaction(api.DBService, ctx)
			})
			protected.POST("/transactions", func(ctx *gin.Context) {
				handlers.CreateTransaction(api.DBService, ctx)
			})
			protected.PATCH("/transactions/:id", func(ctx *gin.Context) {
				handlers.UpdateTransaction(api.DBService, ctx)
			})

			protected.GET("/employees", func(ctx *gin.Context) {
				handlers.ListEmployees(api.DBService, ctx)
			})
			protected.GET("/employees/pagination", func(ctx *gin.Context) {
				handlers.GetEmployeesPagination(api.DBService, ctx)
			})
			protected.GET("/employees/:id", func(ctx *gin.Context) {
				handlers.GetEmployee(api.DBService, ctx)
			})
			protected.POST("/employees", func(ctx *gin.Context) {
				handlers.CreateEmployee(api.DBService, ctx)
			})
			protected.PATCH("/employees/:id", func(ctx *gin.Context) {
				handlers.UpdateEmployee(api.DBService, ctx)
			})
			protected.DELETE("/employees/:id", func(ctx *gin.Context) {
				handlers.DeleteEmployee(api.DBService, ctx)
			})

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

			protected.GET("/currencies", func(ctx *gin.Context) {
				handlers.ListCurrencies(api.DBService, ctx)
			})
			protected.GET("/currencies/:id", func(ctx *gin.Context) {
				handlers.GetCurrency(api.DBService, ctx)
			})
			protected.POST("/currencies", func(ctx *gin.Context) {
				handlers.CreateCurrency(api.DBService, ctx)
			})
			protected.PATCH("/currencies/:id", func(ctx *gin.Context) {
				handlers.UpdateCurrency(api.DBService, ctx)
			})
		}
	}
}