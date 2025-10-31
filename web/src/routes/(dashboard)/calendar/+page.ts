import { api } from '$lib/api';
import type { PageLoad } from './$types';

export const ssr = false;

function calculateDateRange(weeks: number, offset: number): { startDate: string; endDate: string } {
	const now = new Date();

	// Apply offset to current date (offset in weeks)
	const offsetDate = new Date(now);
	offsetDate.setDate(now.getDate() + offset * 7);

	// Find the week start (Monday) for the offset date
	let weekday = offsetDate.getDay();
	if (weekday === 0) {
		weekday = 7; // Sunday is 0, convert to 7
	}
	const daysToMonday = weekday - 1;
	const currentWeekStart = new Date(offsetDate);
	currentWeekStart.setDate(offsetDate.getDate() - daysToMonday);

	// Calculate end date: Sunday of the current week (6 days after Monday)
	const endDate = new Date(currentWeekStart);
	endDate.setDate(currentWeekStart.getDate() + 6);

	// Calculate start date: Monday of N weeks ago
	const startDate = new Date(currentWeekStart);
	startDate.setDate(currentWeekStart.getDate() - (weeks - 1) * 7);

	// Format dates as YYYY-MM-DD in local timezone
	const formatDate = (date: Date): string => {
		const year = date.getFullYear();
		const month = String(date.getMonth() + 1).padStart(2, '0');
		const day = String(date.getDate()).padStart(2, '0');
		return `${year}-${month}-${day}`;
	};

	return {
		startDate: formatDate(startDate),
		endDate: formatDate(endDate)
	};
}

export const load: PageLoad = async ({ url, depends }) => {
	depends('heatmap');
	const weeks = parseInt(url.searchParams.get('weeks') || '4');
	const offset = parseInt(url.searchParams.get('offset') || '0');

	const { startDate, endDate } = calculateDateRange(weeks, offset);

	return {
		heatmap: api.getHeatmap(startDate, endDate),
		offset
	};
};
