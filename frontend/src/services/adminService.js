/**
 * API calls for the Admin Panel.
 * All requests hit routes protected by get_current_admin() on the
 * backend -- if the logged-in user isn't an admin, these calls
 * return 403.
 */
import axiosClient from "../api/axiosClient";

export async function getAdminDashboard() {
  const response = await axiosClient.get("/admin/dashboard");
  return response.data;
}

export async function getAllUsers() {
  const response = await axiosClient.get("/admin/users");
  return response.data;
}

export async function getAllTransactions() {
  const response = await axiosClient.get("/admin/transactions");
  return response.data;
}

export async function updateTransactionStatus(transactionId, status) {
  const response = await axiosClient.patch(`/admin/transactions/${transactionId}/status`, { status });
  return response.data;
}

export async function getAllRefunds() {
  const response = await axiosClient.get("/admin/refunds");
  return response.data;
}

export async function updateRefundStatus(refundId, status) {
  const response = await axiosClient.patch(`/refunds/admin/${refundId}/status`, { status });
  return response.data;
}

export async function getAllTickets() {
  const response = await axiosClient.get("/admin/tickets");
  return response.data;
}

export async function updateTicketStatus(ticketId, status) {
  const response = await axiosClient.patch(`/tickets/admin/${ticketId}/status`, { status });
  return response.data;
}
