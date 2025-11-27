package handlers

import (
	"database/sql"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/vellalasantosh/wound_iq_api_claude/internal/db"
	"github.com/vellalasantosh/wound_iq_api_claude/internal/models"

	"github.com/gin-gonic/gin"
)

// AssessmentHandler handles assessment-related requests
type AssessmentHandler struct {
	db *db.DB
}

// NewAssessmentHandler creates a new assessment handler
func NewAssessmentHandler(database *db.DB) *AssessmentHandler {
	return &AssessmentHandler{db: database}
}

// GetAllAssessments retrieves all assessments with filters and pagination
func (h *AssessmentHandler) GetAllAssessments(c *gin.Context) {
	var filter models.AssessmentFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid query parameters",
			Message: err.Error(),
		})
		return
	}

	// Build query with filters
	query := `
		SELECT a.assessment_id, a.date, p.patient_id, p.full_name, 
		       c.clinician_id, c.full_name, a.location
		FROM assessment a
		JOIN patient p ON p.patient_id = a.patient_id
		JOIN clinician c ON c.clinician_id = a.clinician_id
		WHERE 1=1
	`
	countQuery := "SELECT COUNT(*) FROM assessment a WHERE 1=1"
	args := []interface{}{}
	argPos := 1

	// Apply filters
	if filter.PatientID != nil {
		query += fmt.Sprintf(" AND a.patient_id = $%d", argPos)
		countQuery += fmt.Sprintf(" AND patient_id = $%d", argPos)
		args = append(args, *filter.PatientID)
		argPos++
	}
	if filter.ClinicianID != nil {
		query += fmt.Sprintf(" AND a.clinician_id = $%d", argPos)
		countQuery += fmt.Sprintf(" AND clinician_id = $%d", argPos)
		args = append(args, *filter.ClinicianID)
		argPos++
	}
	if filter.StartDate != "" {
		startDate, err := time.Parse(time.RFC3339, filter.StartDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "Invalid start_date format",
				Message: "Date must be in ISO-8601 format",
			})
			return
		}
		query += fmt.Sprintf(" AND a.date >= $%d", argPos)
		countQuery += fmt.Sprintf(" AND date >= $%d", argPos)
		args = append(args, startDate)
		argPos++
	}
	if filter.EndDate != "" {
		endDate, err := time.Parse(time.RFC3339, filter.EndDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "Invalid end_date format",
				Message: "Date must be in ISO-8601 format",
			})
			return
		}
		query += fmt.Sprintf(" AND a.date <= $%d", argPos)
		countQuery += fmt.Sprintf(" AND date <= $%d", argPos)
		args = append(args, endDate)
		argPos++
	}

	// Get total count
	var totalCount int
	err := h.db.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to count assessments",
			Message: err.Error(),
		})
		return
	}

	// Add ordering and pagination
	query += fmt.Sprintf(" ORDER BY a.date DESC LIMIT $%d OFFSET $%d", argPos, argPos+1)
	args = append(args, filter.GetLimit(), filter.GetOffset())

	rows, err := h.db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to query assessments",
			Message: err.Error(),
		})
		return
	}
	defer rows.Close()

	type AssessmentListItem struct {
		AssessmentID  int       `json:"assessment_id"`
		Date          time.Time `json:"date"`
		PatientID     int       `json:"patient_id"`
		PatientName   string    `json:"patient_name"`
		ClinicianID   int       `json:"clinician_id"`
		ClinicianName string    `json:"clinician_name"`
		Location      string    `json:"location"`
	}

	var assessments []AssessmentListItem
	for rows.Next() {
		var a AssessmentListItem
		if err := rows.Scan(&a.AssessmentID, &a.Date, &a.PatientID, &a.PatientName, &a.ClinicianID, &a.ClinicianName, &a.Location); err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "Failed to scan assessment",
				Message: err.Error(),
			})
			return
		}
		assessments = append(assessments, a)
	}

	totalPages := int(math.Ceil(float64(totalCount) / float64(filter.GetLimit())))

	c.JSON(http.StatusOK, models.PaginatedResponse{
		Data:       assessments,
		Page:       filter.Page,
		PageSize:   filter.GetLimit(),
		TotalCount: totalCount,
		TotalPages: totalPages,
	})
}

// GetAssessmentByID retrieves a single assessment by ID
func (h *AssessmentHandler) GetAssessmentByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid assessment ID",
			Message: "Assessment ID must be a valid integer",
		})
		return
	}

	var assessment models.Assessment
	err = h.db.QueryRow(`
		SELECT assessment_id, clinician_id, patient_id, date, location, etiology, 
		       depth_of_injury, stage, chronicity, healing_status, return_to_clinic
		FROM assessment
		WHERE assessment_id = $1
	`, id).Scan(&assessment.AssessmentID, &assessment.ClinicianID, &assessment.PatientID,
		&assessment.Date, &assessment.Location, &assessment.Etiology, &assessment.DepthOfInjury,
		&assessment.Stage, &assessment.Chronicity, &assessment.HealingStatus, &assessment.ReturnToClinic)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "Assessment not found",
			Message: fmt.Sprintf("Assessment with ID %d does not exist", id),
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to query assessment",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, assessment)
}

// CreateAssessment creates a new assessment (simple version)
func (h *AssessmentHandler) CreateAssessment(c *gin.Context) {
	var req models.CreateAssessmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	// Verify patient exists
	var patientExists bool
	err := h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM patient WHERE patient_id = $1)", req.PatientID).Scan(&patientExists)
	if err != nil || !patientExists {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid patient",
			Message: fmt.Sprintf("Patient with ID %d does not exist", req.PatientID),
		})
		return
	}

	// Verify clinician exists
	var clinicianExists bool
	err = h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM clinician WHERE clinician_id = $1)", req.ClinicianID).Scan(&clinicianExists)
	if err != nil || !clinicianExists {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid clinician",
			Message: fmt.Sprintf("Clinician with ID %d does not exist", req.ClinicianID),
		})
		return
	}

	// Insert assessment
	var newID int
	err = h.db.QueryRow(`
		INSERT INTO assessment (clinician_id, patient_id, date, location, etiology, 
		                       depth_of_injury, stage, chronicity, healing_status, return_to_clinic)
		VALUES ($1, $2, NOW(), $3, $4, $5, $6, $7, $8, $9)
		RETURNING assessment_id
	`, req.ClinicianID, req.PatientID, req.Location, req.Etiology, req.DepthOfInjury,
		req.Stage, req.Chronicity, req.HealingStatus, req.ReturnToClinic).Scan(&newID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to create assessment",
			Message: err.Error(),
		})
		return
	}

	// Retrieve the created assessment
	var assessment models.Assessment
	err = h.db.QueryRow(`
		SELECT assessment_id, clinician_id, patient_id, date, location, etiology, 
		       depth_of_injury, stage, chronicity, healing_status, return_to_clinic
		FROM assessment
		WHERE assessment_id = $1
	`, newID).Scan(&assessment.AssessmentID, &assessment.ClinicianID, &assessment.PatientID,
		&assessment.Date, &assessment.Location, &assessment.Etiology, &assessment.DepthOfInjury,
		&assessment.Stage, &assessment.Chronicity, &assessment.HealingStatus, &assessment.ReturnToClinic)

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to retrieve created assessment",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, assessment)
}

// CreateFullAssessment creates a complete assessment using the add_full_assessment function
func (h *AssessmentHandler) CreateFullAssessment(c *gin.Context) {
	var req models.FullAssessmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	// Call add_full_assessment PostgreSQL function
	var newID int
	err := h.db.QueryRow(`
		SELECT add_full_assessment(
			$1, $2, $3, $4, $5, $6, $7, $8, $9,
			$10, $11, $12, $13, $14, $15,
			$16, $17, $18, $19, $20, $21,
			$22, $23, $24, $25, $26,
			$27, $28, $29, $30, $31, $32, $33, $34, $35,
			$36, $37, $38,
			$39, $40, $41, $42, $43, $44
		)
	`,
		// Assessment fields
		req.ClinicianID, req.PatientID, req.Location, req.Etiology, req.DepthOfInjury,
		req.Stage, req.Chronicity, req.HealingStatus, req.ReturnToClinic,
		// Infection and pain
		req.InfectionPain.LocalizedSymptoms, req.InfectionPain.SystemicSymptoms,
		req.InfectionPain.PainPresent, req.InfectionPain.PainScore,
		req.InfectionPain.CultureResults, req.InfectionPain.Antibiotic,
		// Tissue status
		req.TissueStatus.GranulationPercent, req.TissueStatus.EpithelialPercent,
		req.TissueStatus.SloughPercent, req.TissueStatus.EscharPercent,
		req.TissueStatus.NecroticPercent, req.TissueStatus.Debridement,
		// Vitals
		req.Vitals.BloodPressure, req.Vitals.Temperature, req.Vitals.Pulse,
		req.Vitals.RespirationRate, req.Vitals.OxygenSaturation,
		// Wound condition
		req.WoundCondition.Length, req.WoundCondition.Width, req.WoundCondition.Depth,
		req.WoundCondition.Tunneling, req.WoundCondition.Undermining,
		req.WoundCondition.Edges, req.WoundCondition.SkinCondition,
		req.WoundCondition.Edema, req.WoundCondition.Blister,
		// Exudate
		req.Exudate.ExudateType, req.Exudate.ExudateAmount, req.Exudate.Odor,
		// Treatment
		req.Treatment.PrimaryDressing, req.Treatment.SecondaryDressing,
		req.Treatment.TertiaryDressing, req.Treatment.Frequency,
		req.Treatment.Supplies, req.Treatment.Orders,
	).Scan(&newID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to create full assessment",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, models.SuccessResponse{
		Message: "Full assessment created successfully",
		Data: map[string]int{
			"assessment_id": newID,
		},
	})
}

// UpdateAssessment updates an existing assessment
func (h *AssessmentHandler) UpdateAssessment(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid assessment ID",
			Message: "Assessment ID must be a valid integer",
		})
		return
	}

	var req models.UpdateAssessmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	// Check if assessment exists
	var exists bool
	err = h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM assessment WHERE assessment_id = $1)", id).Scan(&exists)
	if err != nil || !exists {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "Assessment not found",
			Message: fmt.Sprintf("Assessment with ID %d does not exist", id),
		})
		return
	}

	// Build dynamic update query
	query := "UPDATE assessment SET "
	args := []interface{}{}
	argPos := 1

	if req.Location != "" {
		query += fmt.Sprintf("location = $%d, ", argPos)
		args = append(args, req.Location)
		argPos++
	}
	if req.Etiology != "" {
		query += fmt.Sprintf("etiology = $%d, ", argPos)
		args = append(args, req.Etiology)
		argPos++
	}
	if req.DepthOfInjury != "" {
		query += fmt.Sprintf("depth_of_injury = $%d, ", argPos)
		args = append(args, req.DepthOfInjury)
		argPos++
	}
	if req.Stage != "" {
		query += fmt.Sprintf("stage = $%d, ", argPos)
		args = append(args, req.Stage)
		argPos++
	}
	if req.Chronicity != "" {
		query += fmt.Sprintf("chronicity = $%d, ", argPos)
		args = append(args, req.Chronicity)
		argPos++
	}
	if req.HealingStatus != "" {
		query += fmt.Sprintf("healing_status = $%d, ", argPos)
		args = append(args, req.HealingStatus)
		argPos++
	}
	if req.ReturnToClinic != nil {
		query += fmt.Sprintf("return_to_clinic = $%d, ", argPos)
		args = append(args, *req.ReturnToClinic)
		argPos++
	}

	// Remove trailing comma and space
	query = query[:len(query)-2]
	query += fmt.Sprintf(" WHERE assessment_id = $%d", argPos)
	args = append(args, id)

	_, err = h.db.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to update assessment",
			Message: err.Error(),
		})
		return
	}

	// Retrieve updated assessment
	var assessment models.Assessment
	err = h.db.QueryRow(`
		SELECT assessment_id, clinician_id, patient_id, date, location, etiology, 
		       depth_of_injury, stage, chronicity, healing_status, return_to_clinic
		FROM assessment
		WHERE assessment_id = $1
	`, id).Scan(&assessment.AssessmentID, &assessment.ClinicianID, &assessment.PatientID,
		&assessment.Date, &assessment.Location, &assessment.Etiology, &assessment.DepthOfInjury,
		&assessment.Stage, &assessment.Chronicity, &assessment.HealingStatus, &assessment.ReturnToClinic)

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to retrieve updated assessment",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, assessment)
}

// DeleteAssessment deletes an assessment and all related data
func (h *AssessmentHandler) DeleteAssessment(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid assessment ID",
			Message: "Assessment ID must be a valid integer",
		})
		return
	}

	// Check if assessment exists
	var exists bool
	err = h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM assessment WHERE assessment_id = $1)", id).Scan(&exists)
	if err != nil || !exists {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "Assessment not found",
			Message: fmt.Sprintf("Assessment with ID %d does not exist", id),
		})
		return
	}

	// Delete related data (cascade should handle this, but being explicit)
	_, _ = h.db.Exec("DELETE FROM treatment WHERE assessment_id = $1", id)
	_, _ = h.db.Exec("DELETE FROM exudate WHERE assessment_id = $1", id)
	_, _ = h.db.Exec("DELETE FROM wound_condition WHERE assessment_id = $1", id)
	_, _ = h.db.Exec("DELETE FROM vitals WHERE assessment_id = $1", id)
	_, _ = h.db.Exec("DELETE FROM tissue_status WHERE assessment_id = $1", id)
	_, _ = h.db.Exec("DELETE FROM infection_and_pain WHERE assessment_id = $1", id)

	// Delete assessment
	_, err = h.db.Exec("DELETE FROM assessment WHERE assessment_id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to delete assessment",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: fmt.Sprintf("Assessment with ID %d deleted successfully", id),
	})
}
