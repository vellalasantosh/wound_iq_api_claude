package router

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vellalasantosh/wound_iq_api_claude/internal/db"
	"github.com/vellalasantosh/wound_iq_api_claude/internal/handlers"
)

// SetupRouter configures all routes and middleware EXCEPT auth routes
func SetupRouter(database *db.DB) *gin.Engine {

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

	// ⭐ Unified API v1 routes: /api/v1
	v1 := r.Group("/api/v1")

	// ---------------------------
	// Patient routes
	// ---------------------------
	patients := v1.Group("/patients")
	{
		patients.GET("", patientHandler.GetAllPatients)
		patients.GET("/:id", patientHandler.GetPatientByID)
		patients.POST("", patientHandler.CreatePatient)
		patients.PUT("/:id", patientHandler.UpdatePatient)
		patients.DELETE("/:id", patientHandler.DeletePatient)
		patients.GET("/:id/history", reportHandler.GetPatientWoundHistory)
	}

	// ---------------------------
	// Clinician routes
	// ---------------------------
	clinicians := v1.Group("/clinicians")
	{
		clinicians.GET("", clinicianHandler.GetAllClinicians)
		clinicians.GET("/:id", clinicianHandler.GetClinicianByID)
		clinicians.POST("", clinicianHandler.CreateClinician)
		clinicians.PUT("/:id", clinicianHandler.UpdateClinician)
		clinicians.DELETE("/:id", clinicianHandler.DeleteClinician)
	}

	// ---------------------------
	// Assessment routes
	// ---------------------------
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

	// 404 NOT FOUND handler
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Route not found",
			"message": "The requested endpoint does not exist",
		})
	})

	return r
}

// CORS middleware
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers",
			"Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods",
			"POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
