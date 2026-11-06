import React, { useState } from 'react';
import { X } from 'lucide-react';
import { Button } from './Button';
import { targetService, CreateTargetPayload } from '../services/targetService';
import { Target } from '../types/target';

interface CreateTargetModalProps {
  isOpen: boolean;
  onClose: () => void;
  onCreated: (target: Target) => void;
}

export const CreateTargetModal: React.FC<CreateTargetModalProps> = ({
  isOpen,
  onClose,
  onCreated,
}) => {
  const [name, setName] = useState('');
  const [domain, setDomain] = useState('');
  const [notes, setNotes] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState('');

  if (!isOpen) return null;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!domain.trim()) {
      setError('Root domain is required');
      return;
    }

    try {
      setIsLoading(true);
      setError('');
      const payload: CreateTargetPayload = {
        name: name.trim() || domain.trim(),
        domain: domain.trim(),
        notes: notes.trim(),
      };
      const created = await targetService.create(payload);
      onCreated(created);
      onClose();
    } catch (err: any) {
      setError(err.message || 'Failed to create target');
    } finally {
      setIsLoading(false);
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

        <div>
          <h3 className="text-sm font-bold text-slate-100 font-mono">Register Target Scope</h3>
          <p className="text-xs text-slate-400 mt-0.5">Define a root domain or IP host for security assessment</p>
        </div>

        {error && (
          <div className="p-2.5 rounded-lg bg-rose-500/10 border border-rose-500/20 text-rose-300 text-xs font-mono">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-3.5">
          <div>
            <label className="block text-xs font-mono text-slate-300 mb-1">
              Scope Name / Program
            </label>
            <input
              type="text"
              placeholder="e.g. Target Corp Security Scope"
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="w-full bg-white/[0.03] border border-white/[0.08] rounded-lg px-3 py-2 text-xs font-mono text-slate-100 placeholder-slate-600 focus:outline-none focus:border-indigo-500"
            />
          </div>

          <div>
            <label className="block text-xs font-mono text-slate-300 mb-1">
              Target Host or Root Domain <span className="text-rose-400">*</span>
            </label>
            <input
              type="text"
              placeholder="e.g. example.com or 192.168.1.100"
              value={domain}
              onChange={(e) => setDomain(e.target.value)}
              className="w-full bg-white/[0.03] border border-white/[0.08] rounded-lg px-3 py-2 text-xs font-mono text-slate-100 placeholder-slate-600 focus:outline-none focus:border-indigo-500"
              required
            />
          </div>

          <div>
            <label className="block text-xs font-mono text-slate-300 mb-1">
              Scope Notes (Optional)
            </label>
            <textarea
              rows={2}
              placeholder="Out-of-scope boundaries, rules of engagement, etc."
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
              className="w-full bg-white/[0.03] border border-white/[0.08] rounded-lg px-3 py-2 text-xs font-mono text-slate-100 placeholder-slate-600 focus:outline-none focus:border-indigo-500"
            />
          </div>

          <div className="flex justify-end space-x-2 pt-3 border-t border-white/[0.04]">
            <Button variant="ghost" size="sm" type="button" onClick={onClose} disabled={isLoading}>
              Cancel
            </Button>
            <Button variant="primary" size="sm" type="submit" isLoading={isLoading}>
              Save Target Scope
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
};
