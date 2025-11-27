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

// PatientHandler handles patient-related requests
type PatientHandler struct {
	db *db.DB
}

// NewPatientHandler creates a new patient handler
func NewPatientHandler(database *db.DB) *PatientHandler {
	return &PatientHandler{db: database}
}

// GetAllPatients retrieves all patients with pagination
// @Summary Get all patients
// @Tags patients
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Success 200 {object} models.PaginatedResponse
// @Router /v1/patients [get]
func (h *PatientHandler) GetAllPatients(c *gin.Context) {
	var params models.PaginationParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid pagination parameters",
			Message: err.Error(),
		})
		return
	}

	// Get total count
	var totalCount int
	err := h.db.QueryRow("SELECT COUNT(*) FROM patient").Scan(&totalCount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to count patients",
			Message: err.Error(),
		})
		return
	}

	// Query patients with pagination
	rows, err := h.db.Query(`
		SELECT patient_id, full_name, date_of_birth, gender, medical_record_number
		FROM patient
		ORDER BY full_name
		LIMIT $1 OFFSET $2
	`, params.GetLimit(), params.GetOffset())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to query patients",
			Message: err.Error(),
		})
		return
	}
	defer rows.Close()

	var patients []models.Patient
	for rows.Next() {
		var p models.Patient
		if err := rows.Scan(&p.PatientID, &p.FullName, &p.DateOfBirth, &p.Gender, &p.MedicalRecordNumber); err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "Failed to scan patient",
				Message: err.Error(),
			})
			return
		}
		patients = append(patients, p)
	}

	totalPages := int(math.Ceil(float64(totalCount) / float64(params.GetLimit())))

	c.JSON(http.StatusOK, models.PaginatedResponse{
		Data:       patients,
		Page:       params.Page,
		PageSize:   params.GetLimit(),
		TotalCount: totalCount,
		TotalPages: totalPages,
	})
}

// GetPatientByID retrieves a single patient by ID
// @Summary Get patient by ID
// @Tags patients
// @Produce json
// @Param id path int true "Patient ID"
// @Success 200 {object} models.Patient
// @Router /v1/patients/{id} [get]
func (h *PatientHandler) GetPatientByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid patient ID",
			Message: "Patient ID must be a valid integer",
		})
		return
	}

	var patient models.Patient
	err = h.db.QueryRow(`
		SELECT patient_id, full_name, date_of_birth, gender, medical_record_number
		FROM patient
		WHERE patient_id = $1
	`, id).Scan(&patient.PatientID, &patient.FullName, &patient.DateOfBirth, &patient.Gender, &patient.MedicalRecordNumber)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "Patient not found",
			Message: fmt.Sprintf("Patient with ID %d does not exist", id),
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to query patient",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, patient)
}

// CreatePatient creates a new patient using the add_patient function
// @Summary Create a new patient
// @Tags patients
// @Accept json
// @Produce json
// @Param patient body models.CreatePatientRequest true "Patient data"
// @Success 201 {object} models.Patient
// @Router /v1/patients [post]
func (h *PatientHandler) CreatePatient(c *gin.Context) {
	var req models.CreatePatientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	// Parse date of birth
	dob, err := time.Parse(time.RFC3339, req.DateOfBirth)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid date format",
			Message: "Date of birth must be in ISO-8601 format (e.g., 2000-01-15T00:00:00Z)",
		})
		return
	}

	// Call the add_patient PostgreSQL function
	var newID int
	err = h.db.QueryRow(`
		SELECT add_patient($1, $2, $3, $4)
	`, req.FullName, dob, req.Gender, req.MedicalRecordNumber).Scan(&newID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to create patient",
			Message: err.Error(),
		})
		return
	}

	// Retrieve the created patient
	var patient models.Patient
	err = h.db.QueryRow(`
		SELECT patient_id, full_name, date_of_birth, gender, medical_record_number
		FROM patient
		WHERE patient_id = $1
	`, newID).Scan(&patient.PatientID, &patient.FullName, &patient.DateOfBirth, &patient.Gender, &patient.MedicalRecordNumber)

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to retrieve created patient",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, patient)
}

// UpdatePatient updates an existing patient
// @Summary Update a patient
// @Tags patients
// @Accept json
// @Produce json
// @Param id path int true "Patient ID"
// @Param patient body models.UpdatePatientRequest true "Patient data"
// @Success 200 {object} models.Patient
// @Router /v1/patients/{id} [put]
func (h *PatientHandler) UpdatePatient(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid patient ID",
			Message: "Patient ID must be a valid integer",
		})
		return
	}

	var req models.UpdatePatientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	// Check if patient exists
	var exists bool
	err = h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM patient WHERE patient_id = $1)", id).Scan(&exists)
	if err != nil || !exists {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "Patient not found",
			Message: fmt.Sprintf("Patient with ID %d does not exist", id),
		})
		return
	}

	// Build dynamic update query
	query := "UPDATE patient SET "
	args := []interface{}{}
	argPos := 1

	if req.FullName != "" {
		query += fmt.Sprintf("full_name = $%d, ", argPos)
		args = append(args, req.FullName)
		argPos++
	}
	if req.DateOfBirth != "" {
		dob, err := time.Parse(time.RFC3339, req.DateOfBirth)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "Invalid date format",
				Message: "Date of birth must be in ISO-8601 format",
			})
			return
		}
		query += fmt.Sprintf("date_of_birth = $%d, ", argPos)
		args = append(args, dob)
		argPos++
	}
	if req.Gender != "" {
		query += fmt.Sprintf("gender = $%d, ", argPos)
		args = append(args, req.Gender)
		argPos++
	}
	if req.MedicalRecordNumber != "" {
		query += fmt.Sprintf("medical_record_number = $%d, ", argPos)
		args = append(args, req.MedicalRecordNumber)
		argPos++
	}

	// Remove trailing comma and space
	query = query[:len(query)-2]
	query += fmt.Sprintf(" WHERE patient_id = $%d", argPos)
	args = append(args, id)

	_, err = h.db.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to update patient",
			Message: err.Error(),
		})
		return
	}

	// Retrieve updated patient
	var patient models.Patient
	err = h.db.QueryRow(`
		SELECT patient_id, full_name, date_of_birth, gender, medical_record_number
		FROM patient
		WHERE patient_id = $1
	`, id).Scan(&patient.PatientID, &patient.FullName, &patient.DateOfBirth, &patient.Gender, &patient.MedicalRecordNumber)

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to retrieve updated patient",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, patient)
}

// DeletePatient deletes a patient
// @Summary Delete a patient
// @Tags patients
// @Produce json
// @Param id path int true "Patient ID"
// @Success 200 {object} models.SuccessResponse
// @Router /v1/patients/{id} [delete]
func (h *PatientHandler) DeletePatient(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid patient ID",
			Message: "Patient ID must be a valid integer",
		})
		return
	}

	// Check if patient exists
	var exists bool
	err = h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM patient WHERE patient_id = $1)", id).Scan(&exists)
	if err != nil || !exists {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "Patient not found",
			Message: fmt.Sprintf("Patient with ID %d does not exist", id),
		})
		return
	}

	// Delete patient (will cascade to assessments due to FK constraints)
	_, err = h.db.Exec("DELETE FROM patient WHERE patient_id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to delete patient",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: fmt.Sprintf("Patient with ID %d deleted successfully", id),
	})
}
