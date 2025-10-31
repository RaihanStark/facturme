package handlers

import (
	"database/sql"
	"fmt"
	"math/rand"
	"net/http"
	"time"
	"worklio-api/internal/db"
	"worklio-api/internal/models"

	"github.com/labstack/echo/v4"
)

type DemoHandler struct {
	queries *db.Queries
}

func NewDemoHandler(queries *db.Queries) *DemoHandler {
	return &DemoHandler{
		queries: queries,
	}
}

// GenerateDemoData godoc
// @Summary Generate demo data
// @Description Create sample clients, time entries, and invoices for demonstration
// @Tags demo
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/demo/generate [post]
func (h *DemoHandler) GenerateDemoData(c echo.Context) error {
	// Get user ID from context
	userID, ok := c.Get("user_id").(int32)
	if !ok {
		return c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Unauthorized"})
	}

	ctx := c.Request().Context()

	// Create 3 demo clients
	acmeClient, err := h.queries.CreateClient(ctx, db.CreateClientParams{
		UserID:     userID,
		Name:       "🎭 Acme Corp (Demo)",
		Email:      "contact@acmecorp.demo",
		Company:    sql.NullString{String: "Acme Corporation", Valid: true},
		HourlyRate: sql.NullString{String: "75", Valid: true},
		Currency:   "USD",
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to create demo client"})
	}

	techStartClient, err := h.queries.CreateClient(ctx, db.CreateClientParams{
		UserID:     userID,
		Name:       "🎭 TechStart Inc (Demo)",
		Email:      "hello@techstart.demo",
		Company:    sql.NullString{String: "TechStart Incorporated", Valid: true},
		HourlyRate: sql.NullString{String: "100", Valid: true},
		Currency:   "USD",
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to create demo client"})
	}

	designStudioClient, err := h.queries.CreateClient(ctx, db.CreateClientParams{
		UserID:     userID,
		Name:       "🎭 Creative Studio (Demo)",
		Email:      "info@creativestudio.demo",
		Company:    sql.NullString{String: "Creative Design Studio", Valid: true},
		HourlyRate: sql.NullString{String: "85", Valid: true},
		Currency:   "USD",
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to create demo client"})
	}

	// Store demo clients in array - ONLY these clients will get time entries
	demoClients := []db.CreateClientRow{acmeClient, techStartClient, designStudioClient}

	// Store demo client IDs for validation
	demoClientIDs := map[int32]bool{
		acmeClient.ID:        true,
		techStartClient.ID:   true,
		designStudioClient.ID: true,
	}

	// Task descriptions
	tasks := []string{
		"Frontend development",
		"Backend API work",
		"Database optimization",
		"UI/UX improvements",
		"Bug fixes",
		"Code review",
		"Feature development",
		"Team meeting",
		"Client consultation",
		"Project planning",
		"Documentation",
		"Testing and QA",
		"Performance optimization",
		"Security updates",
		"Design mockups",
		"Responsive layout",
		"Integration work",
		"Research and analysis",
	}

	// Create 50 time entries spread over past 60 days
	// IMPORTANT: Only create entries for the 3 demo clients created above
	// Do NOT create entries for any existing clients (including user's first client)
	var acmeEntries []int32
	var techStartEntries []int32
	var designStudioEntries []int32

	for i := 0; i < 50; i++ {
		// Cycle through ONLY the demo clients we just created
		client := demoClients[i%3]

		// Safety check: ensure this is a demo client
		if !demoClientIDs[client.ID] {
			continue // Skip if somehow not a demo client
		}

		daysAgo := int32((float64(i) / 50.0) * 60.0)
		if daysAgo > 0 {
			daysAgo += int32(rand.Intn(3))
		}
		hours := rand.Intn(7) + 2 // 2-8 hours
		task := tasks[rand.Intn(len(tasks))]

		entryDate := time.Now().AddDate(0, 0, -int(daysAgo))

		entry, err := h.queries.CreateTimeEntry(ctx, db.CreateTimeEntryParams{
			UserID:      userID,
			ClientID:    client.ID,
			Date:        entryDate,
			Hours:       fmt.Sprintf("%d", hours),
			Description: sql.NullString{String: task, Valid: true},
			HourlyRate:  client.HourlyRate,
		})
		if err != nil {
			return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to create demo time entry"})
		}

		// Group entries by client
		if client.ID == acmeClient.ID {
			acmeEntries = append(acmeEntries, entry.ID)
		} else if client.ID == techStartClient.ID {
			techStartEntries = append(techStartEntries, entry.ID)
		} else {
			designStudioEntries = append(designStudioEntries, entry.ID)
		}
	}

	// Create 5 invoices with different statuses and link time entries
	usedEntries := make(map[int32]bool) // Track which entries are already in invoices

	// Invoice 1: Paid (oldest)
	if len(acmeEntries) >= 5 {
		invoice1, err := h.queries.CreateInvoice(ctx, db.CreateInvoiceParams{
			UserID:        userID,
			ClientID:      acmeClient.ID,
			InvoiceNumber: "DEMO-001",
			IssueDate:     time.Now().AddDate(0, 0, -50),
			DueDate:       time.Now().AddDate(0, 0, -36),
			Status:        "paid",
			Notes:         sql.NullString{String: "Demo invoice - Paid", Valid: true},
		})
		if err != nil {
			return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to create demo invoice"})
		}

		// Add first 5 acme entries to this invoice
		for i := 0; i < 5 && i < len(acmeEntries); i++ {
			if !usedEntries[acmeEntries[i]] {
				err = h.queries.AddTimeEntryToInvoice(ctx, db.AddTimeEntryToInvoiceParams{
					InvoiceID:   invoice1.ID,
					TimeEntryID: acmeEntries[i],
				})
				if err != nil {
					return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to add time entry to invoice"})
				}
				usedEntries[acmeEntries[i]] = true
			}
		}
	}

	// Invoice 2: Paid
	if len(techStartEntries) >= 4 {
		invoice2, err := h.queries.CreateInvoice(ctx, db.CreateInvoiceParams{
			UserID:        userID,
			ClientID:      techStartClient.ID,
			InvoiceNumber: "DEMO-002",
			IssueDate:     time.Now().AddDate(0, 0, -40),
			DueDate:       time.Now().AddDate(0, 0, -26),
			Status:        "paid",
			Notes:         sql.NullString{String: "Demo invoice - Paid", Valid: true},
		})
		if err != nil {
			return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to create demo invoice"})
		}

		// Add first 4 techStart entries to this invoice
		for i := 0; i < 4 && i < len(techStartEntries); i++ {
			if !usedEntries[techStartEntries[i]] {
				err = h.queries.AddTimeEntryToInvoice(ctx, db.AddTimeEntryToInvoiceParams{
					InvoiceID:   invoice2.ID,
					TimeEntryID: techStartEntries[i],
				})
				if err != nil {
					return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to add time entry to invoice"})
				}
				usedEntries[techStartEntries[i]] = true
			}
		}
	}

	// Invoice 3: Sent (awaiting payment)
	if len(designStudioEntries) >= 6 {
		invoice3, err := h.queries.CreateInvoice(ctx, db.CreateInvoiceParams{
			UserID:        userID,
			ClientID:      designStudioClient.ID,
			InvoiceNumber: "DEMO-003",
			IssueDate:     time.Now().AddDate(0, 0, -20),
			DueDate:       time.Now().AddDate(0, 0, -6),
			Status:        "sent",
			Notes:         sql.NullString{String: "Demo invoice - Awaiting payment", Valid: true},
		})
		if err != nil {
			return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to create demo invoice"})
		}

		// Add first 6 designStudio entries to this invoice
		for i := 0; i < 6 && i < len(designStudioEntries); i++ {
			if !usedEntries[designStudioEntries[i]] {
				err = h.queries.AddTimeEntryToInvoice(ctx, db.AddTimeEntryToInvoiceParams{
					InvoiceID:   invoice3.ID,
					TimeEntryID: designStudioEntries[i],
				})
				if err != nil {
					return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to add time entry to invoice"})
				}
				usedEntries[designStudioEntries[i]] = true
			}
		}
	}

	// Invoice 4: Overdue
	if len(acmeEntries) >= 10 {
		invoice4, err := h.queries.CreateInvoice(ctx, db.CreateInvoiceParams{
			UserID:        userID,
			ClientID:      acmeClient.ID,
			InvoiceNumber: "DEMO-004",
			IssueDate:     time.Now().AddDate(0, 0, -25),
			DueDate:       time.Now().AddDate(0, 0, -11),
			Status:        "overdue",
			Notes:         sql.NullString{String: "Demo invoice - Payment overdue", Valid: true},
		})
		if err != nil {
			return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to create demo invoice"})
		}

		// Add entries 5-9 from acme (skip first 5 already used)
		for i := 5; i < 10 && i < len(acmeEntries); i++ {
			if !usedEntries[acmeEntries[i]] {
				err = h.queries.AddTimeEntryToInvoice(ctx, db.AddTimeEntryToInvoiceParams{
					InvoiceID:   invoice4.ID,
					TimeEntryID: acmeEntries[i],
				})
				if err != nil {
					return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to add time entry to invoice"})
				}
				usedEntries[acmeEntries[i]] = true
			}
		}
	}

	// Invoice 5: Draft (most recent)
	if len(techStartEntries) >= 9 {
		invoice5, err := h.queries.CreateInvoice(ctx, db.CreateInvoiceParams{
			UserID:        userID,
			ClientID:      techStartClient.ID,
			InvoiceNumber: "DEMO-005",
			IssueDate:     time.Now().AddDate(0, 0, -5),
			DueDate:       time.Now().AddDate(0, 0, 9),
			Status:        "draft",
			Notes:         sql.NullString{String: "Demo invoice - Ready to review and send", Valid: true},
		})
		if err != nil {
			return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to create demo invoice"})
		}

		// Add entries 4-8 from techStart (skip first 4 already used)
		for i := 4; i < 9 && i < len(techStartEntries); i++ {
			if !usedEntries[techStartEntries[i]] {
				err = h.queries.AddTimeEntryToInvoice(ctx, db.AddTimeEntryToInvoiceParams{
					InvoiceID:   invoice5.ID,
					TimeEntryID: techStartEntries[i],
				})
				if err != nil {
					return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to add time entry to invoice"})
				}
				usedEntries[techStartEntries[i]] = true
			}
		}
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Demo data generated successfully",
	})
}

// DeleteDemoData godoc
// @Summary Delete demo data
// @Description Delete all demo clients, time entries, and invoices (cascading delete)
// @Tags demo
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/demo [delete]
func (h *DemoHandler) DeleteDemoData(c echo.Context) error {
	// Get user ID from context
	userID, ok := c.Get("user_id").(int32)
	if !ok {
		return c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Unauthorized"})
	}

	ctx := c.Request().Context()

	// Delete all demo clients (this will cascade delete time entries and invoices due to foreign key constraints)
	err := h.queries.DeleteDemoClients(ctx, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to delete demo data"})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Demo data deleted successfully",
	})
}
