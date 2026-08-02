/**
 * API calls related to the user dashboard.
 */
import axiosClient from "../api/axiosClient";

export async function getDashboard() {
  const response = await axiosClient.get("/dashboard/");
  return response.data;
}
