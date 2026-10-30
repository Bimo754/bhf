import React, { useEffect } from 'react';
import { CheckCircle2, AlertCircle, Info, X } from 'lucide-react';

export interface ToastMessage {
  id: string;
  type: 'success' | 'error' | 'info';
  title: string;
  message?: string;
}

interface ToastContainerProps {
  toasts: ToastMessage[];
  onDismiss: (id: string) => void;
}

export const ToastContainer: React.FC<ToastContainerProps> = ({ toasts, onDismiss }) => {
  return (
    <div className="fixed bottom-5 right-5 z-50 flex flex-col space-y-2 pointer-events-none max-w-md w-full">
      {toasts.map((toast) => (
        <ToastItem key={toast.id} toast={toast} onDismiss={onDismiss} />
      ))}
    </div>
  );
};

const ToastItem: React.FC<{ toast: ToastMessage; onDismiss: (id: string) => void }> = ({
  toast,
  onDismiss,
}) => {
  useEffect(() => {
    const timer = setTimeout(() => {
      onDismiss(toast.id);
    }, 4500);
    return () => clearTimeout(timer);
  }, [toast.id, onDismiss]);

  return (
    <div className="pointer-events-auto flex items-start space-x-3 bg-[#161b22] border border-[#30363d] rounded-md p-3 shadow-xl shadow-black/50 text-xs font-mono animate-in slide-in-from-bottom-2 fade-in duration-200">
      <div className="shrink-0 mt-0.5">
        {toast.type === 'success' && <CheckCircle2 className="w-4 h-4 text-emerald-400" />}
        {toast.type === 'error' && <AlertCircle className="w-4 h-4 text-rose-400" />}
        {toast.type === 'info' && <Info className="w-4 h-4 text-blue-400" />}
      </div>

      <div className="flex-1 min-w-0">
        <h5 className="font-semibold text-zinc-100">{toast.title}</h5>
        {toast.message && (
          <p className="text-zinc-400 text-[11px] mt-0.5 break-words whitespace-pre-wrap max-h-48 overflow-y-auto pr-1">
            {toast.message}
          </p>
        )}
      </div>

      <button
        onClick={() => onDismiss(toast.id)}
        className="text-zinc-500 hover:text-zinc-300 p-0.5 rounded transition"
      >
        <X className="w-3.5 h-3.5" />
      </button>
    </div>
  );
};
