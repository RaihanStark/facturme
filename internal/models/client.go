package models

type CreateClientRequest struct {
	Name       string  `json:"name" validate:"required"`
	Email      string  `json:"email" validate:"required,email"`
	Phone      string  `json:"phone"`
	Company    string  `json:"company"`
	Address    string  `json:"address"`
	HourlyRate float64 `json:"hourly_rate"`
	Currency   string  `json:"currency" validate:"required"`
}

type UpdateClientRequest struct {
	Name       string  `json:"name" validate:"required"`
	Email      string  `json:"email" validate:"required,email"`
	Phone      string  `json:"phone"`
	Company    string  `json:"company"`
	Address    string  `json:"address"`
	HourlyRate float64 `json:"hourly_rate"`
	Currency   string  `json:"currency" validate:"required"`
}

type ClientResponse struct {
	ID         int32   `json:"id"`
	UserID     int32   `json:"user_id"`
	Name       string  `json:"name"`
	Email      string  `json:"email"`
	Phone      string  `json:"phone,omitempty"`
	Company    string  `json:"company,omitempty"`
	Address    string  `json:"address,omitempty"`
	HourlyRate float64 `json:"hourly_rate"`
	Currency   string  `json:"currency"`
	CreatedAt  string  `json:"created_at"`
	UpdatedAt  string  `json:"updated_at"`
}
