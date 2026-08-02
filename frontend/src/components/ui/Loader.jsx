/**
 * Simple centered loading spinner/text, reused wherever data is being fetched.
 */
export default function Loader({ label = "Loading..." }) {
  return (
    <div className="flex items-center justify-center py-10 text-gray-500 text-sm">
      <div className="animate-spin h-5 w-5 border-2 border-blue-600 border-t-transparent rounded-full mr-2" />
      {label}
    </div>
  );
}
