package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"worklio-api/internal/db"
	"worklio-api/internal/models"

	"github.com/labstack/echo/v4"
)

type ClientHandler struct {
	queries *db.Queries
}

func NewClientHandler(queries *db.Queries) *ClientHandler {
	return &ClientHandler{
		queries: queries,
	}
}

// CreateClient godoc
// @Summary Create a new client
// @Description Create a new client for the authenticated user
// @Tags clients
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.CreateClientRequest true "Create Client Request"
// @Success 201 {object} models.ClientResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/clients [post]
func (h *ClientHandler) CreateClient(c echo.Context) error {
	userID := c.Get("user_id").(int32)

	var req models.CreateClientRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request body"})
	}

	// Default to USD if currency is not provided
	currency := req.Currency
	if currency == "" {
		currency = "USD"
	}

	client, err := h.queries.CreateClient(c.Request().Context(), db.CreateClientParams{
		UserID:  userID,
		Name:    req.Name,
		Email:   req.Email,
		Phone:   sql.NullString{String: req.Phone, Valid: req.Phone != ""},
		Company: sql.NullString{String: req.Company, Valid: req.Company != ""},
		Address: sql.NullString{String: req.Address, Valid: req.Address != ""},
		HourlyRate: sql.NullString{
			String: fmt.Sprintf("%.2f", req.HourlyRate),
			Valid:  true,
		},
		Currency: currency,
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to create client"})
	}

	return c.JSON(http.StatusCreated, createClientRowToResponse(client))
}

// GetClients godoc
// @Summary Get all clients
// @Description Get all clients for the authenticated user
// @Tags clients
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.ClientResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/clients [get]
func (h *ClientHandler) GetClients(c echo.Context) error {
	userID := c.Get("user_id").(int32)

	clients, err := h.queries.GetClientsByUserID(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to fetch clients"})
	}

	response := make([]models.ClientResponse, len(clients))
	for i, client := range clients {
		response[i] = getClientsByUserIDRowToResponse(client)
	}

	return c.JSON(http.StatusOK, response)
}

// GetClient godoc
// @Summary Get a client by ID
// @Description Get a specific client by ID for the authenticated user
// @Tags clients
// @Produce json
// @Security BearerAuth
// @Param id path int true "Client ID"
// @Success 200 {object} models.ClientResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/clients/{id} [get]
func (h *ClientHandler) GetClient(c echo.Context) error {
	userID := c.Get("user_id").(int32)

	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid client ID"})
	}

	client, err := h.queries.GetClientByID(c.Request().Context(), db.GetClientByIDParams{
		ID:     int32(id),
		UserID: userID,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Client not found"})
		}
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to fetch client"})
	}

	return c.JSON(http.StatusOK, getClientByIDRowToResponse(client))
}

// UpdateClient godoc
// @Summary Update a client
// @Description Update a client's information
// @Tags clients
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Client ID"
// @Param request body models.UpdateClientRequest true "Update Client Request"
// @Success 200 {object} models.ClientResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/clients/{id} [put]
func (h *ClientHandler) UpdateClient(c echo.Context) error {
	userID := c.Get("user_id").(int32)

	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid client ID"})
	}

	var req models.UpdateClientRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request body"})
	}

	// Default to USD if currency is not provided
	currency := req.Currency
	if currency == "" {
		currency = "USD"
	}

	client, err := h.queries.UpdateClient(c.Request().Context(), db.UpdateClientParams{
		ID:      int32(id),
		UserID:  userID,
		Name:    req.Name,
		Email:   req.Email,
		Phone:   sql.NullString{String: req.Phone, Valid: req.Phone != ""},
		Company: sql.NullString{String: req.Company, Valid: req.Company != ""},
		Address: sql.NullString{String: req.Address, Valid: req.Address != ""},
		HourlyRate: sql.NullString{
			String: fmt.Sprintf("%.2f", req.HourlyRate),
			Valid:  true,
		},
		Currency: currency,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Client not found"})
		}
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to update client"})
	}

	return c.JSON(http.StatusOK, updateClientRowToResponse(client))
}

// DeleteClient godoc
// @Summary Delete a client
// @Description Delete a client by ID
// @Tags clients
// @Produce json
// @Security BearerAuth
// @Param id path int true "Client ID"
// @Success 204 "No Content"
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/clients/{id} [delete]
func (h *ClientHandler) DeleteClient(c echo.Context) error {
	userID := c.Get("user_id").(int32)

	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid client ID"})
	}

	err = h.queries.DeleteClient(c.Request().Context(), db.DeleteClientParams{
		ID:     int32(id),
		UserID: userID,
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to delete client"})
	}

	return c.NoContent(http.StatusNoContent)
}

func createClientRowToResponse(client db.CreateClientRow) models.ClientResponse {
	hourlyRate, _ := strconv.ParseFloat(client.HourlyRate.String, 64)
	return models.ClientResponse{
		ID:         client.ID,
		UserID:     client.UserID,
		Name:       client.Name,
		Email:      client.Email,
		Phone:      client.Phone.String,
		Company:    client.Company.String,
		Address:    client.Address.String,
		HourlyRate: hourlyRate,
		Currency:   client.Currency,
		CreatedAt:  client.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:  client.UpdatedAt.Time.Format("2006-01-02T15:04:05Z"),
	}
}

func getClientsByUserIDRowToResponse(client db.GetClientsByUserIDRow) models.ClientResponse {
	hourlyRate, _ := strconv.ParseFloat(client.HourlyRate.String, 64)
	return models.ClientResponse{
		ID:         client.ID,
		UserID:     client.UserID,
		Name:       client.Name,
		Email:      client.Email,
		Phone:      client.Phone.String,
		Company:    client.Company.String,
		Address:    client.Address.String,
		HourlyRate: hourlyRate,
		Currency:   client.Currency,
		CreatedAt:  client.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:  client.UpdatedAt.Time.Format("2006-01-02T15:04:05Z"),
	}
}

func getClientByIDRowToResponse(client db.GetClientByIDRow) models.ClientResponse {
	hourlyRate, _ := strconv.ParseFloat(client.HourlyRate.String, 64)
	return models.ClientResponse{
		ID:         client.ID,
		UserID:     client.UserID,
		Name:       client.Name,
		Email:      client.Email,
		Phone:      client.Phone.String,
		Company:    client.Company.String,
		Address:    client.Address.String,
		HourlyRate: hourlyRate,
		Currency:   client.Currency,
		CreatedAt:  client.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:  client.UpdatedAt.Time.Format("2006-01-02T15:04:05Z"),
	}
}

func updateClientRowToResponse(client db.UpdateClientRow) models.ClientResponse {
	hourlyRate, _ := strconv.ParseFloat(client.HourlyRate.String, 64)
	return models.ClientResponse{
		ID:         client.ID,
		UserID:     client.UserID,
		Name:       client.Name,
		Email:      client.Email,
		Phone:      client.Phone.String,
		Company:    client.Company.String,
		Address:    client.Address.String,
		HourlyRate: hourlyRate,
		Currency:   client.Currency,
		CreatedAt:  client.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:  client.UpdatedAt.Time.Format("2006-01-02T15:04:05Z"),
	}
}
