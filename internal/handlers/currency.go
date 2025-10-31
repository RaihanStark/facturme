package handlers

import (
	"net/http"
	"strconv"
	"worklio-api/internal/services"

	"github.com/labstack/echo/v4"
)

type CurrencyHandler struct {
	exchangeService *services.ExchangeRateService
}

func NewCurrencyHandler(exchangeService *services.ExchangeRateService) *CurrencyHandler {
	return &CurrencyHandler{
		exchangeService: exchangeService,
	}
}

// SupportedCurrency represents a supported currency
type SupportedCurrency struct {
	Code   string `json:"code"`
	Symbol string `json:"symbol"`
	Name   string `json:"name"`
}

// GetSupportedCurrencies godoc
// @Summary Get supported currencies
// @Description Returns a list of all supported currencies
// @Tags currency
// @Produce json
// @Success 200 {object} []SupportedCurrency
// @Router /api/supported-currencies [get]
func (h *CurrencyHandler) GetSupportedCurrencies(c echo.Context) error {
	currencies := []SupportedCurrency{
		{Code: "USD", Symbol: "$", Name: "US Dollar"},
		{Code: "EUR", Symbol: "€", Name: "Euro"},
		{Code: "GBP", Symbol: "£", Name: "British Pound"},
		{Code: "JPY", Symbol: "¥", Name: "Japanese Yen"},
		{Code: "AUD", Symbol: "A$", Name: "Australian Dollar"},
		{Code: "CAD", Symbol: "C$", Name: "Canadian Dollar"},
		{Code: "CHF", Symbol: "CHF", Name: "Swiss Franc"},
		{Code: "CNY", Symbol: "¥", Name: "Chinese Yuan"},
		{Code: "SEK", Symbol: "kr", Name: "Swedish Krona"},
		{Code: "NZD", Symbol: "NZ$", Name: "New Zealand Dollar"},
		{Code: "IDR", Symbol: "Rp", Name: "Indonesian Rupiah"},
		{Code: "SGD", Symbol: "S$", Name: "Singapore Dollar"},
		{Code: "INR", Symbol: "₹", Name: "Indian Rupee"},
	}

	// Verify all currencies are in the service's supported list
	supportedMap := make(map[string]bool)
	for _, code := range services.SupportedCurrencies {
		supportedMap[code] = true
	}

	// Filter to only return currencies that are actually supported
	var result []SupportedCurrency
	for _, currency := range currencies {
		if supportedMap[currency.Code] {
			result = append(result, currency)
		}
	}

	return c.JSON(http.StatusOK, result)
}

// ConvertCurrencyRequest represents a currency conversion request
type ConvertCurrencyRequest struct {
	Amount   float64 `json:"amount" validate:"required"`
	From     string  `json:"from" validate:"required"`
	To       string  `json:"to" validate:"required"`
}

// ConvertCurrencyResponse represents a currency conversion response
type ConvertCurrencyResponse struct {
	Amount         float64 `json:"amount"`
	From           string  `json:"from"`
	To             string  `json:"to"`
	ConvertedAmount float64 `json:"converted_amount"`
	Rate           float64 `json:"rate"`
}

// ConvertCurrency godoc
// @Summary Convert amount between currencies
// @Description Converts an amount from one currency to another using current exchange rates
// @Tags currency
// @Accept json
// @Produce json
// @Param request query ConvertCurrencyRequest true "Conversion Request"
// @Success 200 {object} ConvertCurrencyResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/convert-currency [get]
func (h *CurrencyHandler) ConvertCurrency(c echo.Context) error {
	amountStr := c.QueryParam("amount")
	from := c.QueryParam("from")
	to := c.QueryParam("to")

	if amountStr == "" || from == "" || to == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing required parameters"})
	}

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid amount"})
	}

	convertedAmount, err := h.exchangeService.ConvertAmount(c.Request().Context(), amount, from, to)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// Calculate the rate
	rate := 1.0
	if amount > 0 {
		rate = convertedAmount / amount
	}

	return c.JSON(http.StatusOK, ConvertCurrencyResponse{
		Amount:          amount,
		From:            from,
		To:              to,
		ConvertedAmount: convertedAmount,
		Rate:            rate,
	})
}
