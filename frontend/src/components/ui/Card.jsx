/**
 * Generic white card container. flex-col + h-full lets cards used
 * in a grid (e.g. dashboard summary cards) stretch to equal height.
 */
export default function Card({ children, className = "" }) {
  return (
    <div className={`bg-white rounded-lg shadow-sm border border-gray-100 p-5 flex flex-col h-full ${className}`}>
      {children}
    </div>
  );
}
