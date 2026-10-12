import React, { useState } from 'react';
import { X, Play } from 'lucide-react';
import { Button } from './Button';
import { ToolInfo } from '../types/tool';
import { Target } from '../types/target';

interface RunConfigModalProps {
  isOpen: boolean;
  tool: ToolInfo | null;
  target: Target | null;
  onClose: () => void;
  onExecute: (toolName: string, customArgs: string[]) => void;
}

export const RunConfigModal: React.FC<RunConfigModalProps> = ({
  isOpen,
  tool,
  target,
  onClose,
  onExecute,
}) => {
  const [customArgs, setCustomArgs] = useState('');

  if (!isOpen || !tool || !target) return null;

  const handleRun = () => {
    const args = customArgs.trim() ? customArgs.trim().split(/\s+/) : [];
    onExecute(tool.name, args);
    onClose();
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80">
      <div className="bg-[#161b22] border border-[#30363d] rounded-md max-w-lg w-full p-5 relative shadow-xl">
        <button
          onClick={onClose}
          className="absolute top-4 right-4 text-zinc-400 hover:text-zinc-100"
        >
          <X className="w-4 h-4" />
        </button>

        <div className="mb-4">
          <h3 className="text-sm font-bold text-zinc-100 font-mono">Execute {tool.name}</h3>
          <p className="text-xs text-zinc-400 mt-0.5">
            Target: <span className="font-mono text-zinc-200 font-semibold">{target.domain}</span>
          </p>
        </div>

        <div className="space-y-3">
          <div className="p-2.5 bg-[#0d1117] border border-[#21262d] rounded text-xs text-zinc-400 font-mono">
            <span className="text-zinc-500 block text-[10px] uppercase font-semibold mb-1">Methodology Objective</span>
            {tool.description}
          </div>

          <div>
            <label className="block text-xs font-mono text-zinc-300 mb-1">
              Custom CLI Arguments (Optional Override)
            </label>
            <input
              type="text"
              placeholder="Leave blank to use pre-tuned defaults"
              value={customArgs}
              onChange={(e) => setCustomArgs(e.target.value)}
              className="w-full bg-[#0d1117] border border-[#30363d] rounded px-3 py-1.5 text-xs text-zinc-100 placeholder-zinc-600 focus:outline-none focus:border-zinc-500 font-mono"
            />
            <p className="text-[11px] text-zinc-500 mt-1">
              By default, tuned high-speed non-disruptive flags are supplied automatically.
            </p>
          </div>

          <div className="flex justify-end space-x-2 pt-2 border-t border-[#21262d]">
            <Button variant="ghost" onClick={onClose}>
              Cancel
            </Button>
            <Button variant="primary" onClick={handleRun}>
              <Play className="w-3.5 h-3.5 mr-1 fill-current" />
              Launch Tool
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
};
