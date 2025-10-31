package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"worklio-api/internal/db"
	"worklio-api/internal/models"
	"worklio-api/internal/utils"

	"github.com/jung-kurt/gofpdf"
	"github.com/labstack/echo/v4"
)

type InvoiceHandler struct {
	queries *db.Queries
}

func NewInvoiceHandler(queries *db.Queries) *InvoiceHandler {
	return &InvoiceHandler{
		queries: queries,
	}
}

// CreateInvoice godoc
// @Summary Create a new invoice
// @Description Create a new invoice with time entries for the authenticated user
// @Tags invoices
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.CreateInvoiceRequest true "Create Invoice Request"
// @Success 201 {object} models.InvoiceResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/invoices [post]
func (h *InvoiceHandler) CreateInvoice(c echo.Context) error {
	userID := c.Get("user_id").(int32)

	var req models.CreateInvoiceRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request body"})
	}

	// Parse dates
	issueDate, err := time.Parse("2006-01-02", req.IssueDate)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid issue date format. Use YYYY-MM-DD"})
	}

	dueDate, err := time.Parse("2006-01-02", req.DueDate)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid due date format. Use YYYY-MM-DD"})
	}

	// Create invoice
	invoice, err := h.queries.CreateInvoice(c.Request().Context(), db.CreateInvoiceParams{
		UserID:        userID,
		ClientID:      req.ClientID,
		InvoiceNumber: req.InvoiceNumber,
		IssueDate:     issueDate,
		DueDate:       dueDate,
		Status:        req.Status,
		Notes:         sql.NullString{String: req.Notes, Valid: req.Notes != ""},
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to create invoice"})
	}

	// Add time entries to invoice
	for _, timeEntryID := range req.TimeEntryIDs {
		err := h.queries.AddTimeEntryToInvoice(c.Request().Context(), db.AddTimeEntryToInvoiceParams{
			InvoiceID:   invoice.ID,
			TimeEntryID: timeEntryID,
		})
		if err != nil {
			return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to add time entries to invoice"})
		}
	}

	// Get the complete invoice with time entries
	return h.getInvoiceResponse(c, invoice.ID, userID)
}

// GetInvoices godoc
// @Summary Get all invoices
// @Description Get all invoices for the authenticated user
// @Tags invoices
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.InvoiceResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/invoices [get]
func (h *InvoiceHandler) GetInvoices(c echo.Context) error {
	userID := c.Get("user_id").(int32)

	invoices, err := h.queries.GetInvoicesByUserID(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to fetch invoices"})
	}

	response := make([]models.InvoiceResponse, len(invoices))
	for i, invoice := range invoices {
		// Get time entries for this invoice
		timeEntries, err := h.queries.GetInvoiceTimeEntries(c.Request().Context(), invoice.ID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to fetch invoice time entries"})
		}

		response[i] = h.buildInvoiceResponseWithClient(invoice, timeEntries)
	}

	return c.JSON(http.StatusOK, response)
}

// GetInvoice godoc
// @Summary Get an invoice by ID
// @Description Get a specific invoice by ID for the authenticated user
// @Tags invoices
// @Produce json
// @Security BearerAuth
// @Param id path int true "Invoice ID"
// @Success 200 {object} models.InvoiceResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/invoices/{id} [get]
func (h *InvoiceHandler) GetInvoice(c echo.Context) error {
	userID := c.Get("user_id").(int32)

	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid invoice ID"})
	}

	return h.getInvoiceResponse(c, int32(id), userID)
}

// UpdateInvoice godoc
// @Summary Update an invoice
// @Description Update an invoice's information
// @Tags invoices
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Invoice ID"
// @Param request body models.UpdateInvoiceRequest true "Update Invoice Request"
// @Success 200 {object} models.InvoiceResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/invoices/{id} [put]
func (h *InvoiceHandler) UpdateInvoice(c echo.Context) error {
	userID := c.Get("user_id").(int32)

	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid invoice ID"})
	}

	var req models.UpdateInvoiceRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request body"})
	}

	// Parse dates
	issueDate, err := time.Parse("2006-01-02", req.IssueDate)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid issue date format. Use YYYY-MM-DD"})
	}

	dueDate, err := time.Parse("2006-01-02", req.DueDate)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid due date format. Use YYYY-MM-DD"})
	}

	invoice, err := h.queries.UpdateInvoice(c.Request().Context(), db.UpdateInvoiceParams{
		ID:            int32(id),
		UserID:        userID,
		ClientID:      req.ClientID,
		InvoiceNumber: req.InvoiceNumber,
		IssueDate:     issueDate,
		DueDate:       dueDate,
		Status:        req.Status,
		Notes:         sql.NullString{String: req.Notes, Valid: req.Notes != ""},
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Invoice not found"})
		}
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to update invoice"})
	}

	return h.getInvoiceResponse(c, invoice.ID, userID)
}

// UpdateInvoiceStatus godoc
// @Summary Update invoice status
// @Description Update only the status of an invoice
// @Tags invoices
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Invoice ID"
// @Param request body models.UpdateInvoiceStatusRequest true "Update Invoice Status Request"
// @Success 200 {object} models.InvoiceResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/invoices/{id}/status [patch]
func (h *InvoiceHandler) UpdateInvoiceStatus(c echo.Context) error {
	userID := c.Get("user_id").(int32)

	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid invoice ID"})
	}

	var req models.UpdateInvoiceStatusRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request body"})
	}

	invoice, err := h.queries.UpdateInvoiceStatus(c.Request().Context(), db.UpdateInvoiceStatusParams{
		ID:     int32(id),
		UserID: userID,
		Status: req.Status,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Invoice not found"})
		}
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to update invoice status"})
	}

	return h.getInvoiceResponse(c, invoice.ID, userID)
}

// DeleteInvoice godoc
// @Summary Delete an invoice
// @Description Delete an invoice by ID
// @Tags invoices
// @Produce json
// @Security BearerAuth
// @Param id path int true "Invoice ID"
// @Success 204 "No Content"
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/invoices/{id} [delete]
func (h *InvoiceHandler) DeleteInvoice(c echo.Context) error {
	userID := c.Get("user_id").(int32)

	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid invoice ID"})
	}

	err = h.queries.DeleteInvoice(c.Request().Context(), db.DeleteInvoiceParams{
		ID:     int32(id),
		UserID: userID,
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to delete invoice"})
	}

	return c.NoContent(http.StatusNoContent)
}

// GetAvailableTimeEntries godoc
// @Summary Get available time entries for invoicing
// @Description Get time entries for a client that haven't been invoiced yet
// @Tags invoices
// @Produce json
// @Security BearerAuth
// @Param client_id query int true "Client ID"
// @Success 200 {array} models.TimeEntryResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/invoices/available-time-entries [get]
func (h *InvoiceHandler) GetAvailableTimeEntries(c echo.Context) error {
	userID := c.Get("user_id").(int32)

	clientIDStr := c.QueryParam("client_id")
	if clientIDStr == "" {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "client_id query parameter is required"})
	}

	clientID, err := strconv.ParseInt(clientIDStr, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid client ID"})
	}

	timeEntries, err := h.queries.GetAvailableTimeEntriesForClient(c.Request().Context(), db.GetAvailableTimeEntriesForClientParams{
		ClientID: int32(clientID),
		UserID:   userID,
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to fetch available time entries"})
	}

	response := make([]models.TimeEntryResponse, len(timeEntries))
	for i, entry := range timeEntries {
		hours, _ := strconv.ParseFloat(entry.Hours, 64)
		response[i] = models.TimeEntryResponse{
			ID:          entry.ID,
			UserID:      entry.UserID,
			ClientID:    entry.ClientID,
			Date:        entry.Date.Format("2006-01-02"),
			Hours:       hours,
			Description: entry.Description.String,
			CreatedAt:   entry.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:   entry.UpdatedAt.Time.Format("2006-01-02T15:04:05Z"),
		}
	}

	return c.JSON(http.StatusOK, response)
}

// Helper functions
func (h *InvoiceHandler) getInvoiceResponse(c echo.Context, invoiceID int32, userID int32) error {
	invoice, err := h.queries.GetInvoiceByID(c.Request().Context(), db.GetInvoiceByIDParams{
		ID:     invoiceID,
		UserID: userID,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Invoice not found"})
		}
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to fetch invoice"})
	}

	timeEntries, err := h.queries.GetInvoiceTimeEntries(c.Request().Context(), invoiceID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to fetch invoice time entries"})
	}

	response := h.buildInvoiceResponseWithClient(invoice, timeEntries)
	return c.JSON(http.StatusOK, response)
}

func (h *InvoiceHandler) buildInvoiceResponseWithClient(invoice db.Invoice, timeEntries []db.GetInvoiceTimeEntriesRow) models.InvoiceResponse {
	timeEntryResponses := make([]models.TimeEntryResponse, len(timeEntries))
	totalHours := 0.0
	totalAmount := 0.0

	for i, entry := range timeEntries {
		hours, _ := strconv.ParseFloat(entry.Hours, 64)
		hourlyRate, _ := strconv.ParseFloat(entry.HourlyRate.String, 64)
		totalHours += hours
		totalAmount += hours * hourlyRate

		timeEntryResponses[i] = models.TimeEntryResponse{
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

	return models.InvoiceResponse{
		ID:            invoice.ID,
		UserID:        invoice.UserID,
		ClientID:      invoice.ClientID,
		InvoiceNumber: invoice.InvoiceNumber,
		IssueDate:     invoice.IssueDate.Format("2006-01-02"),
		DueDate:       invoice.DueDate.Format("2006-01-02"),
		Status:        invoice.Status,
		Notes:         invoice.Notes.String,
		TimeEntries:   timeEntryResponses,
		TotalHours:    totalHours,
		TotalAmount:   totalAmount,
		CreatedAt:     invoice.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:     invoice.UpdatedAt.Time.Format("2006-01-02T15:04:05Z"),
	}
}

// DownloadInvoicePDF godoc
// @Summary Download invoice as PDF
// @Description Download an invoice as a PDF file
// @Tags invoices
// @Produce application/pdf
// @Security BearerAuth
// @Param id path int true "Invoice ID"
// @Success 200 {file} binary
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/invoices/{id}/pdf [get]
func (h *InvoiceHandler) DownloadInvoicePDF(c echo.Context) error {
	userID := c.Get("user_id").(int32)

	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid invoice ID"})
	}

	// Get invoice data
	invoice, err := h.queries.GetInvoiceByID(c.Request().Context(), db.GetInvoiceByIDParams{
		ID:     int32(id),
		UserID: userID,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Invoice not found"})
		}
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to fetch invoice"})
	}

	// Get client data
	client, err := h.queries.GetClientByID(c.Request().Context(), db.GetClientByIDParams{
		ID:     invoice.ClientID,
		UserID: userID,
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to fetch client data"})
	}

	// Get time entries
	timeEntries, err := h.queries.GetInvoiceTimeEntries(c.Request().Context(), int32(id))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to fetch time entries"})
	}

	// Use client's currency for invoice
	currency := client.Currency
	if currency == "" {
		currency = "USD" // Default fallback
	}

	// Generate PDF
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetMargins(20, 20, 20)
	pdf.SetAutoPageBreak(true, 20)

	// Header Section with Blue Background
	pdf.SetFillColor(30, 58, 138) // blue-900
	pdf.Rect(0, 0, 210, 50, "F")

	// Invoice Title
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 32)
	pdf.SetY(15)
	pdf.Cell(0, 10, "INVOICE")
	pdf.Ln(12)

	// Invoice Number
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 8, invoice.InvoiceNumber)
	pdf.Ln(20)

	// Reset text color for body
	pdf.SetTextColor(0, 0, 0)

	// Two column layout for Bill To and Invoice Details
	leftX := 20.0
	rightX := 115.0
	currentY := pdf.GetY()

	// Left Column - Bill To Section
	pdf.SetXY(leftX, currentY)
	pdf.SetFillColor(241, 245, 249) // slate-100
	pdf.Rect(leftX, currentY, 85, 45, "F")

	pdf.SetXY(leftX+5, currentY+5)
	pdf.SetFont("Arial", "B", 10)
	pdf.SetTextColor(100, 116, 139) // slate-500
	pdf.Cell(0, 5, "BILL TO")
	pdf.Ln(7)

	pdf.SetX(leftX+5)
	pdf.SetFont("Arial", "B", 12)
	pdf.SetTextColor(0, 0, 0)
	pdf.Cell(0, 6, client.Name)
	pdf.Ln(6)

	pdf.SetX(leftX+5)
	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(71, 85, 105) // slate-600
	pdf.Cell(0, 5, client.Email)
	pdf.Ln(5)

	if client.Company.Valid && client.Company.String != "" {
		pdf.SetX(leftX+5)
		pdf.Cell(0, 5, client.Company.String)
		pdf.Ln(5)
	}

	// Right Column - Invoice Details
	pdf.SetXY(rightX, currentY)
	pdf.SetFillColor(241, 245, 249) // slate-100
	pdf.Rect(rightX, currentY, 75, 45, "F")

	pdf.SetXY(rightX+5, currentY+5)
	pdf.SetFont("Arial", "B", 10)
	pdf.SetTextColor(100, 116, 139) // slate-500
	pdf.Cell(0, 5, "INVOICE DETAILS")
	pdf.Ln(7)

	// Status
	pdf.SetX(rightX+5)
	pdf.SetFont("Arial", "B", 9)
	pdf.SetTextColor(71, 85, 105)
	pdf.Cell(25, 5, "Status:")
	pdf.SetFont("Arial", "", 9)
	statusColor := getStatusColorRGB(invoice.Status)
	pdf.SetTextColor(statusColor[0], statusColor[1], statusColor[2])
	pdf.Cell(0, 5, invoice.Status)
	pdf.Ln(5)

	// Issue Date
	pdf.SetX(rightX+5)
	pdf.SetFont("Arial", "B", 9)
	pdf.SetTextColor(71, 85, 105)
	pdf.Cell(25, 5, "Issue Date:")
	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(0, 0, 0)
	pdf.Cell(0, 5, invoice.IssueDate.Format("Jan 2, 2006"))
	pdf.Ln(5)

	// Due Date
	pdf.SetX(rightX+5)
	pdf.SetFont("Arial", "B", 9)
	pdf.SetTextColor(71, 85, 105)
	pdf.Cell(25, 5, "Due Date:")
	pdf.SetFont("Arial", "", 9)
	if invoice.Status == "overdue" {
		pdf.SetTextColor(239, 68, 68) // red
	} else {
		pdf.SetTextColor(0, 0, 0)
	}
	pdf.Cell(0, 5, invoice.DueDate.Format("Jan 2, 2006"))

	pdf.SetTextColor(0, 0, 0)
	pdf.SetY(currentY + 50)
	pdf.Ln(10)

	// Line Items Header
	pdf.SetFont("Arial", "B", 12)
	pdf.SetTextColor(30, 58, 138) // blue-900
	pdf.Cell(0, 8, "Line Items")
	pdf.Ln(10)

	// Table Header with colored background
	pdf.SetFillColor(30, 58, 138) // blue-900
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 9)
	pdf.SetDrawColor(30, 58, 138)

	pdf.CellFormat(30, 9, "DATE", "1", 0, "L", true, 0, "")
	pdf.CellFormat(68, 9, "DESCRIPTION", "1", 0, "L", true, 0, "")
	pdf.CellFormat(22, 9, "HOURS", "1", 0, "C", true, 0, "")
	pdf.CellFormat(22, 9, "RATE", "1", 0, "C", true, 0, "")
	pdf.CellFormat(28, 9, "AMOUNT", "1", 0, "R", true, 0, "")
	pdf.Ln(-1)

	// Table Body with alternating row colors
	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(0, 0, 0)
	pdf.SetDrawColor(226, 232, 240) // slate-200
	totalHours := 0.0
	totalAmount := 0.0
	hourlyRateFloat, _ := strconv.ParseFloat(client.HourlyRate.String, 64)

	rowIndex := 0
	for _, entry := range timeEntries {
		hours, _ := strconv.ParseFloat(entry.Hours, 64)
		amount := hours * hourlyRateFloat
		totalHours += hours
		totalAmount += amount

		description := entry.Description.String
		if description == "" {
			description = "No description"
		}

		// Truncate long descriptions
		if len(description) > 40 {
			description = description[:37] + "..."
		}

		// Alternate row colors
		if rowIndex%2 == 0 {
			pdf.SetFillColor(248, 250, 252) // slate-50
		} else {
			pdf.SetFillColor(255, 255, 255) // white
		}

		pdf.CellFormat(30, 8, entry.Date.Format("Jan 2, 2006"), "1", 0, "L", true, 0, "")
		pdf.CellFormat(68, 8, description, "1", 0, "L", true, 0, "")
		pdf.CellFormat(22, 8, utils.FormatNumber(hours, 2), "1", 0, "C", true, 0, "")
		pdf.CellFormat(22, 8, utils.FormatCurrencyRateForPDF(hourlyRateFloat, currency), "1", 0, "C", true, 0, "")
		pdf.SetFont("Arial", "B", 9)
		pdf.CellFormat(28, 8, utils.FormatCurrencyForPDF(amount, currency), "1", 0, "R", true, 0, "")
		pdf.SetFont("Arial", "", 9)
		pdf.Ln(-1)
		rowIndex++
	}

	// Subtotal Section
	pdf.SetDrawColor(30, 58, 138)
	pdf.SetLineWidth(0.5)
	pdf.Line(20, pdf.GetY(), 190, pdf.GetY())
	pdf.Ln(3)

	// Summary box on the right - aligned with HOURS, RATE, AMOUNT columns
	// Table structure: 30 (DATE) + 68 (DESC) = 98mm, then HOURS(22) + RATE(22) + AMOUNT(28) = 72mm
	summaryLabelX := 118.0  // Start where HOURS column starts (20 + 30 + 68)
	summaryValueX := 162.0  // Start where AMOUNT column starts (20 + 30 + 68 + 22 + 22)
	summaryY := pdf.GetY()

	// Total Hours Row
	pdf.SetXY(summaryLabelX, summaryY)
	pdf.SetFont("Arial", "B", 10)
	pdf.SetTextColor(71, 85, 105) // slate-600
	pdf.CellFormat(44, 7, "Total Hours:", "", 0, "L", false, 0, "")
	pdf.SetXY(summaryValueX, summaryY)
	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(0, 0, 0)
	pdf.CellFormat(28, 7, utils.FormatNumber(totalHours, 2), "", 0, "R", false, 0, "")
	pdf.Ln(9)

	// Total Amount with colored background
	pdf.SetX(summaryLabelX)
	pdf.SetFillColor(30, 58, 138) // blue-900
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(44, 10, "TOTAL:", "1", 0, "L", true, 0, "")
	pdf.SetFont("Arial", "B", 14)
	pdf.CellFormat(28, 10, utils.FormatCurrencyForPDF(totalAmount, currency), "1", 0, "R", true, 0, "")
	pdf.Ln(15)

	// Notes Section
	pdf.SetTextColor(0, 0, 0)
	if invoice.Notes.Valid && invoice.Notes.String != "" {
		pdf.SetFont("Arial", "B", 11)
		pdf.SetTextColor(30, 58, 138) // blue-900
		pdf.Cell(0, 8, "Notes")
		pdf.Ln(8)

		pdf.SetFont("Arial", "", 10)
		pdf.SetTextColor(71, 85, 105) // slate-600
		pdf.SetFillColor(248, 250, 252) // slate-50

		// Draw background for notes
		noteY := pdf.GetY()
		pdf.Rect(20, noteY, 170, -1, "F") // Will be auto-sized
		pdf.MultiCell(170, 6, invoice.Notes.String, "", "L", false)
		pdf.Ln(5)
	}

	// Footer
	pdf.SetY(-30)
	pdf.SetFont("Arial", "I", 9)
	pdf.SetTextColor(148, 163, 184) // slate-400
	pdf.CellFormat(0, 10, "Thank you for your business!", "", 0, "C", false, 0, "")

	// Generate PDF and return as response
	filename := fmt.Sprintf("%s.pdf", invoice.InvoiceNumber)
	c.Response().Header().Set("Content-Type", "application/pdf")
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	err = pdf.Output(c.Response().Writer)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to generate PDF"})
	}

	return nil
}

func getStatusColorRGB(status string) [3]int {
	switch status {
	case "draft":
		return [3]int{107, 114, 128} // gray
	case "sent":
		return [3]int{59, 130, 246} // blue
	case "paid":
		return [3]int{34, 197, 94} // green
	case "overdue":
		return [3]int{239, 68, 68} // red
	default:
		return [3]int{107, 114, 128} // gray
	}
}
