/**
 * Shared formatting helpers so currency/date formatting stays
 * consistent across every page instead of being repeated inline.
 */

export function formatCurrency(amount) {
  const value = Number(amount);
  return `₹${value.toLocaleString("en-IN", { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
}

export function formatDateTime(isoString) {
  if (!isoString) return "-";
  const date = new Date(isoString);
  return date.toLocaleString("en-IN", {
    year: "numeric",
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}
