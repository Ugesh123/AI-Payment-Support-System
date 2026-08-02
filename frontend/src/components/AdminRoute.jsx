/**
 * Wraps admin-only pages. Reuses AuthContext's `user` (already fetched
 * via GET /auth/me on app load) to check the role -- no extra API call.
 * Redirects non-admins to /dashboard.
 */
import { Navigate } from "react-router-dom";
import { useAuth } from "../context/AuthContext";

export default function AdminRoute({ children }) {
  const { user, loading } = useAuth();

  if (loading) return null; // ProtectedRoute (parent) already shows the loader

  if (user?.role !== "admin") {
    return <Navigate to="/dashboard" replace />;
  }

  return children;
}
