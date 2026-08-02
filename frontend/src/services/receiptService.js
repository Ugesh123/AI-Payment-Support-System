/**
 * API call for downloading a transaction receipt.
 * Uses responseType: "blob" since the backend returns raw PDF bytes.
 */
import axiosClient from "../api/axiosClient";

export async function downloadReceipt(transactionId) {
  const response = await axiosClient.get(`/receipt/${transactionId}`, {
    responseType: "blob",
  });
  return response.data;
}
