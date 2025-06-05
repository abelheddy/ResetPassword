package server

import (
	"database/sql"
	"net/http"
	"time"
	"password-recovery/internal/config"
	"password-recovery/internal/handlers"
	"password-recovery/internal/handlers/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func Start(cfg config.AppConfig, db *sql.DB) error {
	router := gin.Default()
	
	// Configurar CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Middleware de base de datos
	router.Use(middleware.DatabaseMiddleware(db))

	// Configurar rutas
	setupRoutes(router)

	// Iniciar servidor
	return router.Run(":" + cfg.ServerPort)
}

func setupRoutes(router *gin.Engine) {
	// Rutas públicas
	public := router.Group("/")
	{
		public.POST("/send-code", handlers.SendCode)
		public.POST("/verify-code", handlers.VerifyCode)
		public.POST("/reset-password", handlers.ResetPassword)
	}

	// Rutas administrativas
	admin := router.Group("/admin")
	admin.Use(middleware.AuthMiddleware())
	{
		smtp := admin.Group("/smtp-config")
		{
			smtp.GET("", handlers.GetSMTPConfig)
			smtp.POST("", handlers.CreateSMTPConfig)
			smtp.PUT("", handlers.UpdateSMTPConfig)
			smtp.DELETE("", handlers.DeleteSMTPConfig)
			smtp.POST("/test", handlers.TestSMTPConnection)
		}
	}
}