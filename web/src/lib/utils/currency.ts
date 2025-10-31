// Currency formatting utilities

export const CURRENCY_SYMBOLS: Record<string, string> = {
	USD: '$',
	EUR: '€',
	GBP: '£',
	JPY: '¥',
	AUD: 'A$',
	CAD: 'C$',
	CHF: 'CHF',
	CNY: '¥',
	SEK: 'kr',
	NZD: 'NZ$',
	IDR: 'Rp',
	SGD: 'S$',
	INR: '₹'
};

export const CURRENCY_NAMES: Record<string, string> = {
	USD: 'USD - US Dollar ($)',
	EUR: 'EUR - Euro (€)',
	GBP: 'GBP - British Pound (£)',
	JPY: 'JPY - Japanese Yen (¥)',
	AUD: 'AUD - Australian Dollar (A$)',
	CAD: 'CAD - Canadian Dollar (C$)',
	CHF: 'CHF - Swiss Franc (CHF)',
	CNY: 'CNY - Chinese Yuan (¥)',
	SEK: 'SEK - Swedish Krona (kr)',
	NZD: 'NZD - New Zealand Dollar (NZ$)',
	IDR: 'IDR - Indonesian Rupiah (Rp)',
	SGD: 'SGD - Singapore Dollar (S$)',
	INR: 'INR - Indian Rupee (₹)'
};

export function getCurrencySymbol(currency: string): string {
	return CURRENCY_SYMBOLS[currency] || '$';
}

function formatNumber(num: number, decimals: number = 2): string {
	return num.toLocaleString('en-US', {
		minimumFractionDigits: decimals,
		maximumFractionDigits: decimals
	});
}

export function formatCurrency(amount: number, currency: string = 'USD'): string {
	const symbol = getCurrencySymbol(currency);
	return `${symbol}${formatNumber(amount, 2)}`;
}

export function formatCurrencyRate(rate: number, currency: string = 'USD'): string {
	const symbol = getCurrencySymbol(currency);
	return `${symbol}${formatNumber(rate, 0)}/hr`;
}
