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
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80">
      <div className="bg-[#161b22] border border-[#30363d] rounded-md max-w-md w-full p-5 relative shadow-xl">
        <button
          onClick={onClose}
          className="absolute top-4 right-4 text-zinc-400 hover:text-zinc-100"
        >
          <X className="w-4 h-4" />
        </button>

        <div className="flex items-center space-x-2.5 mb-3 text-red-400">
          <Trash2 className="w-4 h-4 text-red-400" />
          <h3 className="text-sm font-bold text-zinc-100 font-mono">Delete Target Scope</h3>
        </div>

        <p className="text-xs text-zinc-300 leading-relaxed mb-3">
          Are you sure you want to delete <strong className="font-mono text-zinc-100">{target.domain}</strong>?
        </p>

        <p className="text-[11px] text-zinc-500 mb-4 bg-[#0d1117] p-2.5 rounded border border-[#21262d]">
          This action will permanently remove this target program, along with all its recorded scans, logs, and discovered assets.
        </p>

        <div className="flex justify-end space-x-2 pt-2 border-t border-[#21262d]">
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
