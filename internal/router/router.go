package router

import (
	"net/http"
	"time"

	"github.com/vellalasantosh/wound_iq_api_claude/internal/db"
	"github.com/vellalasantosh/wound_iq_api_claude/internal/handlers"

	"github.com/gin-gonic/gin"
)

// SetupRouter configures all routes and middleware
func SetupRouter(database *db.DB) *gin.Engine {
	// Set Gin to release mode for production
	// gin.SetMode(gin.ReleaseMode) // Uncomment in production

	r := gin.New()

	// Middleware
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(corsMiddleware())

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	// Initialize handlers
	patientHandler := handlers.NewPatientHandler(database)
	clinicianHandler := handlers.NewClinicianHandler(database)
	assessmentHandler := handlers.NewAssessmentHandler(database)
	reportHandler := handlers.NewReportHandler(database)

	// API v1 routes
	v1 := r.Group("/v1")
	{
		// Patient routes
		patients := v1.Group("/patients")
		{
			patients.GET("", patientHandler.GetAllPatients)
			patients.GET("/:id", patientHandler.GetPatientByID)
			patients.POST("", patientHandler.CreatePatient)
			patients.PUT("/:id", patientHandler.UpdatePatient)
			patients.DELETE("/:id", patientHandler.DeletePatient)
			patients.GET("/:id/history", reportHandler.GetPatientWoundHistory)
		}

		// Clinician routes
		clinicians := v1.Group("/clinicians")
		{
			clinicians.GET("", clinicianHandler.GetAllClinicians)
			clinicians.GET("/:id", clinicianHandler.GetClinicianByID)
			clinicians.POST("", clinicianHandler.CreateClinician)
			clinicians.PUT("/:id", clinicianHandler.UpdateClinician)
			clinicians.DELETE("/:id", clinicianHandler.DeleteClinician)
		}

		// Assessment routes
		assessments := v1.Group("/assessments")
		{
			assessments.GET("", assessmentHandler.GetAllAssessments)
			assessments.GET("/:id", assessmentHandler.GetAssessmentByID)
			assessments.POST("", assessmentHandler.CreateAssessment)
			assessments.POST("/full", assessmentHandler.CreateFullAssessment)
			assessments.PUT("/:id", assessmentHandler.UpdateAssessment)
			assessments.DELETE("/:id", assessmentHandler.DeleteAssessment)
			assessments.GET("/:id/full", reportHandler.GetFullAssessment)
		}
	}

	// 404 handler
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Route not found",
			"message": "The requested endpoint does not exist",
		})
	})

	return r
}

// corsMiddleware adds CORS headers
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
