import { toastStore } from '../../stores/toastStore';
import { X } from 'lucide-react';
import './Toast.scss';

export function ToastContainer() {
  const toasts = toastStore((s) => s.toasts);
  const dismiss = toastStore((s) => s.dismiss);

  if (toasts.length === 0) return null;

  return (
    <div className="toast-container">
      {toasts.map((toast) => (
        <div key={toast.id} className={`toast-item toast-item--${toast.type}`}>
          <span>{toast.message}</span>
          <button
            onClick={() => dismiss(toast.id)}
            className="toast-item__close"
          >
            <X size={14} />
          </button>
        </div>
      ))}
    </div>
  );
}
