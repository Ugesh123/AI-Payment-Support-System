/**
 * Reusable inline error banner for failed API calls.
 */
export default function ErrorMessage({ message = "Something went wrong." }) {
  return (
    <div className="bg-red-50 text-red-600 text-sm px-4 py-3 rounded border border-red-100">
      {message}
    </div>
  );
}
