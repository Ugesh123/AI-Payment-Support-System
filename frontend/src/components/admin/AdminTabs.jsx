/**
 * Secondary navigation shown only within admin pages.
 * Kept separate from the main Layout navbar so regular users never
 * see admin-only tabs.
 */
import { NavLink } from "react-router-dom";

const ADMIN_TABS = [
  { to: "/admin", label: "Overview", end: true },
  { to: "/admin/users", label: "Users" },
  { to: "/admin/transactions", label: "Transactions" },
  { to: "/admin/refunds", label: "Refunds" },
  { to: "/admin/tickets", label: "Tickets" },
];

export default function AdminTabs() {
  const tabClass = ({ isActive }) =>
    `px-3 py-2 text-sm rounded transition ${
      isActive ? "bg-blue-600 text-white" : "text-gray-600 hover:bg-gray-100"
    }`;

  return (
    <div className="flex flex-wrap gap-2 mb-5">
      {ADMIN_TABS.map((tab) => (
        <NavLink key={tab.to} to={tab.to} end={tab.end} className={tabClass}>
          {tab.label}
        </NavLink>
      ))}
    </div>
  );
}
