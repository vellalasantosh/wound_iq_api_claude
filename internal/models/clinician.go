package models

// Clinician represents a clinician in the system
type Clinician struct {
	ClinicianID   int    `json:"clinician_id"`
	FullName      string `json:"full_name"`
	Role          string `json:"role"`
	Department    string `json:"department"`
	ContactInfo   string `json:"contact_info"`
	LicenseNumber string `json:"license_number"`
}

// CreateClinicianRequest represents the request body for creating a clinician
type CreateClinicianRequest struct {
	FullName      string `json:"full_name" binding:"required,min=2,max=100"`
	Role          string `json:"role" binding:"required,min=2,max=20"`
	Department    string `json:"department" binding:"required,min=2,max=20"`
	ContactInfo   string `json:"contact_info" binding:"required,min=5,max=500"`
	LicenseNumber string `json:"license_number" binding:"required,min=3,max=50"`
}

// UpdateClinicianRequest represents the request body for updating a clinician
type UpdateClinicianRequest struct {
	FullName      string `json:"full_name" binding:"omitempty,min=2,max=100"`
	Role          string `json:"role" binding:"omitempty,min=2,max=20"`
	Department    string `json:"department" binding:"omitempty,min=2,max=20"`
	ContactInfo   string `json:"contact_info" binding:"omitempty,min=5,max=500"`
	LicenseNumber string `json:"license_number" binding:"omitempty,min=3,max=50"`
}
