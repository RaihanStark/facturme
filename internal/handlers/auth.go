// Package handlers provides HTTP request handlers for the API endpoints.
// It includes authentication handlers for user registration and login.
package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"net/http"
	"time"

	"worklio-api/internal/db"
	"worklio-api/internal/email"
	"worklio-api/internal/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	queries      *db.Queries
	jwtSecret    string
	emailService *email.Service
}

func NewAuthHandler(queries *db.Queries, jwtSecret string, emailService *email.Service) *AuthHandler {
	return &AuthHandler{
		queries:      queries,
		jwtSecret:    jwtSecret,
		emailService: emailService,
	}
}

type Claims struct {
	UserID int32  `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// Register godoc
// @Summary Register a new user
// @Description Create a new user account with email, password, and name
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.RegisterRequest true "Register Request"
// @Success 201 {object} models.AuthResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 409 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/auth/register [post]
func (h *AuthHandler) Register(c echo.Context) error {
	var req models.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request body"})
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to hash password"})
	}

	// Generate verification token
	verificationToken, err := h.generateVerificationToken()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to generate verification token"})
	}

	// Create user with verification token (expires in 24 hours)
	user, err := h.queries.CreateUser(c.Request().Context(), db.CreateUserParams{
		Email:                    req.Email,
		PasswordHash:             sql.NullString{String: string(hashedPassword), Valid: true},
		Name:                     req.Name,
		VerificationToken:        sql.NullString{String: verificationToken, Valid: true},
		VerificationTokenExpires: sql.NullTime{Time: time.Now().Add(24 * time.Hour), Valid: true},
	})
	if err != nil {
		return c.JSON(http.StatusConflict, models.ErrorResponse{Error: "Email already exists"})
	}

	// Send verification email
	if h.emailService != nil {
		err = h.emailService.SendVerificationEmail(c.Request().Context(), user.Email, user.Name, verificationToken)
		if err != nil {
			c.Logger().Error("Failed to send verification email: ", err)
			// Don't fail the registration if email fails, just log it
		}
	} else {
		// Fallback: log token for testing when email service is not configured
		c.Logger().Info("Verification token for ", user.Email, ": ", verificationToken)
	}

	// Generate JWT token
	token, err := h.generateToken(user.ID, user.Email)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to generate token"})
	}

	return c.JSON(http.StatusCreated, models.AuthResponse{
		Token: token,
		User:  h.createUserToUserInfo(user),
	})
}

// Login godoc
// @Summary Login user
// @Description Authenticate user with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.LoginRequest true "Login Request"
// @Success 200 {object} models.AuthResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/auth/login [post]
func (h *AuthHandler) Login(c echo.Context) error {
	var req models.LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request body"})
	}

	// Get user by email
	user, err := h.queries.GetUserByEmail(c.Request().Context(), req.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Invalid credentials"})
		}
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to fetch user"})
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash.String), []byte(req.Password)); err != nil {
		return c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Invalid credentials"})
	}

	// Generate JWT token
	token, err := h.generateToken(user.ID, user.Email)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to generate token"})
	}

	return c.JSON(http.StatusOK, models.AuthResponse{
		Token: token,
		User:  h.getUserByEmailToUserInfo(user),
	})
}

func (h *AuthHandler) createUserToUserInfo(user db.CreateUserRow) models.UserInfo {
	return models.UserInfo{
		ID:                  user.ID,
		Email:               user.Email,
		Name:                user.Name,
		EmailVerified:       user.EmailVerified.Bool,
		OnboardingCompleted: user.OnboardingCompleted.Bool,
		TourCompleted:       user.TourCompleted.Bool,
		Currency:            user.Currency.String,
	}
}

func (h *AuthHandler) getUserByEmailToUserInfo(user db.GetUserByEmailRow) models.UserInfo {
	return models.UserInfo{
		ID:                  user.ID,
		Email:               user.Email,
		Name:                user.Name,
		EmailVerified:       user.EmailVerified.Bool,
		OnboardingCompleted: user.OnboardingCompleted.Bool,
		TourCompleted:       user.TourCompleted.Bool,
		Currency:            user.Currency.String,
	}
}

func (h *AuthHandler) completeOnboardingToUserInfo(user db.CompleteOnboardingRow) models.UserInfo {
	return models.UserInfo{
		ID:                  user.ID,
		Email:               user.Email,
		Name:                user.Name,
		EmailVerified:       user.EmailVerified.Bool,
		OnboardingCompleted: user.OnboardingCompleted.Bool,
		TourCompleted:       user.TourCompleted.Bool,
		Currency:            user.Currency.String,
	}
}

// CompleteOnboarding godoc
// @Summary Complete user onboarding
// @Description Update user settings after completing onboarding
// @Tags users
// @Accept json
// @Produce json
// @Param request body models.CompleteOnboardingRequest true "Onboarding Request"
// @Success 200 {object} models.UserInfo
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/users/complete-onboarding [post]
func (h *AuthHandler) CompleteOnboarding(c echo.Context) error {
	// Get user ID from context (set by auth middleware)
	userID, ok := c.Get("user_id").(int32)
	if !ok {
		return c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Unauthorized"})
	}

	var req models.CompleteOnboardingRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request body"})
	}

	// Update user with onboarding completion
	user, err := h.queries.CompleteOnboarding(c.Request().Context(), db.CompleteOnboardingParams{
		ID:         userID,
		Currency:   sql.NullString{String: req.Currency, Valid: true},
		DateFormat: sql.NullString{String: "MM/DD/YYYY", Valid: true},
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to complete onboarding"})
	}

	return c.JSON(http.StatusOK, h.completeOnboardingToUserInfo(user))
}

func (h *AuthHandler) generateToken(userID int32, email string) (string, error) {
	claims := &Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.jwtSecret))
}

func (h *AuthHandler) completeTourToUserInfo(user db.CompleteTourRow) models.UserInfo {
	return models.UserInfo{
		ID:                  user.ID,
		Email:               user.Email,
		Name:                user.Name,
		EmailVerified:       user.EmailVerified.Bool,
		OnboardingCompleted: user.OnboardingCompleted.Bool,
		TourCompleted:       user.TourCompleted.Bool,
		Currency:            user.Currency.String,
	}
}

// CompleteTour godoc
// @Summary Complete user tour
// @Description Mark the tour as completed for the user
// @Tags users
// @Produce json
// @Success 200 {object} models.UserInfo
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/users/complete-tour [post]
func (h *AuthHandler) CompleteTour(c echo.Context) error {
	// Get user ID from context (set by auth middleware)
	userID, ok := c.Get("user_id").(int32)
	if !ok {
		return c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Unauthorized"})
	}

	// Update user with tour completion
	user, err := h.queries.CompleteTour(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to complete tour"})
	}

	return c.JSON(http.StatusOK, h.completeTourToUserInfo(user))
}


func (h *AuthHandler) generateVerificationToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (h *AuthHandler) verifyUserEmailToUserInfo(user db.VerifyUserEmailRow) models.UserInfo {
	return models.UserInfo{
		ID:                  user.ID,
		Email:               user.Email,
		Name:                user.Name,
		EmailVerified:       user.EmailVerified.Bool,
		OnboardingCompleted: user.OnboardingCompleted.Bool,
		TourCompleted:       user.TourCompleted.Bool,
		Currency:            user.Currency.String,
	}
}

// VerifyEmail godoc
// @Summary Verify user email
// @Description Verify user's email address using verification token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.VerifyEmailRequest true "Verify Email Request"
// @Success 200 {object} models.UserInfo
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 410 {object} models.ErrorResponse
// @Router /api/auth/verify-email [post]
func (h *AuthHandler) VerifyEmail(c echo.Context) error {
	var req models.VerifyEmailRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request body"})
	}

	// Get user by verification token
	user, err := h.queries.GetUserByVerificationToken(c.Request().Context(), sql.NullString{
		String: req.Token,
		Valid:  true,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Invalid verification token"})
		}
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to verify email"})
	}

	// Check if token has expired
	if user.VerificationTokenExpires.Valid && user.VerificationTokenExpires.Time.Before(time.Now()) {
		return c.JSON(http.StatusGone, models.ErrorResponse{Error: "Verification token has expired"})
	}

	// Check if already verified
	if user.EmailVerified.Bool {
		return c.JSON(http.StatusOK, models.ErrorResponse{Error: "Email already verified"})
	}

	// Verify the email
	verifiedUser, err := h.queries.VerifyUserEmail(c.Request().Context(), user.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to verify email"})
	}

	return c.JSON(http.StatusOK, h.verifyUserEmailToUserInfo(verifiedUser))
}

// ResendVerificationEmail godoc
// @Summary Resend verification email
// @Description Send a new verification email to the user
// @Tags auth
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/auth/resend-verification [post]
func (h *AuthHandler) ResendVerificationEmail(c echo.Context) error {
	// Get user ID from context (set by auth middleware)
	userID, ok := c.Get("user_id").(int32)
	if !ok {
		return c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Unauthorized"})
	}

	// Get user details
	user, err := h.queries.GetUserByID(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to fetch user"})
	}

	// Check if already verified
	if user.EmailVerified.Bool {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Email is already verified"})
	}

	// Generate new verification token
	verificationToken, err := h.generateVerificationToken()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to generate verification token"})
	}

	// Update verification token in database
	_, err = h.queries.UpdateVerificationToken(c.Request().Context(), db.UpdateVerificationTokenParams{
		ID:                       userID,
		VerificationToken:        sql.NullString{String: verificationToken, Valid: true},
		VerificationTokenExpires: sql.NullTime{Time: time.Now().Add(24 * time.Hour), Valid: true},
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to update verification token"})
	}

	// Send verification email
	if h.emailService != nil {
		err = h.emailService.SendVerificationEmail(c.Request().Context(), user.Email, user.Name, verificationToken)
		if err != nil {
			c.Logger().Error("Failed to send verification email: ", err)
			return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to send verification email"})
		}
	} else {
		// Fallback: log token for testing when email service is not configured
		c.Logger().Info("Verification token for ", user.Email, ": ", verificationToken)
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Verification email sent successfully",
	})
}

// ForgotPassword godoc
// @Summary Request password reset
// @Description Send password reset email to user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body map[string]string true "Forgot Password Request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/auth/forgot-password [post]
func (h *AuthHandler) ForgotPassword(c echo.Context) error {
	var req struct {
		Email string `json:"email" validate:"required,email"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request body"})
	}

	// Get user by email
	user, err := h.queries.GetUserByEmail(c.Request().Context(), req.Email)
	if err != nil {
		// Don't reveal if email exists or not (security best practice)
		return c.JSON(http.StatusOK, map[string]string{
			"message": "If an account exists with this email, you will receive a password reset link shortly",
		})
	}

	// Generate password reset token
	resetToken, err := h.generateVerificationToken()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to generate reset token"})
	}

	// Update password reset token in database (expires in 1 hour)
	_, err = h.queries.UpdatePasswordResetToken(c.Request().Context(), db.UpdatePasswordResetTokenParams{
		Email:                      req.Email,
		PasswordResetToken:         sql.NullString{String: resetToken, Valid: true},
		PasswordResetTokenExpires:  sql.NullTime{Time: time.Now().Add(1 * time.Hour), Valid: true},
	})
	if err != nil {
		c.Logger().Error("Failed to update password reset token: ", err)
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to process password reset"})
	}

	// Send password reset email
	if h.emailService != nil {
		err = h.emailService.SendPasswordResetEmail(c.Request().Context(), user.Email, user.Name, resetToken)
		if err != nil {
			c.Logger().Error("Failed to send password reset email: ", err)
			// Don't fail the request if email fails
		}
	} else {
		// Fallback: log token for testing when email service is not configured
		c.Logger().Info("Password reset token for ", user.Email, ": ", resetToken)
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "If an account exists with this email, you will receive a password reset link shortly",
	})
}

// ResetPassword godoc
// @Summary Reset password with token
// @Description Reset user password using reset token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body map[string]string true "Reset Password Request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 410 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/auth/reset-password [post]
func (h *AuthHandler) ResetPassword(c echo.Context) error {
	var req struct {
		Token    string `json:"token" validate:"required"`
		Password string `json:"password" validate:"required,min=6"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request body"})
	}

	// Validate password length
	if len(req.Password) < 6 {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Password must be at least 6 characters"})
	}

	// Get user by password reset token
	user, err := h.queries.GetUserByPasswordResetToken(c.Request().Context(), sql.NullString{
		String: req.Token,
		Valid:  true,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Invalid or expired reset token"})
		}
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to reset password"})
	}

	// Check if token has expired
	if user.PasswordResetTokenExpires.Valid && user.PasswordResetTokenExpires.Time.Before(time.Now()) {
		return c.JSON(http.StatusGone, models.ErrorResponse{Error: "Reset token has expired"})
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to hash password"})
	}

	// Update password and clear reset token
	_, err = h.queries.ResetPassword(c.Request().Context(), db.ResetPasswordParams{
		ID:           user.ID,
		PasswordHash: sql.NullString{String: string(hashedPassword), Valid: true},
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to reset password"})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Password reset successfully",
	})
}

func (h *AuthHandler) changePasswordToUserInfo(user db.ChangePasswordRow) models.UserInfo {
	return models.UserInfo{
		ID:                  user.ID,
		Email:               user.Email,
		Name:                user.Name,
		EmailVerified:       user.EmailVerified.Bool,
		OnboardingCompleted: user.OnboardingCompleted.Bool,
		TourCompleted:       user.TourCompleted.Bool,
		Currency:            user.Currency.String,
	}
}

// ChangePassword godoc
// @Summary Change user password
// @Description Change user's password with current password verification
// @Tags users
// @Accept json
// @Produce json
// @Param request body map[string]string true "Change Password Request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/users/change-password [post]
func (h *AuthHandler) ChangePassword(c echo.Context) error {
	// Get user ID from context (set by auth middleware)
	userID, ok := c.Get("user_id").(int32)
	if !ok {
		return c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Unauthorized"})
	}

	var req struct {
		CurrentPassword string `json:"current_password" validate:"required"`
		NewPassword     string `json:"new_password" validate:"required,min=8"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request body"})
	}

	// Validate new password length
	if len(req.NewPassword) < 8 {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "New password must be at least 8 characters"})
	}

	// Get user by ID to verify current password
	user, err := h.queries.GetUserByID(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to fetch user"})
	}

	// Get the password hash from GetUserByEmail (which includes password_hash)
	userWithPassword, err := h.queries.GetUserByEmail(c.Request().Context(), user.Email)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to fetch user"})
	}

	// Check if user has a password (OAuth users can't change password)
	if !userWithPassword.PasswordHash.Valid {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Cannot change password for OAuth accounts"})
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(userWithPassword.PasswordHash.String), []byte(req.CurrentPassword)); err != nil {
		return c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Current password is incorrect"})
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to hash password"})
	}

	// Update password
	_, err = h.queries.ChangePassword(c.Request().Context(), db.ChangePasswordParams{
		ID:           userID,
		PasswordHash: sql.NullString{String: string(hashedPassword), Valid: true},
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to change password"})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Password changed successfully",
	})
}

// GetCurrentUser godoc
// @Summary Get current user info
// @Description Get the current authenticated user's information
// @Tags users
// @Produce json
// @Success 200 {object} models.UserInfo
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/users/me [get]
func (h *AuthHandler) GetCurrentUser(c echo.Context) error {
	// Get user ID from context (set by auth middleware)
	userID, ok := c.Get("user_id").(int32)
	if !ok {
		return c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Unauthorized"})
	}

	// Get user by ID
	user, err := h.queries.GetUserByID(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to fetch user"})
	}

	return c.JSON(http.StatusOK, h.getUserByIDToUserInfo(user))
}

// UpdateCurrency godoc
// @Summary Update user currency
// @Description Update the currency preference for the current user
// @Tags users
// @Accept json
// @Produce json
// @Param currency body models.UpdateCurrencyRequest true "Currency"
// @Success 200 {object} models.UserInfo
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/users/currency [post]
func (h *AuthHandler) UpdateCurrency(c echo.Context) error {
	// Get user ID from context
	userID, ok := c.Get("user_id").(int32)
	if !ok {
		return c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Unauthorized"})
	}

	// Parse request
	var req models.UpdateCurrencyRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request"})
	}

	// Validate currency
	if req.Currency == "" {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Currency is required"})
	}

	// Update currency
	user, err := h.queries.UpdateUserCurrency(c.Request().Context(), db.UpdateUserCurrencyParams{
		ID:       userID,
		Currency: sql.NullString{String: req.Currency, Valid: true},
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to update currency"})
	}

	return c.JSON(http.StatusOK, h.updateUserCurrencyToUserInfo(user))
}

func (h *AuthHandler) updateUserCurrencyToUserInfo(user db.UpdateUserCurrencyRow) models.UserInfo {
	return models.UserInfo{
		ID:                  user.ID,
		Email:               user.Email,
		Name:                user.Name,
		EmailVerified:       user.EmailVerified.Bool,
		OnboardingCompleted: user.OnboardingCompleted.Bool,
		TourCompleted:       user.TourCompleted.Bool,
		Currency:            user.Currency.String,
	}
}

func (h *AuthHandler) getUserByIDToUserInfo(user db.GetUserByIDRow) models.UserInfo {
	return models.UserInfo{
		ID:                  user.ID,
		Email:               user.Email,
		Name:                user.Name,
		EmailVerified:       user.EmailVerified.Bool,
		OnboardingCompleted: user.OnboardingCompleted.Bool,
		TourCompleted:       user.TourCompleted.Bool,
		Currency:            user.Currency.String,
	}
}
