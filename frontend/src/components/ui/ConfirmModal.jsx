/**
 * Reusable confirmation dialog, built on top of the existing Modal
 * component. Used before any destructive/state-changing admin action.
 */
import Modal from "./Modal";
import Button from "./Button";

export default function ConfirmModal({ isOpen, title, message, onConfirm, onCancel, loading }) {
  return (
    <Modal isOpen={isOpen} onClose={onCancel} title={title}>
      <p className="text-sm text-gray-600 mb-5">{message}</p>
      <div className="flex justify-end gap-2">
        <Button variant="outline" onClick={onCancel} disabled={loading}>
          Cancel
        </Button>
        <Button variant="primary" onClick={onConfirm} loading={loading}>
          Confirm
        </Button>
      </div>
    </Modal>
  );
}
