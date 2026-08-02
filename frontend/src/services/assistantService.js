/**
 * API calls related to the AI Assistant.
 */
import axiosClient from "../api/axiosClient";

export async function askAssistant(transactionId, question) {
  const response = await axiosClient.post("/assistant/chat", {
    transaction_id: transactionId,
    question,
  });
  return response.data;
}
