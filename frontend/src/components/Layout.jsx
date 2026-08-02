/**
 * Shared layout for all logged-in pages.
 * Uses NavLink so the active page is automatically highlighted.
 * Includes a mobile hamburger menu since the full nav (5 links +
 * admin + username + logout) doesn't fit on small screens.
 */
import { useState } from "react";
import { NavLink, useNavigate } from "react-router-dom";
import { useAuth } from "../context/AuthContext";

const NAV_LINKS = [
  { to: "/dashboard", label: "Dashboard" },
  { to: "/transactions", label: "Transactions" },
  { to: "/refunds", label: "Refunds" },
  { to: "/tickets", label: "Support" },
  { to: "/assistant", label: "AI Assistant" },
];

export default function Layout({ children }) {
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  const [menuOpen, setMenuOpen] = useState(false);

  const handleLogout = () => {
    logout();
    navigate("/login");
  };

  const linkClass = ({ isActive }) =>
    `text-sm transition ${isActive ? "text-blue-600 font-medium" : "text-gray-600 hover:text-blue-600"}`;

  const mobileLinkClass = ({ isActive }) =>
    `block px-2 py-2 text-sm rounded ${isActive ? "bg-blue-50 text-blue-600 font-medium" : "text-gray-600 hover:bg-gray-50"}`;

  return (
    <div className="min-h-screen bg-gray-50">
      <nav className="bg-white border-b shadow-sm">
        <div className="max-w-6xl mx-auto px-4 py-3 flex items-center justify-between">
          <span className="font-semibold text-gray-800">AI Payment Support</span>

          <div className="hidden md:flex items-center gap-5">
            {NAV_LINKS.map((link) => (
              <NavLink key={link.to} to={link.to} className={linkClass}>{link.label}</NavLink>
            ))}
            {user?.role === "admin" && (
              <NavLink to="/admin" className={linkClass}>Admin</NavLink>
            )}
          </div>

          <div className="hidden md:flex items-center gap-3">
            <span className="text-sm text-gray-500">{user?.name}</span>
            <button
              onClick={handleLogout}
              className="text-sm bg-red-50 text-red-600 px-3 py-1 rounded hover:bg-red-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-red-300 transition"
            >
              Logout
            </button>
          </div>

          <button
            className="md:hidden text-gray-600 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-gray-300 rounded p-1"
            onClick={() => setMenuOpen((prev) => !prev)}
            aria-label="Toggle menu"
            aria-expanded={menuOpen}
          >
            <svg className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" />
            </svg>
          </button>
        </div>

        {menuOpen && (
          <div className="md:hidden border-t px-4 py-3 space-y-1">
            {NAV_LINKS.map((link) => (
              <NavLink key={link.to} to={link.to} className={mobileLinkClass} onClick={() => setMenuOpen(false)}>
                {link.label}
              </NavLink>
            ))}
            {user?.role === "admin" && (
              <NavLink to="/admin" className={mobileLinkClass} onClick={() => setMenuOpen(false)}>
                Admin
              </NavLink>
            )}
            <div className="flex items-center justify-between pt-2 mt-2 border-t">
              <span className="text-sm text-gray-500">{user?.name}</span>
              <button
                onClick={handleLogout}
                className="text-sm bg-red-50 text-red-600 px-3 py-1 rounded hover:bg-red-100 transition"
              >
                Logout
              </button>
            </div>
          </div>
        )}
      </nav>

      <main className="max-w-6xl mx-auto px-4 py-6">{children}</main>
    </div>
  );
}
