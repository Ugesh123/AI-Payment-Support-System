/**
 * API calls related to support tickets.
 */
import axiosClient from "../api/axiosClient";

export async function getMyTickets() {
  const response = await axiosClient.get("/tickets/");
  return response.data;
}

export async function createTicket(data) {
  const response = await axiosClient.post("/tickets/", data);
  return response.data;
}

export async function getTicketById(ticketId) {
  const response = await axiosClient.get(`/tickets/${ticketId}`);
  return response.data;
}
