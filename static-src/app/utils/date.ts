/**
 * Safe date formatting utilities.
 *
 * Use these helpers instead of raw Moment to avoid "Invalid date" strings
 * being silently passed through the system.
 */
import Moment from "moment";

/**
 * Safely format a date string. Returns empty string for empty, null, undefined, or invalid dates.
 */
export function formatDate(d: string | null | undefined, format: string = 'YYYY-MM-DD'): string {
  if (d == null || d === '') return '';
  const m = Moment(d);
  return m.isValid() ? m.format(format) : '';
}

/**
 * Get relative time (e.g., "2 hours ago") or empty string if invalid.
 */
export function fromNow(d: string | null | undefined): string {
  if (d == null || d === '') return '';
  const m = Moment(d);
  return m.isValid() ? m.fromNow() : '';
}

/**
 * Get a date N days from today.
 * Returns empty string if moment creation fails (should not happen in practice).
 */
export function daysFromNow(days: number, format: string = 'YYYY-MM-DD'): string {
  const m = Moment().add(days, 'days');
  return m.isValid() ? m.format(format) : '';
}

/**
 * Get today's date as a formatted string.
 */
export function today(format: string = 'YYYY-MM-DD'): string {
  return Moment().format(format);
}

/**
 * Get the previous month's year-month string (e.g., "2026-03").
 */
export function prevMonth(format: string = 'YYYY-MM'): string {
  return Moment().subtract(1, 'months').format(format);
}

/**
 * Calculate hours between two time strings (HH:mm format).
 * Returns the difference in hours as a decimal, or null if invalid.
 */
export function timeDiffHours(start: string, stop: string): number | null {
  const startM = Moment(start, 'HH:mm');
  const stopM = Moment(stop, 'HH:mm');
  if (!startM.isValid() || !stopM.isValid()) return null;
  return stopM.diff(startM) / 1000 / 60 / 60;
}
