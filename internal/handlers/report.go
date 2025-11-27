package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/vellalasantosh/wound_iq_api_claude/internal/db"
	"github.com/vellalasantosh/wound_iq_api_claude/internal/models"

	"github.com/gin-gonic/gin"
)

// ReportHandler handles report-related requests
type ReportHandler struct {
	db *db.DB
}

// NewReportHandler creates a new report handler
func NewReportHandler(database *db.DB) *ReportHandler {
	return &ReportHandler{db: database}
}

// GetPatientWoundHistory retrieves wound history for a patient using get_patient_wound_history function
func (h *ReportHandler) GetPatientWoundHistory(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid patient ID",
			Message: "Patient ID must be a valid integer",
		})
		return
	}

	// Verify patient exists
	var exists bool
	err = h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM patient WHERE patient_id = $1)", id).Scan(&exists)
	if err != nil || !exists {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "Patient not found",
			Message: fmt.Sprintf("Patient with ID %d does not exist", id),
		})
		return
	}

	// Call get_patient_wound_history function
	rows, err := h.db.Query("SELECT * FROM get_patient_wound_history($1)", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to retrieve wound history",
			Message: err.Error(),
		})
		return
	}
	defer rows.Close()

	var history []models.WoundHistory
	for rows.Next() {
		var h models.WoundHistory
		if err := rows.Scan(&h.AssessmentID, &h.AssessmentDate, &h.Location, &h.Stage, &h.HealingStatus); err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "Failed to scan wound history",
				Message: err.Error(),
			})
			return
		}
		history = append(history, h)
	}

	c.JSON(http.StatusOK, gin.H{
		"patient_id": id,
		"history":    history,
	})
}

// GetFullAssessment retrieves full assessment details using get_assessment_full function
func (h *ReportHandler) GetFullAssessment(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid assessment ID",
			Message: "Assessment ID must be a valid integer",
		})
		return
	}

	// Call get_assessment_full function
	var result models.FullAssessmentResponse
	err = h.db.QueryRow("SELECT * FROM get_assessment_full($1)", id).Scan(
		&result.AssessmentID,
		&result.AssessmentDate,
		&result.PatientID,
		&result.PatientName,
		&result.ClinicianID,
		&result.ClinicianName,
		&result.Location,
		&result.Etiology,
		&result.Stage,
		&result.HealingStatus,
		&result.PainScore,
		&result.GranulationPercent,
		&result.Length,
		&result.Width,
	)

	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "Assessment not found",
			Message: fmt.Sprintf("Assessment with ID %d does not exist or could not be retrieved", id),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}
