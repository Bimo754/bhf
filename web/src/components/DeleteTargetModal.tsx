import React, { useState } from 'react';
import { X, Trash2 } from 'lucide-react';
import { Target } from '../types/target';
import { Button } from './Button';

interface DeleteTargetModalProps {
  isOpen: boolean;
  target: Target | null;
  onClose: () => void;
  onConfirmDelete: (target: Target) => Promise<void>;
}

export const DeleteTargetModal: React.FC<DeleteTargetModalProps> = ({
  isOpen,
  target,
  onClose,
  onConfirmDelete,
}) => {
  const [isDeleting, setIsDeleting] = useState(false);

  if (!isOpen || !target) return null;

  const handleDelete = async () => {
    try {
      setIsDeleting(true);
      await onConfirmDelete(target);
      onClose();
    } catch (err: any) {
      alert(`Failed to delete target: ${err.message}`);
    } finally {
      setIsDeleting(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm animate-in fade-in duration-150">
      <div className="bg-[#0e1420] border border-white/[0.08] rounded-2xl max-w-md w-full p-6 relative shadow-2xl space-y-4">
        <button
          onClick={onClose}
          className="absolute top-5 right-5 text-slate-400 hover:text-white transition-colors"
        >
          <X className="w-4 h-4" />
        </button>

        <div className="flex items-center space-x-2.5 text-rose-400">
          <Trash2 className="w-4 h-4 text-rose-400" />
          <h3 className="text-sm font-bold text-slate-100 font-mono">Delete Target Scope</h3>
        </div>

        <p className="text-xs text-slate-300 leading-relaxed">
          Are you sure you want to delete <strong className="font-mono text-slate-100">{target.domain}</strong>?
        </p>

        <p className="text-[11px] text-slate-400 bg-white/[0.02] p-3 rounded-xl border border-white/[0.04]">
          This action permanently removes this target scope, along with its recorded scans, logs, traffic, and discovered assets.
        </p>

        <div className="flex justify-end space-x-2 pt-3 border-t border-white/[0.04]">
          <Button variant="ghost" size="sm" onClick={onClose} disabled={isDeleting}>
            Cancel
          </Button>
          <Button variant="danger" size="sm" onClick={handleDelete} isLoading={isDeleting}>
            Delete Target
          </Button>
        </div>
      </div>
    </div>
  );
};
