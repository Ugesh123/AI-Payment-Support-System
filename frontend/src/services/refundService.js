/**
 * API calls related to refunds.
 */
import axiosClient from "../api/axiosClient";

export async function requestRefund(data) {
  const response = await axiosClient.post("/refunds/", data);
  return response.data;
}

export async function getRefundStatus(transactionId) {
  const response = await axiosClient.get(`/refunds/${transactionId}`);
  return response.data;
}
