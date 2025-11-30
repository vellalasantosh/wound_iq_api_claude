package service

import (
	"fmt"
	"log"
	"time"

	"github.com/vellalasantosh/wound_iq_api_claude/internal/models"
	"github.com/vellalasantosh/wound_iq_api_claude/internal/repository"
	"github.com/vellalasantosh/wound_iq_api_claude/internal/utils"
)

type AuthService struct {
	authRepo *repository.AuthRepository
}

func NewAuthService(authRepo *repository.AuthRepository) *AuthService {
	return &AuthService{authRepo: authRepo}
}

// Register creates a new user account
func (s *AuthService) Register(req models.RegisterRequest) (*models.LoginResponse, error) {
	// Validate password strength
	if err := utils.ValidatePasswordStrength(req.Password); err != nil {
		return nil, err
	}

	// Check if email already exists
	exists, err := s.authRepo.EmailExists(req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, utils.ErrEmailExists
	}

	// Create user (this creates both user and profile records)
	user, err := s.authRepo.CreateUser(
		req.Email,
		req.Password,
		req.Role,
		req.FirstName,
		req.LastName,
	)
	if err != nil {
		return nil, err
	}

	// Generate tokens
	accessToken, err := utils.GenerateAccessToken(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, err
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}

	// Save refresh token
	expiresAt := time.Now().Add(utils.RefreshTokenExpiry)
	if err := s.authRepo.SaveRefreshToken(user.ID, refreshToken, expiresAt); err != nil {
		return nil, err
	}

	// Get user profile (should always succeed since we just created it)
	profile, err := s.authRepo.GetUserWithProfile(user.ID)
	if err != nil {
		return nil, fmt.Errorf("user created but profile fetch failed: %w", err)
	}

	return &models.LoginResponse{
		User:         *profile,
		Token:        accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// Login authenticates a user and returns tokens
// Authentication uses ONLY the users table
// Profile is fetched separately for display purposes
func (s *AuthService) Login(req models.LoginRequest) (*models.LoginResponse, error) {

	log.Printf("[AUTH] Login attempt for email: %s", req.Email)

	// ============================================================
	// AUTHENTICATION PHASE - Uses ONLY users table
	// ============================================================

	// Step 1: Get user by email from users table
	user, err := s.authRepo.GetUserByEmail(req.Email)
	if err != nil {
		log.Printf("[AUTH] Login failed: User not found - %s", req.Email)
		return nil, utils.ErrInvalidCredentials
	}

	log.Printf("[AUTH] User found: ID=%d, Email=%s, Role=%s", user.ID, user.Email, user.Role)

	// Step 2: Check if user is active
	if !user.IsActive {
		log.Printf("[AUTH] Login failed: User inactive - %s", req.Email)
		return nil, utils.ErrUserInactive
	}

	// Step 3: Verify password
	if !utils.CheckPassword(req.Password, user.PasswordHash) {
		log.Printf("[AUTH] Login failed: Invalid password - %s", req.Email)
		return nil, utils.ErrInvalidCredentials
	}

	log.Printf("[AUTH] ✅ Authentication successful for user: %s", req.Email)

	// ============================================================
	// AUTHENTICATION COMPLETE
	// Everything below uses authenticated user info
	// ============================================================

	// Step 4: Generate tokens
	accessToken, err := utils.GenerateAccessToken(user.ID, user.Email, user.Role)
	if err != nil {
		log.Printf("[AUTH] Failed to generate access token: %v", err)
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID)
	if err != nil {
		log.Printf("[AUTH] Failed to generate refresh token: %v", err)
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Step 5: Save refresh token
	expiresAt := time.Now().Add(utils.RefreshTokenExpiry)
	if err := s.authRepo.SaveRefreshToken(user.ID, refreshToken, expiresAt); err != nil {
		log.Printf("[AUTH] Warning: Failed to save refresh token: %v", err)
		// Continue anyway - user can still use access token
	}

	// ============================================================
	// PROFILE FETCHING - Separate from authentication
	// ============================================================

	// Step 6: Fetch profile from clinicians/patients table
	log.Printf("[AUTH] Fetching profile for user: %d", user.ID)
	profile, err := s.authRepo.GetUserWithProfile(user.ID)
	if err != nil {
		// This is a DATA INTEGRITY issue - user exists but no profile
		log.Printf("[AUTH] ❌ ERROR: User %d (%s) has no profile in %s table!",
			user.ID, user.Email, user.Role)
		log.Printf("[AUTH] This indicates a data integrity problem")

		// Return error with helpful message
		return nil, fmt.Errorf(
			"authentication successful but profile not found - "+
				"user_id %d exists in users table but has no matching record in %s table. "+
				"Please contact administrator",
			user.ID, user.Role,
		)
	}

	log.Printf("[AUTH] ✅ Profile fetched: %s %s", profile.FirstName, profile.LastName)
	log.Printf("[AUTH] ✅ Login complete for: %s", req.Email)

	return &models.LoginResponse{
		User:         *profile,
		Token:        accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// RefreshToken generates new tokens using a refresh token
func (s *AuthService) RefreshToken(refreshToken string) (*models.LoginResponse, error) {
	// Validate refresh token
	rt, err := s.authRepo.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, utils.ErrInvalidToken
	}

	// Get user
	user, err := s.authRepo.GetUserByID(rt.UserID)
	if err != nil {
		return nil, err
	}

	if !user.IsActive {
		return nil, utils.ErrUserInactive
	}

	// Generate new tokens
	newAccessToken, err := utils.GenerateAccessToken(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := utils.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}

	// Revoke old refresh token
	if err := s.authRepo.RevokeRefreshToken(refreshToken); err != nil {
		return nil, err
	}

	// Save new refresh token
	expiresAt := time.Now().Add(utils.RefreshTokenExpiry)
	if err := s.authRepo.SaveRefreshToken(user.ID, newRefreshToken, expiresAt); err != nil {
		return nil, err
	}

	// Get user profile (required)
	profile, err := s.authRepo.GetUserWithProfile(user.ID)
	if err != nil {
		return nil, fmt.Errorf("user exists but profile not found: %w", err)
	}

	return &models.LoginResponse{
		User:         *profile,
		Token:        newAccessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

// Logout revokes all user tokens
func (s *AuthService) Logout(userID int) error {
	return s.authRepo.RevokeAllUserTokens(userID)
}

// ChangePassword changes user password
func (s *AuthService) ChangePassword(userID int, req models.ChangePasswordRequest) error {
	// Get user from users table only
	user, err := s.authRepo.GetUserByID(userID)
	if err != nil {
		return err
	}

	// Verify old password
	if !utils.CheckPassword(req.OldPassword, user.PasswordHash) {
		return utils.ErrInvalidCredentials
	}

	// Validate new password
	if err := utils.ValidatePasswordStrength(req.NewPassword); err != nil {
		return err
	}

	// Update password in users table
	return s.authRepo.UpdatePassword(userID, req.NewPassword)
}

// GetUserProfile gets user profile by ID
// This is used by the /profile endpoint
func (s *AuthService) GetUserProfile(userID int) (*models.UserWithProfile, error) {
	profile, err := s.authRepo.GetUserWithProfile(userID)
	if err != nil {
		return nil, fmt.Errorf("profile not found for user %d: %w", userID, err)
	}
	return profile, nil
}
