/**
 * Root component. Sets up all routing for the application.
 */
import { Routes, Route, Navigate } from "react-router-dom";
import { AuthProvider } from "./context/AuthContext";
import ProtectedRoute from "./components/ProtectedRoute";
import AdminRoute from "./components/AdminRoute";
import Layout from "./components/Layout";

import Login from "./pages/Login";
import Register from "./pages/Register";
import Dashboard from "./pages/Dashboard";
import Transactions from "./pages/Transactions";
import Refunds from "./pages/Refunds";
import AiAssistant from "./pages/AiAssistant";
import SupportTickets from "./pages/SupportTickets";

import AdminDashboard from "./pages/admin/AdminDashboard";
import AdminUsers from "./pages/admin/AdminUsers";
import AdminTransactions from "./pages/admin/AdminTransactions";
import AdminRefunds from "./pages/admin/AdminRefunds";
import AdminTickets from "./pages/admin/AdminTickets";

export default function App() {
  return (
    <AuthProvider>
      <Routes>
        <Route path="/" element={<Navigate to="/login" replace />} />
        <Route path="/login" element={<Login />} />
        <Route path="/register" element={<Register />} />

        {/* User routes */}
        <Route path="/dashboard" element={<ProtectedRoute><Layout><Dashboard /></Layout></ProtectedRoute>} />
        <Route path="/transactions" element={<ProtectedRoute><Layout><Transactions /></Layout></ProtectedRoute>} />
        <Route path="/refunds" element={<ProtectedRoute><Layout><Refunds /></Layout></ProtectedRoute>} />
        <Route path="/assistant" element={<ProtectedRoute><Layout><AiAssistant /></Layout></ProtectedRoute>} />
        <Route path="/tickets" element={<ProtectedRoute><Layout><SupportTickets /></Layout></ProtectedRoute>} />

        {/* Admin routes -- nested: ProtectedRoute -> AdminRoute -> Layout -> page */}
        <Route path="/admin" element={<ProtectedRoute><AdminRoute><Layout><AdminDashboard /></Layout></AdminRoute></ProtectedRoute>} />
        <Route path="/admin/users" element={<ProtectedRoute><AdminRoute><Layout><AdminUsers /></Layout></AdminRoute></ProtectedRoute>} />
        <Route path="/admin/transactions" element={<ProtectedRoute><AdminRoute><Layout><AdminTransactions /></Layout></AdminRoute></ProtectedRoute>} />
        <Route path="/admin/refunds" element={<ProtectedRoute><AdminRoute><Layout><AdminRefunds /></Layout></AdminRoute></ProtectedRoute>} />
        <Route path="/admin/tickets" element={<ProtectedRoute><AdminRoute><Layout><AdminTickets /></Layout></AdminRoute></ProtectedRoute>} />
      </Routes>
    </AuthProvider>
  );
}
