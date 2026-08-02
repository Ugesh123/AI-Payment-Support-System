/**
 * Shared logic for "click button -> confirm in modal -> call API ->
 * toast -> reload", extracted after noticing the same pattern
 * duplicated across the admin Transactions/Refunds/Tickets pages.
 * Available for reuse; existing pages keep their own inline
 * implementation unless explicitly refactored.
 */
import { useState } from "react";
import { useToast } from "../context/ToastContext";

export function useConfirmedAction(actionFn, onDone) {
  const [pendingAction, setPendingAction] = useState(null); // { id, payload } or null
  const [updatingId, setUpdatingId] = useState(null);
  const { showToast } = useToast();

  const requestAction = (id, payload) => setPendingAction({ id, payload });
  const cancelAction = () => setPendingAction(null);

  const confirmAction = async (successMessage) => {
    if (!pendingAction) return;
    const { id, payload } = pendingAction;
    setUpdatingId(id);
    try {
      await actionFn(id, payload);
      showToast(typeof successMessage === "function" ? successMessage(payload) : successMessage);
      await onDone();
    } catch (err) {
      showToast(err.response?.data?.detail || "Update failed.", "error");
    } finally {
      setUpdatingId(null);
      setPendingAction(null);
    }
  };

  return { pendingAction, updatingId, requestAction, cancelAction, confirmAction };
}
