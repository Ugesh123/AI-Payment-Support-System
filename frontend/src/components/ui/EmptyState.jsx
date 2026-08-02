/**
 * Shown when a list/table has no data yet (e.g. no transactions).
 * Includes a simple inline SVG icon so empty screens feel intentional.
 */
export default function EmptyState({ title = "Nothing here yet", subtitle = "" }) {
  return (
    <div className="text-center py-12 text-gray-400">
      <svg
        className="mx-auto h-10 w-10 text-gray-300 mb-3"
        fill="none"
        viewBox="0 0 24 24"
        stroke="currentColor"
        aria-hidden="true"
      >
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5}
          d="M9 13h6m-3-3v6m-9 1V7a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V17a2 2 0 01-2 2H5a2 2 0 01-2-2z" />
      </svg>
      <p className="text-base font-medium text-gray-500">{title}</p>
      {subtitle && <p className="text-sm mt-1">{subtitle}</p>}
    </div>
  );
}
