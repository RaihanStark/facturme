package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"worklio-api/internal/db"
	"worklio-api/internal/models"
	"worklio-api/internal/services"

	"github.com/labstack/echo/v4"
)

type TimeEntryHandler struct {
	queries         *db.Queries
	exchangeService *services.ExchangeRateService
}

func NewTimeEntryHandler(queries *db.Queries, exchangeService *services.ExchangeRateService) *TimeEntryHandler {
	return &TimeEntryHandler{
		queries:         queries,
		exchangeService: exchangeService,
	}
}

// CreateTimeEntry godoc
// @Summary Create a new time entry
// @Description Create a new time entry for the authenticated user
// @Tags time-entries
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.CreateTimeEntryRequest true "Create Time Entry Request"
// @Success 201 {object} models.TimeEntryResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/time-entries [post]
func (h *TimeEntryHandler) CreateTimeEntry(c echo.Context) error {
	userID := c.Get("user_id").(int32)

	var req models.CreateTimeEntryRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request body"})
	}

	// Parse date in local timezone
	date, err := time.ParseInLocation("2006-01-02", req.Date, time.Local)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid date format. Use YYYY-MM-DD"})
	}

	// Fetch client to get their current hourly rate
	client, err := h.queries.GetClientByID(c.Request().Context(), db.GetClientByIDParams{
		ID:     req.ClientID,
		UserID: userID,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Client not found"})
		}
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to fetch client"})
	}

	timeEntry, err := h.queries.CreateTimeEntry(c.Request().Context(), db.CreateTimeEntryParams{
		UserID:      userID,
		ClientID:    req.ClientID,
		Date:        date,
		Hours:       fmt.Sprintf("%.2f", req.Hours),
		Description: sql.NullString{String: req.Description, Valid: req.Description != ""},
		HourlyRate:  client.HourlyRate,
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to create time entry"})
	}

	return c.JSON(http.StatusCreated, createTimeEntryRowToResponse(timeEntry))
}

// GetTimeEntries godoc
// @Summary Get time entries with optional filtering
// @Description Get time entries for the authenticated user. Supports filtering by view_mode (daily/weekly/monthly) and date.
// @Tags time-entries
// @Produce json
// @Security BearerAuth
// @Param view_mode query string false "View mode: daily, weekly, or monthly"
// @Param date query string false "Date in YYYY-MM-DD format (required if view_mode is set)"
// @Success 200 {array} models.TimeEntryResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/time-entries [get]
func (h *TimeEntryHandler) GetTimeEntries(c echo.Context) error {
	userID := c.Get("user_id").(int32)

	viewMode := c.QueryParam("view_mode")
	dateStr := c.QueryParam("date")

	// Get all time entries first
	timeEntries, err := h.queries.GetTimeEntriesByUserID(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to fetch time entries"})
	}

	// If no filtering, return all entries
	if viewMode == "" || dateStr == "" {
		response := make([]models.TimeEntryResponse, len(timeEntries))
		for i, entry := range timeEntries {
			response[i] = getTimeEntriesByUserIDRowToResponse(entry)
		}
		return c.JSON(http.StatusOK, response)
	}

	// Apply filtering
	filtered, err := h.filterTimeEntries(timeEntries, viewMode, dateStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
	}

	// Get clients to add client names
	clients, err := h.queries.GetClientsByUserID(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to get clients"})
	}

	clientsMap := make(map[int32]string)
	for _, client := range clients {
		clientsMap[client.ID] = client.Name
	}

	response := make([]models.TimeEntryResponse, len(filtered))
	for i, entry := range filtered {
		response[i] = getTimeEntriesByUserIDRowToResponse(entry)
		if clientName, ok := clientsMap[entry.ClientID]; ok {
			response[i].ClientName = clientName
		}
	}

	return c.JSON(http.StatusOK, response)
}

// GetTimeEntriesStats godoc
// @Summary Get time entries statistics
// @Description Get statistics for time entries filtered by view_mode and date
// @Tags time-entries
// @Produce json
// @Security BearerAuth
// @Param view_mode query string true "View mode: daily, weekly, or monthly"
// @Param date query string true "Date in YYYY-MM-DD format"
// @Success 200 {object} models.TimeEntriesWithStatsResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/time-entries/stats [get]
func (h *TimeEntryHandler) GetTimeEntriesStats(c echo.Context) error {
	userID := c.Get("user_id").(int32)

	viewMode := c.QueryParam("view_mode")
	dateStr := c.QueryParam("date")

	if viewMode == "" || dateStr == "" {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "view_mode and date parameters are required"})
	}

	return h.getFilteredTimeEntriesWithStats(c, userID, viewMode, dateStr)
}

func (h *TimeEntryHandler) filterTimeEntries(timeEntries []db.GetTimeEntriesByUserIDRow, viewMode string, dateStr string) ([]db.GetTimeEntriesByUserIDRow, error) {
	// Validate view mode
	if viewMode != "daily" && viewMode != "weekly" && viewMode != "monthly" {
		return nil, fmt.Errorf("view_mode must be daily, weekly, or monthly")
	}

	// Parse date
	currentDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, fmt.Errorf("invalid date format. Use YYYY-MM-DD")
	}

	// Calculate date range
	var startDate, endDate time.Time

	if viewMode == "daily" {
		startDate = time.Date(currentDate.Year(), currentDate.Month(), currentDate.Day(), 0, 0, 0, 0, currentDate.Location())
		endDate = time.Date(currentDate.Year(), currentDate.Month(), currentDate.Day(), 23, 59, 59, 999999999, currentDate.Location())
	} else if viewMode == "weekly" {
		day := int(currentDate.Weekday())
		daysToSubtract := day - 1
		if day == 0 {
			daysToSubtract = 6
		}

		startDate = currentDate.AddDate(0, 0, -daysToSubtract)
		startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())

		endDate = startDate.AddDate(0, 0, 6)
		endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 999999999, endDate.Location())
	} else {
		startDate = time.Date(currentDate.Year(), currentDate.Month(), 1, 0, 0, 0, 0, currentDate.Location())
		endDate = startDate.AddDate(0, 1, 0).Add(-time.Nanosecond)
	}

	// Filter entries
	var filtered []db.GetTimeEntriesByUserIDRow
	for _, entry := range timeEntries {
		if !entry.Date.Before(startDate) && !entry.Date.After(endDate) {
			filtered = append(filtered, entry)
		}
	}

	return filtered, nil
}

func (h *TimeEntryHandler) getFilteredTimeEntriesWithStats(c echo.Context, userID int32, viewMode string, dateStr string) error {
	// Validate view mode
	if viewMode != "daily" && viewMode != "weekly" && viewMode != "monthly" {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "view_mode must be daily, weekly, or monthly"})
	}

	// Parse date
	currentDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid date format. Use YYYY-MM-DD"})
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

	// Calculate date range based on view mode
	var startDate, endDate time.Time

	if viewMode == "daily" {
		startDate = time.Date(currentDate.Year(), currentDate.Month(), currentDate.Day(), 0, 0, 0, 0, currentDate.Location())
		endDate = time.Date(currentDate.Year(), currentDate.Month(), currentDate.Day(), 23, 59, 59, 999999999, currentDate.Location())
	} else if viewMode == "weekly" {
		// Start week on Monday
		day := int(currentDate.Weekday())
		daysToSubtract := day - 1
		if day == 0 { // Sunday
			daysToSubtract = 6
		}

		startDate = currentDate.AddDate(0, 0, -daysToSubtract)
		startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())

		endDate = startDate.AddDate(0, 0, 6)
		endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 999999999, endDate.Location())
	} else {
		// Monthly
		startDate = time.Date(currentDate.Year(), currentDate.Month(), 1, 0, 0, 0, 0, currentDate.Location())
		endDate = startDate.AddDate(0, 1, 0).Add(-time.Nanosecond)
	}

	// Get all time entries
	timeEntries, err := h.queries.GetTimeEntriesByUserID(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to get time entries"})
	}

	// Get clients for currency conversion and names
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

	// Filter entries by date range and calculate stats
	var filteredEntries []models.TimeEntryResponse
	var totalHours float64
	var totalRevenue float64

	for _, entry := range timeEntries {
		// Apply date filter
		if entry.Date.Before(startDate) || entry.Date.After(endDate) {
			continue
		}

		hours, _ := strconv.ParseFloat(entry.Hours, 64)
		totalHours += hours

		// Get client info
		clientName := "Unknown"
		hourlyRate := 0.0
		clientCurrency := userCurrency

		if client, ok := clientsMap[entry.ClientID]; ok {
			clientName = client.Name
			hourlyRate, _ = strconv.ParseFloat(client.HourlyRate.String, 64)
			clientCurrency = client.Currency

			// Calculate revenue with currency conversion
			entryAmount := hours * hourlyRate
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

		entryResponse := getTimeEntriesByUserIDRowToResponse(entry)
		entryResponse.ClientName = clientName
		entryResponse.ClientCurrency = clientCurrency
		filteredEntries = append(filteredEntries, entryResponse)
	}

	return c.JSON(http.StatusOK, models.TimeEntriesWithStatsResponse{
		Entries:      filteredEntries,
		TotalHours:   totalHours,
		TotalRevenue: totalRevenue,
	})
}

// GetTimeEntry godoc
// @Summary Get a time entry by ID
// @Description Get a specific time entry by ID for the authenticated user
// @Tags time-entries
// @Produce json
// @Security BearerAuth
// @Param id path int true "Time Entry ID"
// @Success 200 {object} models.TimeEntryResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/time-entries/{id} [get]
func (h *TimeEntryHandler) GetTimeEntry(c echo.Context) error {
	userID := c.Get("user_id").(int32)

	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid time entry ID"})
	}

	timeEntry, err := h.queries.GetTimeEntryByID(c.Request().Context(), db.GetTimeEntryByIDParams{
		ID:     int32(id),
		UserID: userID,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Time entry not found"})
		}
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to fetch time entry"})
	}

	return c.JSON(http.StatusOK, getTimeEntryByIDRowToResponse(timeEntry))
}

// UpdateTimeEntry godoc
// @Summary Update a time entry
// @Description Update a time entry's information
// @Tags time-entries
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Time Entry ID"
// @Param request body models.UpdateTimeEntryRequest true "Update Time Entry Request"
// @Success 200 {object} models.TimeEntryResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/time-entries/{id} [put]
func (h *TimeEntryHandler) UpdateTimeEntry(c echo.Context) error {
	userID := c.Get("user_id").(int32)

	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid time entry ID"})
	}

	var req models.UpdateTimeEntryRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request body"})
	}

	// Parse date in local timezone
	date, err := time.ParseInLocation("2006-01-02", req.Date, time.Local)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid date format. Use YYYY-MM-DD"})
	}

	// Get existing time entry to check if client changed
	existingEntry, err := h.queries.GetTimeEntryByID(c.Request().Context(), db.GetTimeEntryByIDParams{
		ID:     int32(id),
		UserID: userID,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Time entry not found"})
		}
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to fetch time entry"})
	}

	// Determine hourly rate: if client changed, fetch new client's rate; otherwise keep existing rate
	hourlyRate := existingEntry.HourlyRate
	if existingEntry.ClientID != req.ClientID {
		client, err := h.queries.GetClientByID(c.Request().Context(), db.GetClientByIDParams{
			ID:     req.ClientID,
			UserID: userID,
		})
		if err != nil {
			if err == sql.ErrNoRows {
				return c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Client not found"})
			}
			return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to fetch client"})
		}
		hourlyRate = client.HourlyRate
	}

	timeEntry, err := h.queries.UpdateTimeEntry(c.Request().Context(), db.UpdateTimeEntryParams{
		ID:          int32(id),
		UserID:      userID,
		ClientID:    req.ClientID,
		Date:        date,
		Hours:       fmt.Sprintf("%.2f", req.Hours),
		Description: sql.NullString{String: req.Description, Valid: req.Description != ""},
		HourlyRate:  hourlyRate,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Time entry not found"})
		}
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to update time entry"})
	}

	return c.JSON(http.StatusOK, updateTimeEntryRowToResponse(timeEntry))
}

// DeleteTimeEntry godoc
// @Summary Delete a time entry
// @Description Delete a time entry by ID
// @Tags time-entries
// @Produce json
// @Security BearerAuth
// @Param id path int true "Time Entry ID"
// @Success 204 "No Content"
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/time-entries/{id} [delete]
func (h *TimeEntryHandler) DeleteTimeEntry(c echo.Context) error {
	userID := c.Get("user_id").(int32)

	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid time entry ID"})
	}

	err = h.queries.DeleteTimeEntry(c.Request().Context(), db.DeleteTimeEntryParams{
		ID:     int32(id),
		UserID: userID,
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to delete time entry"})
	}

	return c.NoContent(http.StatusNoContent)
}

// GetHeatmap godoc
// @Summary Get heatmap data for time entries
// @Description Get heatmap data for a specific date range
// @Tags time-entries
// @Produce json
// @Security BearerAuth
// @Param start_date query string true "Start date in YYYY-MM-DD format"
// @Param end_date query string true "End date in YYYY-MM-DD format"
// @Success 200 {object} models.HeatmapResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/time-entries/heatmap [get]
func (h *TimeEntryHandler) GetHeatmap(c echo.Context) error {
	userID := c.Get("user_id").(int32)

	// Get start and end dates from query params
	startDateStr := c.QueryParam("start_date")
	endDateStr := c.QueryParam("end_date")

	if startDateStr == "" || endDateStr == "" {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "start_date and end_date parameters are required (format: YYYY-MM-DD)"})
	}

	// Parse dates
	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid start_date format. Use YYYY-MM-DD"})
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid end_date format. Use YYYY-MM-DD"})
	}

	// Validate date range
	if endDate.Before(startDate) {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "end_date must be after start_date"})
	}

	// Set times to cover the full day range
	startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, time.UTC)
	endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 999999999, time.UTC)

	// Get time entries only within the date range
	entries, err := h.queries.GetTimeEntriesByDateRange(c.Request().Context(), db.GetTimeEntriesByDateRangeParams{
		UserID: userID,
		Date:   startDate,
		Date_2: endDate,
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to fetch heatmap data"})
	}

	// Build heatmap data (aggregated hours per day)
	heatmapData := make(map[string]float64)
	for _, entry := range entries {
		dateKey := entry.Date.Format("2006-01-02")
		hours, _ := strconv.ParseFloat(entry.TotalHours, 64)
		heatmapData[dateKey] = hours
	}

	// Get individual time entries within date range (for tooltips)
	detailedEntries, err := h.queries.GetDetailedTimeEntriesByDateRange(c.Request().Context(), db.GetDetailedTimeEntriesByDateRangeParams{
		UserID: userID,
		Date:   startDate,
		Date_2: endDate,
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to fetch detailed time entries"})
	}

	// Get clients to include client names
	clients, err := h.queries.GetClientsByUserID(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to fetch clients"})
	}

	// Create clients map
	clientsMap := make(map[int32]string)
	for _, client := range clients {
		clientsMap[client.ID] = client.Name
	}

	// Build entries map for tooltips
	entriesMap := make(map[string][]models.TimeEntryResponse)
	for _, entry := range detailedEntries {
		dateKey := entry.Date.Format("2006-01-02")
		clientName := "Unknown"
		if name, ok := clientsMap[entry.ClientID]; ok {
			clientName = name
		}

		entryResponse := toTimeEntryResponse(db.TimeEntry{
			ID:          entry.ID,
			UserID:      entry.UserID,
			ClientID:    entry.ClientID,
			Date:        entry.Date,
			Hours:       entry.Hours,
			Description: entry.Description,
			HourlyRate:  entry.HourlyRate,
			CreatedAt:   entry.CreatedAt,
			UpdatedAt:   entry.UpdatedAt,
		})
		entryResponse.ClientName = clientName
		entriesMap[dateKey] = append(entriesMap[dateKey], entryResponse)
	}

	// Calculate statistics
	totalHours := 0.0
	daysWorked := 0
	for _, hours := range heatmapData {
		totalHours += hours
		if hours > 0 {
			daysWorked++
		}
	}

	// Calculate total days in the date range
	totalDays := int(endDate.Sub(startDate).Hours()/24) + 1
	daysOff := totalDays - daysWorked
	averageHours := 0.0
	if daysWorked > 0 {
		averageHours = totalHours / float64(daysWorked)
	}

	response := models.HeatmapResponse{
		StartDate:    startDate.Format("2006-01-02"),
		EndDate:      endDate.Format("2006-01-02"),
		Data:         heatmapData,
		Entries:      entriesMap,
		TotalHours:   totalHours,
		DaysWorked:   daysWorked,
		DaysOff:      daysOff,
		AverageHours: averageHours,
	}

	return c.JSON(http.StatusOK, response)
}

func toTimeEntryResponse(entry db.TimeEntry) models.TimeEntryResponse {
	hours, _ := strconv.ParseFloat(entry.Hours, 64)
	hourlyRate, _ := strconv.ParseFloat(entry.HourlyRate.String, 64)
	return models.TimeEntryResponse{
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

func createTimeEntryRowToResponse(entry db.CreateTimeEntryRow) models.TimeEntryResponse {
	hours, _ := strconv.ParseFloat(entry.Hours, 64)
	hourlyRate, _ := strconv.ParseFloat(entry.HourlyRate.String, 64)
	return models.TimeEntryResponse{
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

func getTimeEntryByIDRowToResponse(entry db.GetTimeEntryByIDRow) models.TimeEntryResponse {
	hours, _ := strconv.ParseFloat(entry.Hours, 64)
	hourlyRate, _ := strconv.ParseFloat(entry.HourlyRate.String, 64)
	return models.TimeEntryResponse{
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

func getTimeEntriesByUserIDRowToResponse(entry db.GetTimeEntriesByUserIDRow) models.TimeEntryResponse {
	hours, _ := strconv.ParseFloat(entry.Hours, 64)
	hourlyRate, _ := strconv.ParseFloat(entry.HourlyRate.String, 64)
	return models.TimeEntryResponse{
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

func updateTimeEntryRowToResponse(entry db.UpdateTimeEntryRow) models.TimeEntryResponse {
	hours, _ := strconv.ParseFloat(entry.Hours, 64)
	hourlyRate, _ := strconv.ParseFloat(entry.HourlyRate.String, 64)
	return models.TimeEntryResponse{
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
