package models

import "time"

// PaginationParams holds pagination parameters
type PaginationParams struct {
	Page     int `form:"page" binding:"omitempty,min=1"`
	PageSize int `form:"page_size" binding:"omitempty,min=1,max=100"`
}

// GetOffset calculates the offset for SQL queries
func (p *PaginationParams) GetOffset() int {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.PageSize <= 0 {
		p.PageSize = 10
	}
	return (p.Page - 1) * p.PageSize
}

// GetLimit returns the page size
func (p *PaginationParams) GetLimit() int {
	if p.PageSize <= 0 {
		p.PageSize = 10
	}
	return p.PageSize
}

// PaginatedResponse wraps paginated data
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalCount int         `json:"total_count"`
	TotalPages int         `json:"total_pages"`
}

// ErrorResponse represents an API error
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// SuccessResponse represents a success message
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// NullTime is a wrapper for time.Time that can be null
type NullTime struct {
	Time  time.Time
	Valid bool
}

// MarshalJSON implements json.Marshaler
func (nt NullTime) MarshalJSON() ([]byte, error) {
	if !nt.Valid {
		return []byte("null"), nil
	}
	return []byte(`"` + nt.Time.Format(time.RFC3339) + `"`), nil
}
