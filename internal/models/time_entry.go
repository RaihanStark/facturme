package models

type CreateTimeEntryRequest struct {
	ClientID    int32   `json:"client_id" validate:"required"`
	Date        string  `json:"date" validate:"required"`
	Hours       float64 `json:"hours" validate:"required,gt=0"`
	Description string  `json:"description"`
}

type UpdateTimeEntryRequest struct {
	ClientID    int32   `json:"client_id" validate:"required"`
	Date        string  `json:"date" validate:"required"`
	Hours       float64 `json:"hours" validate:"required,gt=0"`
	Description string  `json:"description"`
}

type TimeEntryResponse struct {
	ID             int32   `json:"id"`
	UserID         int32   `json:"user_id"`
	ClientID       int32   `json:"client_id"`
	ClientName     string  `json:"client_name,omitempty"`
	ClientCurrency string  `json:"client_currency,omitempty"`
	Date           string  `json:"date"`
	Hours          float64 `json:"hours"`
	Description    string  `json:"description,omitempty"`
	HourlyRate     float64 `json:"hourly_rate"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
}

type HeatmapResponse struct {
	StartDate    string                        `json:"start_date"`
	EndDate      string                        `json:"end_date"`
	Data         map[string]float64            `json:"data"`
	Entries      map[string][]TimeEntryResponse `json:"entries"`
	TotalHours   float64                       `json:"total_hours"`
	DaysWorked   int                           `json:"days_worked"`
	DaysOff      int                           `json:"days_off"`
	AverageHours float64                       `json:"average_hours"`
}

type TimeEntriesWithStatsResponse struct {
	Entries      []TimeEntryResponse `json:"entries"`
	TotalHours   float64            `json:"total_hours"`
	TotalRevenue float64            `json:"total_revenue"`
}
