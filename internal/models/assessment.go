package models

import "time"

// Assessment represents a wound assessment
type Assessment struct {
	AssessmentID   int       `json:"assessment_id"`
	ClinicianID    int       `json:"clinician_id"`
	PatientID      int       `json:"patient_id"`
	Date           time.Time `json:"date"`
	Location       string    `json:"location"`
	Etiology       string    `json:"etiology"`
	DepthOfInjury  string    `json:"depth_of_injury"`
	Stage          string    `json:"stage"`
	Chronicity     string    `json:"chronicity"`
	HealingStatus  string    `json:"healing_status"`
	ReturnToClinic bool      `json:"return_to_clinic"`
}

// AssessmentFilter holds filter parameters for assessments
type AssessmentFilter struct {
	PatientID   *int   `form:"patient_id"`
	ClinicianID *int   `form:"clinician_id"`
	StartDate   string `form:"start_date"`
	EndDate     string `form:"end_date"`
	PaginationParams
}

// CreateAssessmentRequest represents request for creating an assessment
type CreateAssessmentRequest struct {
	ClinicianID    int    `json:"clinician_id" binding:"required"`
	PatientID      int    `json:"patient_id" binding:"required"`
	Location       string `json:"location" binding:"required,max=50"`
	Etiology       string `json:"etiology" binding:"required,max=50"`
	DepthOfInjury  string `json:"depth_of_injury" binding:"required,max=50"`
	Stage          string `json:"stage" binding:"required,max=15"`
	Chronicity     string `json:"chronicity" binding:"required,max=15"`
	HealingStatus  string `json:"healing_status" binding:"required,max=20"`
	ReturnToClinic bool   `json:"return_to_clinic"`
}

// FullAssessmentRequest includes all related data
type FullAssessmentRequest struct {
	CreateAssessmentRequest
	InfectionPain  InfectionPainRequest  `json:"infection_pain" binding:"required"`
	TissueStatus   TissueStatusRequest   `json:"tissue_status" binding:"required"`
	Vitals         VitalsRequest         `json:"vitals" binding:"required"`
	WoundCondition WoundConditionRequest `json:"wound_condition" binding:"required"`
	Exudate        ExudateRequest        `json:"exudate" binding:"required"`
	Treatment      TreatmentRequest      `json:"treatment" binding:"required"`
}

// InfectionPainRequest for infection and pain data
type InfectionPainRequest struct {
	LocalizedSymptoms string `json:"localized_symptoms" binding:"required,max=20"`
	SystemicSymptoms  string `json:"systemic_symptoms" binding:"required,max=20"`
	PainPresent       string `json:"pain_present" binding:"required,max=20"`
	PainScore         string `json:"pain_score" binding:"required,max=20"`
	CultureResults    string `json:"culture_results" binding:"required,max=20"`
	Antibiotic        string `json:"antibiotic" binding:"required,max=20"`
}

// TissueStatusRequest for tissue status data
type TissueStatusRequest struct {
	GranulationPercent int    `json:"granulation_percent" binding:"min=0,max=100"`
	EpithelialPercent  int    `json:"epithelial_percent" binding:"min=0,max=100"`
	SloughPercent      int    `json:"slough_percent" binding:"min=0,max=100"`
	EscharPercent      int    `json:"eschar_percent" binding:"min=0,max=100"`
	NecroticPercent    int    `json:"necrotic_percent" binding:"min=0,max=100"`
	Debridement        string `json:"debridement" binding:"required,max=15"`
}

// VitalsRequest for vital signs
type VitalsRequest struct {
	BloodPressure    string  `json:"blood_pressure" binding:"required,max=10"`
	Temperature      float64 `json:"temperature" binding:"required,min=30,max=45"`
	Pulse            int     `json:"pulse" binding:"required,min=30,max=200"`
	RespirationRate  int     `json:"respiration_rate" binding:"required,min=5,max=60"`
	OxygenSaturation int     `json:"oxygen_saturation" binding:"required,min=50,max=100"`
}

// WoundConditionRequest for wound condition
type WoundConditionRequest struct {
	Length        float64 `json:"length" binding:"required,min=0"`
	Width         float64 `json:"width" binding:"required,min=0"`
	Depth         float64 `json:"depth" binding:"required,min=0"`
	Tunneling     bool    `json:"tunneling"`
	Undermining   bool    `json:"undermining"`
	Edges         string  `json:"edges" binding:"required,max=15"`
	SkinCondition string  `json:"skin_condition" binding:"required,max=20"`
	Edema         string  `json:"edema" binding:"required,max=20"`
	Blister       string  `json:"blister" binding:"required,max=20"`
}

// ExudateRequest for exudate data
type ExudateRequest struct {
	ExudateType   string `json:"exudate_type" binding:"required,max=20"`
	ExudateAmount string `json:"exudate_amount" binding:"required,max=20"`
	Odor          string `json:"odor" binding:"required,max=20"`
}

// TreatmentRequest for treatment data
type TreatmentRequest struct {
	PrimaryDressing   string `json:"primary_dressing" binding:"required,max=15"`
	SecondaryDressing string `json:"secondary_dressing" binding:"required,max=15"`
	TertiaryDressing  string `json:"tertiary_dressing" binding:"required,max=15"`
	Frequency         string `json:"frequency" binding:"required,max=15"`
	Supplies          string `json:"supplies" binding:"max=500"`
	Orders            string `json:"orders" binding:"max=200"`
}

// UpdateAssessmentRequest for updating assessment
type UpdateAssessmentRequest struct {
	Location       string `json:"location" binding:"omitempty,max=50"`
	Etiology       string `json:"etiology" binding:"omitempty,max=50"`
	DepthOfInjury  string `json:"depth_of_injury" binding:"omitempty,max=50"`
	Stage          string `json:"stage" binding:"omitempty,max=15"`
	Chronicity     string `json:"chronicity" binding:"omitempty,max=15"`
	HealingStatus  string `json:"healing_status" binding:"omitempty,max=20"`
	ReturnToClinic *bool  `json:"return_to_clinic"`
}

// FullAssessmentResponse includes all related data
type FullAssessmentResponse struct {
	AssessmentID       int       `json:"assessment_id"`
	AssessmentDate     time.Time `json:"assessment_date"`
	PatientID          int       `json:"patient_id"`
	PatientName        string    `json:"patient_name"`
	ClinicianID        int       `json:"clinician_id"`
	ClinicianName      string    `json:"clinician_name"`
	Location           string    `json:"location"`
	Etiology           string    `json:"etiology"`
	Stage              string    `json:"stage"`
	HealingStatus      string    `json:"healing_status"`
	PainScore          string    `json:"pain_score"`
	GranulationPercent int       `json:"granulation_percent"`
	Length             float64   `json:"length"`
	Width              float64   `json:"width"`
}
