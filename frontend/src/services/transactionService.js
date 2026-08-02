/**
 * API calls related to transactions.
 */
import axiosClient from "../api/axiosClient";

export async function getTransactions() {
  const response = await axiosClient.get("/transactions/");
  return response.data;
}

export async function getTransactionById(transactionId) {
  const response = await axiosClient.get(`/transactions/${transactionId}`);
  return response.data;
}

export async function createTransaction(data) {
  const response = await axiosClient.post("/transactions/", data);
  return response.data;
}
