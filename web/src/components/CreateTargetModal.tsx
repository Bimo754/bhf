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
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80">
      <div className="bg-[#161b22] border border-[#30363d] rounded-md max-w-md w-full p-5 relative shadow-xl">
        <button
          onClick={onClose}
          className="absolute top-4 right-4 text-zinc-400 hover:text-zinc-100"
        >
          <X className="w-4 h-4" />
        </button>

        <div className="mb-4">
          <h3 className="text-sm font-bold text-zinc-100 font-mono">Add Target Scope</h3>
          <p className="text-xs text-zinc-400 mt-0.5">Define a root domain or scope program for recon</p>
        </div>

        {error && (
          <div className="mb-3 p-2 rounded bg-red-950/70 border border-red-800 text-red-300 text-xs font-mono">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-3">
          <div>
            <label className="block text-xs font-mono text-zinc-300 mb-1">
              Target / Program Name
            </label>
            <input
              type="text"
              placeholder="e.g. HackerOne - Target Corp"
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="w-full bg-[#0d1117] border border-[#30363d] rounded px-3 py-1.5 text-xs text-zinc-100 placeholder-zinc-600 focus:outline-none focus:border-zinc-500 font-mono"
            />
          </div>

          <div>
            <label className="block text-xs font-mono text-zinc-300 mb-1">
              Root Domain <span className="text-red-400">*</span>
            </label>
            <input
              type="text"
              placeholder="e.g. target.com"
              value={domain}
              onChange={(e) => setDomain(e.target.value)}
              required
              className="w-full bg-[#0d1117] border border-[#30363d] rounded px-3 py-1.5 text-xs text-zinc-100 placeholder-zinc-600 focus:outline-none focus:border-zinc-500 font-mono"
            />
          </div>

          <div>
            <label className="block text-xs font-mono text-zinc-300 mb-1">
              Scope Notes (Optional)
            </label>
            <textarea
              rows={3}
              placeholder="Out-of-scope subdomains, bounty rules, rate limits..."
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
              className="w-full bg-[#0d1117] border border-[#30363d] rounded px-3 py-1.5 text-xs text-zinc-100 placeholder-zinc-600 focus:outline-none focus:border-zinc-500 resize-none font-mono"
            />
          </div>

          <div className="flex justify-end space-x-2 pt-2 border-t border-[#21262d]">
            <Button variant="ghost" type="button" onClick={onClose}>
              Cancel
            </Button>
            <Button variant="primary" type="submit" isLoading={isLoading}>
              Add Target
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
};
