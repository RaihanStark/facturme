package utils

import (
	"fmt"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var currencySymbols = map[string]string{
	"USD": "$",
	"EUR": "€",
	"GBP": "£",
	"JPY": "¥",
	"AUD": "A$",
	"CAD": "C$",
	"CHF": "CHF",
	"CNY": "¥",
	"SEK": "kr",
	"NZD": "NZ$",
	"IDR": "Rp",
	"SGD": "S$",
	"INR": "₹",
}

// GetCurrencySymbol returns the symbol for a given currency code
func GetCurrencySymbol(currency string) string {
	if symbol, ok := currencySymbols[currency]; ok {
		return symbol
	}
	return "$" // Default to USD
}

// FormatCurrency formats an amount with currency symbol and thousand separators
func FormatCurrency(amount float64, currency string) string {
	symbol := GetCurrencySymbol(currency)
	p := message.NewPrinter(language.English)
	return fmt.Sprintf("%s%s", symbol, p.Sprintf("%.2f", amount))
}

// FormatCurrencyRate formats an hourly rate with currency symbol
func FormatCurrencyRate(rate float64, currency string) string {
	symbol := GetCurrencySymbol(currency)
	p := message.NewPrinter(language.English)
	return fmt.Sprintf("%s%s", symbol, p.Sprintf("%.0f", rate))
}

// FormatNumber formats a number with thousand separators
func FormatNumber(num float64, decimals int) string {
	p := message.NewPrinter(language.English)
	format := fmt.Sprintf("%%.%df", decimals)
	return p.Sprintf(format, num)
}

// FormatCurrencyForPDF formats an amount with ASCII-safe currency code for PDF generation
// Uses currency codes (EUR, GBP, USD) instead of symbols to avoid UTF-8 issues in PDFs
func FormatCurrencyForPDF(amount float64, currency string) string {
	p := message.NewPrinter(language.English)
	formattedAmount := p.Sprintf("%.2f", amount)
	return fmt.Sprintf("%s %s", currency, formattedAmount)
}

// FormatCurrencyRateForPDF formats an hourly rate with ASCII-safe currency code for PDF generation
func FormatCurrencyRateForPDF(rate float64, currency string) string {
	p := message.NewPrinter(language.English)
	formattedRate := p.Sprintf("%.0f", rate)
	return fmt.Sprintf("%s %s", currency, formattedRate)
}
