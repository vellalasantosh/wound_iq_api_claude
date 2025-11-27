package models

import "time"

// Patient represents a patient in the system
type Patient struct {
	PatientID           int       `json:"patient_id"`
	FullName            string    `json:"full_name"`
	DateOfBirth         time.Time `json:"date_of_birth"`
	Gender              string    `json:"gender"`
	MedicalRecordNumber string    `json:"medical_record_number"`
}

// CreatePatientRequest represents the request body for creating a patient
type CreatePatientRequest struct {
	FullName            string `json:"full_name" binding:"required,min=2,max=100"`
	DateOfBirth         string `json:"date_of_birth" binding:"required"` // ISO-8601 format
	Gender              string `json:"gender" binding:"required,oneof=Male Female Other"`
	MedicalRecordNumber string `json:"medical_record_number" binding:"required,min=1,max=50"`
}

// UpdatePatientRequest represents the request body for updating a patient
type UpdatePatientRequest struct {
	FullName            string `json:"full_name" binding:"omitempty,min=2,max=100"`
	DateOfBirth         string `json:"date_of_birth" binding:"omitempty"`
	Gender              string `json:"gender" binding:"omitempty,oneof=Male Female Other"`
	MedicalRecordNumber string `json:"medical_record_number" binding:"omitempty,min=1,max=50"`
}

// WoundHistory represents a simplified wound history entry
type WoundHistory struct {
	AssessmentID   int       `json:"assessment_id"`
	AssessmentDate time.Time `json:"assessment_date"`
	Location       string    `json:"location"`
	Stage          string    `json:"stage"`
	HealingStatus  string    `json:"healing_status"`
}
