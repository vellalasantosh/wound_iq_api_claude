package service

import (
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

	// Create user
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

	// Get user profile
	profile, err := s.authRepo.GetUserWithProfile(user.ID)
	if err != nil {
		return nil, err
	}

	return &models.LoginResponse{
		User:         *profile,
		Token:        accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// Login authenticates a user and returns tokens
func (s *AuthService) Login(req models.LoginRequest) (*models.LoginResponse, error) {
	// Get user by email
	user, err := s.authRepo.GetUserByEmail(req.Email)
	if err != nil {
		return nil, utils.ErrInvalidCredentials
	}

	// Check if user is active
	if !user.IsActive {
		return nil, utils.ErrUserInactive
	}

	// Verify password
	if !utils.CheckPassword(req.Password, user.PasswordHash) {
		return nil, utils.ErrInvalidCredentials
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

	// Get user profile
	profile, err := s.authRepo.GetUserWithProfile(user.ID)
	if err != nil {
		return nil, err
	}

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

	// Get user profile
	profile, err := s.authRepo.GetUserWithProfile(user.ID)
	if err != nil {
		return nil, err
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
	// Get user
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

	// Update password
	return s.authRepo.UpdatePassword(userID, req.NewPassword)
}

// GetUserProfile gets user profile by ID
func (s *AuthService) GetUserProfile(userID int) (*models.UserWithProfile, error) {
	return s.authRepo.GetUserWithProfile(userID)
}
