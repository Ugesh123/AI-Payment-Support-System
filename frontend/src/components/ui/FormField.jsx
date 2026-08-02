/**
 * Wraps a label + input/select/textarea + optional inline error,
 * for consistent field-level validation messaging in forms.
 */
export default function FormField({ label, htmlFor, error, children }) {
  return (
    <div className="mb-3">
      <label htmlFor={htmlFor} className="block text-sm text-gray-600 mb-1">
        {label}
      </label>
      {children}
      {error && <p className="text-xs text-red-500 mt-1">{error}</p>}
    </div>
  );
}
