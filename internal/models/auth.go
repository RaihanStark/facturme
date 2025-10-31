// Package models defines the data structures for API requests and responses.
// It includes models for authentication, user information, and error handling.
package models

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Name     string `json:"name" validate:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type AuthResponse struct {
	Token string    `json:"token"`
	User  UserInfo  `json:"user"`
}

type UserInfo struct {
	ID                  int32  `json:"id"`
	Email               string `json:"email"`
	Name                string `json:"name"`
	EmailVerified       bool   `json:"email_verified"`
	OnboardingCompleted bool   `json:"onboarding_completed"`
	TourCompleted       bool   `json:"tour_completed"`
	Currency            string `json:"currency"`
}

type CompleteOnboardingRequest struct {
	Currency string `json:"currency" validate:"required"`
}

type VerifyEmailRequest struct {
	Token string `json:"token" validate:"required"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type UpdateCurrencyRequest struct {
	Currency string `json:"currency" validate:"required"`
}
