package models

type CreateInvoiceRequest struct {
	ClientID      int32   `json:"client_id" validate:"required"`
	InvoiceNumber string  `json:"invoice_number" validate:"required"`
	IssueDate     string  `json:"issue_date" validate:"required"`
	DueDate       string  `json:"due_date" validate:"required"`
	Status        string  `json:"status" validate:"required,oneof=draft sent paid overdue"`
	Notes         string  `json:"notes"`
	TimeEntryIDs  []int32 `json:"time_entry_ids" validate:"required,min=1"`
}

type UpdateInvoiceRequest struct {
	ClientID      int32  `json:"client_id" validate:"required"`
	InvoiceNumber string `json:"invoice_number" validate:"required"`
	IssueDate     string `json:"issue_date" validate:"required"`
	DueDate       string `json:"due_date" validate:"required"`
	Status        string `json:"status" validate:"required,oneof=draft sent paid overdue"`
	Notes         string `json:"notes"`
}

type UpdateInvoiceStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=draft sent paid overdue"`
}

type InvoiceResponse struct {
	ID             int32               `json:"id"`
	UserID         int32               `json:"user_id"`
	ClientID       int32               `json:"client_id"`
	ClientName     string              `json:"client_name,omitempty"`
	ClientCurrency string              `json:"client_currency,omitempty"`
	InvoiceNumber  string              `json:"invoice_number"`
	IssueDate      string              `json:"issue_date"`
	DueDate        string              `json:"due_date"`
	Status         string              `json:"status"`
	Notes          string              `json:"notes,omitempty"`
	TimeEntries    []TimeEntryResponse `json:"time_entries"`
	TotalHours     float64             `json:"total_hours"`
	TotalAmount    float64             `json:"total_amount"`
	CreatedAt      string              `json:"created_at"`
	UpdatedAt      string              `json:"updated_at"`
}
