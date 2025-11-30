package routes

import (
	"github.com/vellalasantosh/wound_iq_api_claude/internal/handlers"
	"github.com/vellalasantosh/wound_iq_api_claude/internal/middleware"

	"github.com/gin-gonic/gin"
)

// SetupAuthRoutes sets up authentication routes
func SetupAuthRoutes(router *gin.RouterGroup, authHandler *handlers.AuthHandler) {
	auth := router.Group("/auth")
	{
		// Public routes (no authentication required)
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.RefreshToken)

		// Protected routes (authentication required)
		protected := auth.Group("")
		protected.Use(middleware.AuthMiddleware())
		{
			protected.POST("/logout", authHandler.Logout)
			protected.GET("/profile", authHandler.GetProfile)
			protected.POST("/change-password", authHandler.ChangePassword)
		}
	}
}
