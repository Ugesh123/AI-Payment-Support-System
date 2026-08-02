/**
 * Reusable button with consistent variants.
 * Includes a visible focus ring for keyboard navigation and
 * aria-busy during loading.
 */
const VARIANTS = {
  primary: "bg-blue-600 text-white hover:bg-blue-700 focus-visible:ring-blue-400",
  danger: "bg-red-600 text-white hover:bg-red-700 focus-visible:ring-red-400",
  secondary: "bg-gray-100 text-gray-700 hover:bg-gray-200 focus-visible:ring-gray-400",
  outline: "border border-gray-300 text-gray-700 hover:bg-gray-50 focus-visible:ring-gray-400",
};

export default function Button({
  children,
  variant = "primary",
  disabled = false,
  loading = false,
  type = "button",
  onClick,
  className = "",
}) {
  return (
    <button
      type={type}
      onClick={onClick}
      disabled={disabled || loading}
      aria-busy={loading}
      className={`px-4 py-2 rounded text-sm font-medium transition disabled:opacity-50 disabled:cursor-not-allowed focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-1 ${VARIANTS[variant]} ${className}`}
    >
      {loading ? "Please wait..." : children}
    </button>
  );
}
