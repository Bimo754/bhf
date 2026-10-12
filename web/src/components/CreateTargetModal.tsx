import React, { useState } from 'react';
import { X, Target as TargetIcon, Plus } from 'lucide-react';
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
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/70 backdrop-blur-sm">
      <div className="bg-dark-800 border border-dark-600 rounded-xl max-w-md w-full p-6 shadow-2xl relative">
        <button
          onClick={onClose}
          className="absolute top-4 right-4 text-gray-400 hover:text-gray-100"
        >
          <X className="w-5 h-5" />
        </button>

        <div className="flex items-center space-x-3 mb-6">
          <div className="p-2.5 rounded-lg bg-cyber-cyan/10 text-cyber-cyan border border-cyber-cyan/30">
            <TargetIcon className="w-6 h-6" />
          </div>
          <div>
            <h3 className="text-lg font-bold text-gray-100">Add Target Program</h3>
            <p className="text-xs text-gray-400">Define target scope domain for reconnaissance</p>
          </div>
        </div>

        {error && (
          <div className="mb-4 p-3 rounded-lg bg-cyber-rose/10 border border-cyber-rose/30 text-cyber-rose text-xs">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-xs font-medium text-gray-300 mb-1">
              Program / Target Name
            </label>
            <input
              type="text"
              placeholder="e.g. Acme Bug Bounty"
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="w-full bg-dark-900 border border-dark-600 rounded-lg px-3 py-2 text-sm text-gray-100 placeholder-gray-500 focus:outline-none focus:border-cyber-cyan"
            />
          </div>

          <div>
            <label className="block text-xs font-medium text-gray-300 mb-1">
              Root Scope Domain <span className="text-cyber-rose">*</span>
            </label>
            <input
              type="text"
              placeholder="e.g. example.com"
              value={domain}
              onChange={(e) => setDomain(e.target.value)}
              required
              className="w-full bg-dark-900 border border-dark-600 rounded-lg px-3 py-2 text-sm text-gray-100 placeholder-gray-500 focus:outline-none focus:border-cyber-cyan font-mono"
            />
          </div>

          <div>
            <label className="block text-xs font-medium text-gray-300 mb-1">
              Scope Notes & Rules (Optional)
            </label>
            <textarea
              rows={3}
              placeholder="Out of scope domains, rate limits, bounty policy notes..."
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
              className="w-full bg-dark-900 border border-dark-600 rounded-lg px-3 py-2 text-sm text-gray-100 placeholder-gray-500 focus:outline-none focus:border-cyber-cyan resize-none"
            />
          </div>

          <div className="flex justify-end space-x-3 pt-2">
            <Button variant="ghost" type="button" onClick={onClose}>
              Cancel
            </Button>
            <Button variant="primary" type="submit" isLoading={isLoading}>
              <Plus className="w-4 h-4 mr-1.5" />
              Add Target
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
};
