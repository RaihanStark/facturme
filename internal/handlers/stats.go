package handlers

import (
	"net/http"
	"strconv"
	"time"
	"worklio-api/internal/db"
	"worklio-api/internal/models"
	"worklio-api/internal/services"

	"github.com/labstack/echo/v4"
)

type StatsHandler struct {
	queries         *db.Queries
	exchangeService *services.ExchangeRateService
}

func NewStatsHandler(queries *db.Queries, exchangeService *services.ExchangeRateService) *StatsHandler {
	return &StatsHandler{
		queries:         queries,
		exchangeService: exchangeService,
	}
}

// DashboardStatsResponse represents the response for dashboard stats
type DashboardStatsResponse struct {
	TotalHours      float64 `json:"total_hours"`
	TotalRevenue    float64 `json:"total_revenue"`
	UnpaidInvoices  float64 `json:"unpaid_invoices"`
	PaidInvoices    float64 `json:"paid_invoices"`
}

// GetDashboardStats godoc
// @Summary Get dashboard statistics
// @Description Get calculated dashboard statistics with currency conversion
// @Tags stats
// @Produce json
// @Security BearerAuth
// @Param from query string false "Start date (YYYY-MM-DD format)"
// @Param to query string false "End date (YYYY-MM-DD format)"
// @Success 200 {object} DashboardStatsResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/stats/dashboard [get]
func (h *StatsHandler) GetDashboardStats(c echo.Context) error {
	userID := c.Get("user_id").(int32)

	// Get user's currency preference
	user, err := h.queries.GetUserByID(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to get user info"})
	}

	userCurrency := "USD"
	if user.Currency.Valid {
		userCurrency = user.Currency.String
	}

	// Parse date range filters
	var fromDate, toDate *time.Time
	if fromStr := c.QueryParam("from"); fromStr != "" {
		parsed, err := time.Parse("2006-01-02", fromStr)
		if err != nil {
			return c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid from date format. Use YYYY-MM-DD"})
		}
		fromDate = &parsed
	}
	if toStr := c.QueryParam("to"); toStr != "" {
		parsed, err := time.Parse("2006-01-02", toStr)
		if err != nil {
			return c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid to date format. Use YYYY-MM-DD"})
		}
		// Set to end of day
		endOfDay := time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 23, 59, 59, 999999999, parsed.Location())
		toDate = &endOfDay
	}

	// Get time entries
	timeEntries, err := h.queries.GetTimeEntriesByUserID(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to get time entries"})
	}

	// Get clients for currency conversion
	clients, err := h.queries.GetClientsByUserID(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to get clients"})
	}

	// Create clients map
	clientsMap := make(map[int32]db.GetClientsByUserIDRow)
	for _, client := range clients {
		clientsMap[client.ID] = client
	}

	// Get unique currencies needed for conversion
	currenciesNeeded := make(map[string]bool)
	for _, client := range clients {
		if client.Currency != userCurrency {
			currenciesNeeded[client.Currency] = true
		}
	}

	// Fetch conversion rates
	conversionRates := make(map[string]float64)
	for currency := range currenciesNeeded {
		convertedAmount, err := h.exchangeService.ConvertAmount(c.Request().Context(), 1.0, currency, userCurrency)
		if err != nil {
			// Fallback to 1:1 if conversion fails
			conversionRates[currency] = 1.0
		} else {
			conversionRates[currency] = convertedAmount
		}
	}

	// Calculate total hours and revenue
	var totalHours float64
	var totalRevenue float64

	for _, entry := range timeEntries {
		// Apply date filter
		if fromDate != nil && entry.Date.Before(*fromDate) {
			continue
		}
		if toDate != nil && entry.Date.After(*toDate) {
			continue
		}

		hours, _ := strconv.ParseFloat(entry.Hours, 64)
		totalHours += hours

		// Calculate revenue with currency conversion
		if client, ok := clientsMap[entry.ClientID]; ok {
			hourlyRate, _ := strconv.ParseFloat(client.HourlyRate.String, 64)
			entryAmount := hours * hourlyRate
			clientCurrency := client.Currency

			if clientCurrency != userCurrency {
				if rate, ok := conversionRates[clientCurrency]; ok {
					totalRevenue += entryAmount * rate
				} else {
					totalRevenue += entryAmount
				}
			} else {
				totalRevenue += entryAmount
			}
		}
	}

	// Get invoices
	invoices, err := h.queries.GetInvoicesByUserID(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to get invoices"})
	}

	// Calculate unpaid and paid invoices
	var unpaidInvoices float64
	var paidInvoices float64

	for _, invoice := range invoices {
		// Apply date filter
		if fromDate != nil && invoice.IssueDate.Before(*fromDate) {
			continue
		}
		if toDate != nil && invoice.IssueDate.After(*toDate) {
			continue
		}

		// Get time entries for this invoice to calculate total
		invoiceTimeEntries, err := h.queries.GetInvoiceTimeEntries(c.Request().Context(), invoice.ID)
		if err != nil {
			continue
		}

		var invoiceTotal float64
		for _, entry := range invoiceTimeEntries {
			hours, _ := strconv.ParseFloat(entry.Hours, 64)
			hourlyRate, _ := strconv.ParseFloat(entry.HourlyRate.String, 64)
			invoiceTotal += hours * hourlyRate
		}

		// Get client for currency conversion
		client, ok := clientsMap[invoice.ClientID]
		if !ok {
			continue
		}

		clientCurrency := client.Currency
		convertedAmount := invoiceTotal
		if clientCurrency != userCurrency {
			if rate, ok := conversionRates[clientCurrency]; ok {
				convertedAmount = invoiceTotal * rate
			}
		}

		if invoice.Status == "sent" || invoice.Status == "overdue" {
			unpaidInvoices += convertedAmount
		} else if invoice.Status == "paid" {
			paidInvoices += convertedAmount
		}
	}

	return c.JSON(http.StatusOK, DashboardStatsResponse{
		TotalHours:     totalHours,
		TotalRevenue:   totalRevenue,
		UnpaidInvoices: unpaidInvoices,
		PaidInvoices:   paidInvoices,
	})
}

// RecentTimeEntryResponse represents a time entry with client information
type RecentTimeEntryResponse struct {
	ID          int32   `json:"id"`
	UserID      int32   `json:"user_id"`
	ClientID    int32   `json:"client_id"`
	ClientName  string  `json:"client_name"`
	Date        string  `json:"date"`
	Hours       float64 `json:"hours"`
	Description string  `json:"description"`
	HourlyRate  float64 `json:"hourly_rate"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

// GetRecentTimeEntries godoc
// @Summary Get recent time entries
// @Description Get recent time entries filtered by date range, sorted by date descending
// @Tags stats
// @Produce json
// @Security BearerAuth
// @Param from query string false "Start date (YYYY-MM-DD format)"
// @Param to query string false "End date (YYYY-MM-DD format)"
// @Param limit query int false "Number of entries to return (default: 5)"
// @Success 200 {array} RecentTimeEntryResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/stats/recent-time-entries [get]
func (h *StatsHandler) GetRecentTimeEntries(c echo.Context) error {
	userID := c.Get("user_id").(int32)

	// Parse date range filters
	var fromDate, toDate *time.Time
	if fromStr := c.QueryParam("from"); fromStr != "" {
		parsed, err := time.Parse("2006-01-02", fromStr)
		if err != nil {
			return c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid from date format. Use YYYY-MM-DD"})
		}
		fromDate = &parsed
	}
	if toStr := c.QueryParam("to"); toStr != "" {
		parsed, err := time.Parse("2006-01-02", toStr)
		if err != nil {
			return c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid to date format. Use YYYY-MM-DD"})
		}
		// Set to end of day
		endOfDay := time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 23, 59, 59, 999999999, parsed.Location())
		toDate = &endOfDay
	}

	// Parse limit
	limit := 5
	if limitStr := c.QueryParam("limit"); limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	// Get time entries
	timeEntries, err := h.queries.GetTimeEntriesByUserID(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to get time entries"})
	}

	// Get clients
	clients, err := h.queries.GetClientsByUserID(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to get clients"})
	}

	// Create clients map
	clientsMap := make(map[int32]string)
	for _, client := range clients {
		clientsMap[client.ID] = client.Name
	}

	// Filter and sort time entries
	var filtered []db.GetTimeEntriesByUserIDRow
	for _, entry := range timeEntries {
		// Apply date filter
		if fromDate != nil && entry.Date.Before(*fromDate) {
			continue
		}
		if toDate != nil && entry.Date.After(*toDate) {
			continue
		}
		filtered = append(filtered, entry)
	}

	// Sort by date descending
	for i := 0; i < len(filtered); i++ {
		for j := i + 1; j < len(filtered); j++ {
			if filtered[j].Date.After(filtered[i].Date) {
				filtered[i], filtered[j] = filtered[j], filtered[i]
			}
		}
	}

	// Limit results
	if len(filtered) > limit {
		filtered = filtered[:limit]
	}

	// Convert to response format
	response := make([]RecentTimeEntryResponse, len(filtered))
	for i, entry := range filtered {
		hours, _ := strconv.ParseFloat(entry.Hours, 64)
		hourlyRate, _ := strconv.ParseFloat(entry.HourlyRate.String, 64)
		clientName := "Unknown"
		if name, ok := clientsMap[entry.ClientID]; ok {
			clientName = name
		}
		response[i] = RecentTimeEntryResponse{
			ID:          entry.ID,
			UserID:      entry.UserID,
			ClientID:    entry.ClientID,
			ClientName:  clientName,
			Date:        entry.Date.Format("2006-01-02"),
			Hours:       hours,
			Description: entry.Description.String,
			HourlyRate:  hourlyRate,
			CreatedAt:   entry.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:   entry.UpdatedAt.Time.Format("2006-01-02T15:04:05Z"),
		}
	}

	return c.JSON(http.StatusOK, response)
}

// RecentInvoiceResponse represents an invoice with client information
type RecentInvoiceResponse struct {
	ID             int32                    `json:"id"`
	UserID         int32                    `json:"user_id"`
	ClientID       int32                    `json:"client_id"`
	ClientName     string                   `json:"client_name"`
	ClientCurrency string                   `json:"client_currency"`
	InvoiceNumber  string                   `json:"invoice_number"`
	IssueDate      string                   `json:"issue_date"`
	DueDate        string                   `json:"due_date"`
	Status         string                   `json:"status"`
	Notes          string                   `json:"notes"`
	TimeEntries    []models.TimeEntryResponse `json:"time_entries"`
	TotalHours     float64                  `json:"total_hours"`
	TotalAmount    float64                  `json:"total_amount"`
	CreatedAt      string                   `json:"created_at"`
	UpdatedAt      string                   `json:"updated_at"`
}

// GetRecentInvoices godoc
// @Summary Get recent invoices
// @Description Get recent invoices filtered by date range, sorted by issue date descending
// @Tags stats
// @Produce json
// @Security BearerAuth
// @Param from query string false "Start date (YYYY-MM-DD format)"
// @Param to query string false "End date (YYYY-MM-DD format)"
// @Param limit query int false "Number of invoices to return (default: 5)"
// @Success 200 {array} RecentInvoiceResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/stats/recent-invoices [get]
func (h *StatsHandler) GetRecentInvoices(c echo.Context) error {
	userID := c.Get("user_id").(int32)

	// Parse date range filters
	var fromDate, toDate *time.Time
	if fromStr := c.QueryParam("from"); fromStr != "" {
		parsed, err := time.Parse("2006-01-02", fromStr)
		if err != nil {
			return c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid from date format. Use YYYY-MM-DD"})
		}
		fromDate = &parsed
	}
	if toStr := c.QueryParam("to"); toStr != "" {
		parsed, err := time.Parse("2006-01-02", toStr)
		if err != nil {
			return c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid to date format. Use YYYY-MM-DD"})
		}
		// Set to end of day
		endOfDay := time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 23, 59, 59, 999999999, parsed.Location())
		toDate = &endOfDay
	}

	// Parse limit
	limit := 5
	if limitStr := c.QueryParam("limit"); limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	// Get invoices
	invoices, err := h.queries.GetInvoicesByUserID(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to get invoices"})
	}

	// Get clients
	clients, err := h.queries.GetClientsByUserID(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to get clients"})
	}

	// Create clients map with name and currency
	type ClientInfo struct {
		Name     string
		Currency string
	}
	clientsMap := make(map[int32]ClientInfo)
	for _, client := range clients {
		clientsMap[client.ID] = ClientInfo{
			Name:     client.Name,
			Currency: client.Currency,
		}
	}

	// Filter and sort invoices
	var filtered []db.Invoice
	for _, invoice := range invoices {
		// Apply date filter
		if fromDate != nil && invoice.IssueDate.Before(*fromDate) {
			continue
		}
		if toDate != nil && invoice.IssueDate.After(*toDate) {
			continue
		}
		filtered = append(filtered, invoice)
	}

	// Sort by issue date descending
	for i := 0; i < len(filtered); i++ {
		for j := i + 1; j < len(filtered); j++ {
			if filtered[j].IssueDate.After(filtered[i].IssueDate) {
				filtered[i], filtered[j] = filtered[j], filtered[i]
			}
		}
	}

	// Limit results
	if len(filtered) > limit {
		filtered = filtered[:limit]
	}

	// Convert to response format
	response := make([]RecentInvoiceResponse, len(filtered))
	for i, invoice := range filtered {
		// Get time entries for this invoice
		timeEntries, err := h.queries.GetInvoiceTimeEntries(c.Request().Context(), invoice.ID)
		if err != nil {
			continue
		}

		timeEntryResponses := make([]models.TimeEntryResponse, len(timeEntries))
		totalHours := 0.0
		totalAmount := 0.0

		for j, entry := range timeEntries {
			hours, _ := strconv.ParseFloat(entry.Hours, 64)
			hourlyRate, _ := strconv.ParseFloat(entry.HourlyRate.String, 64)
			totalHours += hours
			totalAmount += hours * hourlyRate

			timeEntryResponses[j] = models.TimeEntryResponse{
				ID:          entry.ID,
				UserID:      entry.UserID,
				ClientID:    entry.ClientID,
				Date:        entry.Date.Format("2006-01-02"),
				Hours:       hours,
				Description: entry.Description.String,
				HourlyRate:  hourlyRate,
				CreatedAt:   entry.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
				UpdatedAt:   entry.UpdatedAt.Time.Format("2006-01-02T15:04:05Z"),
			}
		}

		// Get client info
		clientName := "Unknown"
		clientCurrency := "USD"
		if clientInfo, ok := clientsMap[invoice.ClientID]; ok {
			clientName = clientInfo.Name
			clientCurrency = clientInfo.Currency
		}

		response[i] = RecentInvoiceResponse{
			ID:             invoice.ID,
			UserID:         invoice.UserID,
			ClientID:       invoice.ClientID,
			ClientName:     clientName,
			ClientCurrency: clientCurrency,
			InvoiceNumber:  invoice.InvoiceNumber,
			IssueDate:      invoice.IssueDate.Format("2006-01-02"),
			DueDate:        invoice.DueDate.Format("2006-01-02"),
			Status:         invoice.Status,
			Notes:          invoice.Notes.String,
			TimeEntries:    timeEntryResponses,
			TotalHours:     totalHours,
			TotalAmount:    totalAmount,
			CreatedAt:      invoice.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:      invoice.UpdatedAt.Time.Format("2006-01-02T15:04:05Z"),
		}
	}

	return c.JSON(http.StatusOK, response)
}

// InvoiceStatsResponse represents the response for invoice stats
type InvoiceStatsResponse struct {
	Invoices            []models.InvoiceResponse `json:"invoices"`
	TotalInvoices       int                      `json:"total_invoices"`
	TotalAmount         float64                  `json:"total_amount"`
	PaidAmount          float64                  `json:"paid_amount"`
	UnpaidAmount        float64                  `json:"unpaid_amount"`
}

// GetInvoiceStats godoc
// @Summary Get invoice statistics
// @Description Get all invoices with calculated totals in user's currency
// @Tags stats
// @Produce json
// @Security BearerAuth
// @Param status query string false "Filter by status (all, draft, sent, paid, overdue)" default(all)
// @Success 200 {object} InvoiceStatsResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/stats/invoices [get]
func (h *StatsHandler) GetInvoiceStats(c echo.Context) error {
	userID := c.Get("user_id").(int32)
	statusFilter := c.QueryParam("status")
	if statusFilter == "" {
		statusFilter = "all"
	}

	// Get user's currency preference
	user, err := h.queries.GetUserByID(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to get user info"})
	}

	userCurrency := "USD"
	if user.Currency.Valid {
		userCurrency = user.Currency.String
	}

	// Get all invoices
	invoices, err := h.queries.GetInvoicesByUserID(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to get invoices"})
	}

	// Get clients for currency conversion
	clients, err := h.queries.GetClientsByUserID(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to get clients"})
	}

	// Create clients map
	clientsMap := make(map[int32]db.GetClientsByUserIDRow)
	for _, client := range clients {
		clientsMap[client.ID] = client
	}

	// Get unique currencies needed for conversion
	currenciesNeeded := make(map[string]bool)
	for _, client := range clients {
		if client.Currency != userCurrency {
			currenciesNeeded[client.Currency] = true
		}
	}

	// Fetch conversion rates
	conversionRates := make(map[string]float64)
	for currency := range currenciesNeeded {
		convertedAmount, err := h.exchangeService.ConvertAmount(c.Request().Context(), 1.0, currency, userCurrency)
		if err != nil {
			// Log the error but continue with fallback
			c.Logger().Errorf("Failed to convert %s to %s: %v", currency, userCurrency, err)
			conversionRates[currency] = 1.0
		} else {
			c.Logger().Infof("Loaded conversion rate: 1.0 %s = %f %s", currency, convertedAmount, userCurrency)
			conversionRates[currency] = convertedAmount
		}
	}

	var totalAmount, paidAmount, unpaidAmount float64
	invoiceResponses := make([]models.InvoiceResponse, 0)

	for _, invoice := range invoices {
		// Get time entries for this invoice
		timeEntries, err := h.queries.GetInvoiceTimeEntries(c.Request().Context(), invoice.ID)
		if err != nil {
			continue
		}

		// Convert time entries to response format
		timeEntryResponses := make([]models.TimeEntryResponse, 0)
		var totalHours float64
		var invoiceTotal float64

		for _, entry := range timeEntries {
			hours, _ := strconv.ParseFloat(entry.Hours, 64)
			hourlyRate, _ := strconv.ParseFloat(entry.HourlyRate.String, 64)
			totalHours += hours
			invoiceTotal += hours * hourlyRate

			timeEntryResponses = append(timeEntryResponses, models.TimeEntryResponse{
				ID:          entry.ID,
				UserID:      entry.UserID,
				ClientID:    entry.ClientID,
				Date:        entry.Date.Format("2006-01-02"),
				Hours:       hours,
				Description: entry.Description.String,
				HourlyRate:  hourlyRate,
				CreatedAt:   entry.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
				UpdatedAt:   entry.UpdatedAt.Time.Format("2006-01-02T15:04:05Z"),
			})
		}

		// Get client info
		clientName := "Unknown"
		clientCurrency := "USD"
		if client, ok := clientsMap[invoice.ClientID]; ok {
			clientName = client.Name
			clientCurrency = client.Currency
		}

		// Convert to user currency for totals
		convertedAmount := invoiceTotal
		if clientCurrency != userCurrency {
			if rate, ok := conversionRates[clientCurrency]; ok {
				convertedAmount = invoiceTotal * rate
				c.Logger().Infof("Invoice %d: Converting %f %s to %f %s (rate: %f)", invoice.ID, invoiceTotal, clientCurrency, convertedAmount, userCurrency, rate)
			} else {
				c.Logger().Warnf("Invoice %d: No rate found for %s to %s (have %d rates)", invoice.ID, clientCurrency, userCurrency, len(conversionRates))
			}
		} else {
			c.Logger().Infof("Invoice %d: Same currency %s, no conversion needed", invoice.ID, clientCurrency)
		}

		// Always calculate totals for ALL invoices (regardless of filter)
		totalAmount += convertedAmount
		if invoice.Status == "paid" {
			paidAmount += convertedAmount
		} else if invoice.Status == "sent" || invoice.Status == "overdue" {
			unpaidAmount += convertedAmount
		}

		// Filter by status for the invoice list only
		if statusFilter != "all" && invoice.Status != statusFilter {
			continue
		}

		invoiceResponse := models.InvoiceResponse{
			ID:             invoice.ID,
			UserID:         invoice.UserID,
			ClientID:       invoice.ClientID,
			ClientName:     clientName,
			ClientCurrency: clientCurrency,
			InvoiceNumber:  invoice.InvoiceNumber,
			IssueDate:      invoice.IssueDate.Format("2006-01-02"),
			DueDate:        invoice.DueDate.Format("2006-01-02"),
			Status:         invoice.Status,
			Notes:          invoice.Notes.String,
			TimeEntries:    timeEntryResponses,
			TotalHours:     totalHours,
			TotalAmount:    invoiceTotal,
			CreatedAt:      invoice.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:      invoice.UpdatedAt.Time.Format("2006-01-02T15:04:05Z"),
		}

		invoiceResponses = append(invoiceResponses, invoiceResponse)
	}

	response := InvoiceStatsResponse{
		Invoices:      invoiceResponses,
		TotalInvoices: len(invoiceResponses),
		TotalAmount:   totalAmount,
		PaidAmount:    paidAmount,
		UnpaidAmount:  unpaidAmount,
	}

	return c.JSON(http.StatusOK, response)
}
