package handlers

import (
	"database/sql"
	"fmt"
	"math"
	"net/http"
	"strconv"

	"github.com/vellalasantosh/wound_iq_api_claude/internal/db"
	"github.com/vellalasantosh/wound_iq_api_claude/internal/models"

	"github.com/gin-gonic/gin"
)

// ClinicianHandler handles clinician-related requests
type ClinicianHandler struct {
	db *db.DB
}

// NewClinicianHandler creates a new clinician handler
func NewClinicianHandler(database *db.DB) *ClinicianHandler {
	return &ClinicianHandler{db: database}
}

// GetAllClinicians retrieves all clinicians with pagination
func (h *ClinicianHandler) GetAllClinicians(c *gin.Context) {
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
	err := h.db.QueryRow("SELECT COUNT(*) FROM clinician").Scan(&totalCount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to count clinicians",
			Message: err.Error(),
		})
		return
	}

	// Query clinicians with pagination
	rows, err := h.db.Query(`
		SELECT clinician_id, full_name, role, department, contact_info, license_number
		FROM clinician
		ORDER BY full_name
		LIMIT $1 OFFSET $2
	`, params.GetLimit(), params.GetOffset())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to query clinicians",
			Message: err.Error(),
		})
		return
	}
	defer rows.Close()

	var clinicians []models.Clinician
	for rows.Next() {
		var cl models.Clinician
		if err := rows.Scan(&cl.ClinicianID, &cl.FullName, &cl.Role, &cl.Department, &cl.ContactInfo, &cl.LicenseNumber); err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "Failed to scan clinician",
				Message: err.Error(),
			})
			return
		}
		clinicians = append(clinicians, cl)
	}

	totalPages := int(math.Ceil(float64(totalCount) / float64(params.GetLimit())))

	c.JSON(http.StatusOK, models.PaginatedResponse{
		Data:       clinicians,
		Page:       params.Page,
		PageSize:   params.GetLimit(),
		TotalCount: totalCount,
		TotalPages: totalPages,
	})
}

// GetClinicianByID retrieves a single clinician by ID
func (h *ClinicianHandler) GetClinicianByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid clinician ID",
			Message: "Clinician ID must be a valid integer",
		})
		return
	}

	var clinician models.Clinician
	err = h.db.QueryRow(`
		SELECT clinician_id, full_name, role, department, contact_info, license_number
		FROM clinician
		WHERE clinician_id = $1
	`, id).Scan(&clinician.ClinicianID, &clinician.FullName, &clinician.Role, &clinician.Department, &clinician.ContactInfo, &clinician.LicenseNumber)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "Clinician not found",
			Message: fmt.Sprintf("Clinician with ID %d does not exist", id),
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to query clinician",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, clinician)
}

// CreateClinician creates a new clinician
func (h *ClinicianHandler) CreateClinician(c *gin.Context) {
	var req models.CreateClinicianRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	// Insert clinician
	var newID int
	err := h.db.QueryRow(`
		INSERT INTO clinician (full_name, role, department, contact_info, license_number)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING clinician_id
	`, req.FullName, req.Role, req.Department, req.ContactInfo, req.LicenseNumber).Scan(&newID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to create clinician",
			Message: err.Error(),
		})
		return
	}

	// Retrieve the created clinician
	var clinician models.Clinician
	err = h.db.QueryRow(`
		SELECT clinician_id, full_name, role, department, contact_info, license_number
		FROM clinician
		WHERE clinician_id = $1
	`, newID).Scan(&clinician.ClinicianID, &clinician.FullName, &clinician.Role, &clinician.Department, &clinician.ContactInfo, &clinician.LicenseNumber)

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to retrieve created clinician",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, clinician)
}

// UpdateClinician updates an existing clinician
func (h *ClinicianHandler) UpdateClinician(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid clinician ID",
			Message: "Clinician ID must be a valid integer",
		})
		return
	}

	var req models.UpdateClinicianRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	// Check if clinician exists
	var exists bool
	err = h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM clinician WHERE clinician_id = $1)", id).Scan(&exists)
	if err != nil || !exists {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "Clinician not found",
			Message: fmt.Sprintf("Clinician with ID %d does not exist", id),
		})
		return
	}

	// Build dynamic update query
	query := "UPDATE clinician SET "
	args := []interface{}{}
	argPos := 1

	if req.FullName != "" {
		query += fmt.Sprintf("full_name = $%d, ", argPos)
		args = append(args, req.FullName)
		argPos++
	}
	if req.Role != "" {
		query += fmt.Sprintf("role = $%d, ", argPos)
		args = append(args, req.Role)
		argPos++
	}
	if req.Department != "" {
		query += fmt.Sprintf("department = $%d, ", argPos)
		args = append(args, req.Department)
		argPos++
	}
	if req.ContactInfo != "" {
		query += fmt.Sprintf("contact_info = $%d, ", argPos)
		args = append(args, req.ContactInfo)
		argPos++
	}
	if req.LicenseNumber != "" {
		query += fmt.Sprintf("license_number = $%d, ", argPos)
		args = append(args, req.LicenseNumber)
		argPos++
	}

	// Remove trailing comma and space
	query = query[:len(query)-2]
	query += fmt.Sprintf(" WHERE clinician_id = $%d", argPos)
	args = append(args, id)

	_, err = h.db.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to update clinician",
			Message: err.Error(),
		})
		return
	}

	// Retrieve updated clinician
	var clinician models.Clinician
	err = h.db.QueryRow(`
		SELECT clinician_id, full_name, role, department, contact_info, license_number
		FROM clinician
		WHERE clinician_id = $1
	`, id).Scan(&clinician.ClinicianID, &clinician.FullName, &clinician.Role, &clinician.Department, &clinician.ContactInfo, &clinician.LicenseNumber)

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to retrieve updated clinician",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, clinician)
}

// DeleteClinician deletes a clinician
func (h *ClinicianHandler) DeleteClinician(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid clinician ID",
			Message: "Clinician ID must be a valid integer",
		})
		return
	}

	// Check if clinician exists
	var exists bool
	err = h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM clinician WHERE clinician_id = $1)", id).Scan(&exists)
	if err != nil || !exists {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "Clinician not found",
			Message: fmt.Sprintf("Clinician with ID %d does not exist", id),
		})
		return
	}

	// Delete clinician
	_, err = h.db.Exec("DELETE FROM clinician WHERE clinician_id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to delete clinician",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: fmt.Sprintf("Clinician with ID %d deleted successfully", id),
	})
}
