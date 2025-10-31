package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"worklio-api/internal/db"
)

// Supported currencies list
var SupportedCurrencies = []string{
	"USD", "EUR", "GBP", "JPY", "AUD",
	"CAD", "CHF", "CNY", "SEK", "NZD",
	"IDR", "SGD", "INR",
}

// ExchangeRateService handles currency exchange rate operations
type ExchangeRateService struct {
	queries *db.Queries
}

// NewExchangeRateService creates a new exchange rate service
func NewExchangeRateService(queries *db.Queries) *ExchangeRateService {
	return &ExchangeRateService{queries: queries}
}

// ExchangeAPIResponse represents the response from exchangerate-api.com
type ExchangeAPIResponse struct {
	Result           string             `json:"result"`
	BaseCode         string             `json:"base_code"`
	ConversionRates  map[string]float64 `json:"conversion_rates"`
	TimeLastUpdateUTC string            `json:"time_last_update_utc"`
}

// GetExchangeRate gets a single exchange rate from database or API
func (s *ExchangeRateService) GetExchangeRate(ctx context.Context, baseCurrency, targetCurrency string) (float64, error) {
	if baseCurrency == targetCurrency {
		return 1.0, nil
	}

	// Try to get from database using sqlc
	rateRow, err := s.queries.GetExchangeRate(ctx, db.GetExchangeRateParams{
		BaseCurrency:   baseCurrency,
		TargetCurrency: targetCurrency,
	})

	if err == nil {
		// Parse the rate string to float64
		rate, parseErr := strconv.ParseFloat(rateRow.Rate, 64)
		if parseErr != nil {
			return 0, fmt.Errorf("failed to parse exchange rate: %w", parseErr)
		}
		return rate, nil
	}

	if err == sql.ErrNoRows {
		// Not in database
		log.Printf("Exchange rate not found for %s -> %s, needs to be updated", baseCurrency, targetCurrency)
		return 0, fmt.Errorf("exchange rate not available, please run update job")
	}

	return 0, fmt.Errorf("failed to query exchange rate: %w", err)
}

// UpdateAllRates fetches and updates all exchange rates from the API
func (s *ExchangeRateService) UpdateAllRates(ctx context.Context) error {
	baseCurrency := "USD"

	log.Printf("Updating exchange rates for base currency: %s", baseCurrency)

	// Fetch rates from API (using free frankfurter.app - no API key needed)
	url := fmt.Sprintf("https://api.frankfurter.app/latest?from=%s", baseCurrency)

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch exchange rates: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var apiResp struct {
		Base  string             `json:"base"`
		Date  string             `json:"date"`
		Rates map[string]float64 `json:"rates"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return fmt.Errorf("failed to decode API response: %w", err)
	}

	// Add USD rate (base currency)
	apiResp.Rates["USD"] = 1.0

	// Update rates for supported currencies only using sqlc
	updatedCount := 0
	for _, targetCurrency := range SupportedCurrencies {
		rate, exists := apiResp.Rates[targetCurrency]
		if !exists {
			log.Printf("Warning: Rate not available for %s", targetCurrency)
			continue
		}

		// Upsert the rate using sqlc
		err := s.queries.UpsertExchangeRate(ctx, db.UpsertExchangeRateParams{
			BaseCurrency:   baseCurrency,
			TargetCurrency: targetCurrency,
			Rate:           fmt.Sprintf("%.10f", rate),
		})

		if err != nil {
			return fmt.Errorf("failed to update rate for %s: %w", targetCurrency, err)
		}
		updatedCount++
	}

	log.Printf("Successfully updated %d exchange rates", updatedCount)
	return nil
}

// ConvertAmount converts an amount from one currency to another
func (s *ExchangeRateService) ConvertAmount(ctx context.Context, amount float64, fromCurrency, toCurrency string) (float64, error) {
	if fromCurrency == toCurrency {
		return amount, nil
	}

	rate, err := s.GetExchangeRate(ctx, "USD", toCurrency)
	if err != nil {
		return 0, err
	}

	fromRate, err := s.GetExchangeRate(ctx, "USD", fromCurrency)
	if err != nil {
		return 0, err
	}

	// Convert: amount in fromCurrency -> USD -> toCurrency
	usdAmount := amount / fromRate
	convertedAmount := usdAmount * rate

	return convertedAmount, nil
}
