/**
 * Small colored label for statuses (SUCCESS, FAILED, PENDING, etc.)
 * Centralizing the color logic here means every page shows the
 * same status the same way.
 */
const STATUS_STYLES = {
  SUCCESS: "bg-green-100 text-green-700",
  COMPLETED: "bg-green-100 text-green-700",
  FAILED: "bg-red-100 text-red-700",
  PENDING: "bg-yellow-100 text-yellow-700",
  REQUESTED: "bg-yellow-100 text-yellow-700",
  PROCESSING: "bg-blue-100 text-blue-700",
  OPEN: "bg-yellow-100 text-yellow-700",
  IN_PROGRESS: "bg-blue-100 text-blue-700",
  CLOSED: "bg-gray-200 text-gray-700",
};

export default function Badge({ status }) {
  const style = STATUS_STYLES[status] || "bg-gray-100 text-gray-600";
  return (
    <span className={`text-xs font-medium px-2.5 py-1 rounded-full ${style}`}>
      {status}
    </span>
  );
}
