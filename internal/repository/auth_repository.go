package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/vellalasantosh/wound_iq_api_claude/internal/models"
	"github.com/vellalasantosh/wound_iq_api_claude/internal/utils"
)

type AuthRepository struct {
	db *sql.DB
}

func NewAuthRepository(db *sql.DB) *AuthRepository {
	return &AuthRepository{db: db}
}

// ------------------------------------------------------------
// CREATE USER + PROFILE (PATIENT / CLINICIAN)
// ------------------------------------------------------------
func (r *AuthRepository) CreateUser(email, password, role, firstName, lastName string) (*models.User, error) {

	// Hash password
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, err
	}

	fullName := firstName + " " + lastName

	// Start transaction
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	//----------------------------------------------------------
	// 1. Insert into USERS table
	//----------------------------------------------------------
	var user models.User

	err = tx.QueryRow(`
		INSERT INTO Users (email, password_hash, role, is_active, email_verified)
		VALUES ($1, $2, $3, true, false)
		RETURNING id, email, role, is_active, email_verified, created_at, updated_at
	`, email, hashedPassword, role).Scan(
		&user.ID, &user.Email, &user.Role, &user.IsActive,
		&user.EmailVerified, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	//----------------------------------------------------------
	// 2. Insert into role-specific profile table
	//----------------------------------------------------------
	switch role {

	// ----------------------------------------------------------
	// PATIENT PROFILE
	// ----------------------------------------------------------
	case "patient":
		_, err = tx.Exec(`
			INSERT INTO Patient (
				user_id, first_name, last_name, full_name,
				date_of_birth, gender, medical_record_number
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`,
			user.ID,
			firstName,
			lastName,
			fullName,
			"1900-01-01",               // default DOB
			"Unknown",                  // default gender
			"MRN-"+fmt.Sprint(user.ID), // generated MRN
		)

	// ----------------------------------------------------------
	// CLINICIAN PROFILE
	// ----------------------------------------------------------
	case "clinician":
		_, err = tx.Exec(`
			INSERT INTO Clinician (
				user_id, first_name, last_name, full_name,
				role, department, contact_info, license_number
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`,
			user.ID,
			firstName,
			lastName,
			fullName,
			"Clinician",                // default job title
			"General Medicine",         // default department
			"Not Provided",             // default contact info
			"LIC-"+fmt.Sprint(user.ID), // generated license number
		)

	default:
		err = errors.New("invalid role specified")
	}

	if err != nil {
		return nil, err
	}

	// Commit
	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return &user, nil
}

// ------------------------------------------------------------
// GET USER BY EMAIL
// ------------------------------------------------------------
func (r *AuthRepository) GetUserByEmail(email string) (*models.User, error) {
	var user models.User

	err := r.db.QueryRow(`
		SELECT id, email, password_hash, role, is_active, email_verified, created_at, updated_at
		FROM Users
		WHERE email = $1
	`, email).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Role,
		&user.IsActive, &user.EmailVerified, &user.CreatedAt, &user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, utils.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// ------------------------------------------------------------
// GET USER BY ID
// ------------------------------------------------------------
func (r *AuthRepository) GetUserByID(userID int) (*models.User, error) {
	var user models.User

	err := r.db.QueryRow(`
		SELECT id, email, password_hash, role, is_active, email_verified, created_at, updated_at
		FROM Users
		WHERE id = $1
	`, userID).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Role,
		&user.IsActive, &user.EmailVerified, &user.CreatedAt, &user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, utils.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// ------------------------------------------------------------
// GET USER WITH PROFILE (PATIENT / CLINICIAN)
// ------------------------------------------------------------
// GetUserWithProfile - UPDATED VERSION
func (r *AuthRepository) GetUserWithProfile(userID int) (*models.UserWithProfile, error) {
	var profile models.UserWithProfile

	// Get basic user info from users table
	user, err := r.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Fill basic user info
	profile.ID = user.ID
	profile.Email = user.Email
	profile.Role = user.Role
	profile.IsActive = user.IsActive
	profile.EmailVerified = user.EmailVerified
	profile.CreatedAt = user.CreatedAt
	profile.UpdatedAt = user.UpdatedAt

	// Fetch role-specific profile data
	var firstName, lastName sql.NullString

	switch user.Role {
	case "patient":
		err = r.db.QueryRow(`
			SELECT first_name, last_name
			FROM Patient
			WHERE user_id = $1
		`, userID).Scan(&firstName, &lastName)

		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("patient profile not found for user_id %d", userID)
		}

	case "clinician":
		err = r.db.QueryRow(`
			SELECT first_name, last_name
			FROM Clinician
			WHERE user_id = $1
		`, userID).Scan(&firstName, &lastName)

		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("clinician profile not found for user_id %d", userID)
		}

	default:
		return nil, fmt.Errorf("unknown role: %s", user.Role)
	}

	// Handle any other database errors
	if err != nil {
		return nil, fmt.Errorf("error fetching profile: %w", err)
	}

	// Set names (handle NULL values)
	profile.FirstName = firstName.String
	profile.LastName = lastName.String

	return &profile, nil
}

// ------------------------------------------------------------
// SAVE REFRESH TOKEN
// ------------------------------------------------------------
func (r *AuthRepository) SaveRefreshToken(userID int, token string, expiresAt time.Time) error {
	_, err := r.db.Exec(`
		INSERT INTO refresh_tokens (user_id, token, expires_at)
		VALUES ($1, $2, $3)
	`, userID, token, expiresAt)
	return err
}

// ------------------------------------------------------------
// VALIDATE REFRESH TOKEN
// ------------------------------------------------------------
func (r *AuthRepository) ValidateRefreshToken(token string) (*models.RefreshToken, error) {
	var rt models.RefreshToken

	err := r.db.QueryRow(`
		SELECT id, user_id, token, expires_at, created_at, revoked
		FROM refresh_tokens
		WHERE token = $1 AND revoked = false AND expires_at > NOW()
	`, token).Scan(
		&rt.ID, &rt.UserID, &rt.Token, &rt.ExpiresAt,
		&rt.CreatedAt, &rt.Revoked,
	)

	if err == sql.ErrNoRows {
		return nil, utils.ErrInvalidToken
	}
	if err != nil {
		return nil, err
	}

	return &rt, nil
}

// ------------------------------------------------------------
// REVOKE REFRESH TOKEN
// ------------------------------------------------------------
func (r *AuthRepository) RevokeRefreshToken(token string) error {
	result, err := r.db.Exec(`
		UPDATE refresh_tokens
		SET revoked = true
		WHERE token = $1
	`, token)
	if err != nil {
		return err
	}

	n, _ := result.RowsAffected()
	if n == 0 {
		return errors.New("token not found")
	}
	return nil
}

// ------------------------------------------------------------
// REVOKE ALL TOKENS FOR USER
// ------------------------------------------------------------
func (r *AuthRepository) RevokeAllUserTokens(userID int) error {
	_, err := r.db.Exec(`
		UPDATE refresh_tokens
		SET revoked = true
		WHERE user_id = $1 AND revoked = false
	`, userID)
	return err
}

// ------------------------------------------------------------
// UPDATE PASSWORD
// ------------------------------------------------------------
func (r *AuthRepository) UpdatePassword(userID int, newPassword string) error {
	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(`
		UPDATE Users
		SET password_hash = $1, updated_at = NOW()
		WHERE id = $2
	`, hashedPassword, userID)

	return err
}

// ------------------------------------------------------------
// CHECK EMAIL EXISTS
// ------------------------------------------------------------
func (r *AuthRepository) EmailExists(email string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM Users WHERE email = $1)
	`, email).Scan(&exists)
	return exists, err
}
