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

function parseCommandLine(cmd: string): string[] {
  const args: string[] = [];
  let current = '';
  let inQuotes = false;
  let quoteChar = '';

  for (let i = 0; i < cmd.length; i++) {
    const char = cmd[i];
    if ((char === '"' || char === "'") && (!inQuotes || quoteChar === char)) {
      inQuotes = !inQuotes;
      if (inQuotes) {
        quoteChar = char;
      } else {
        quoteChar = '';
      }
    } else if ((char === ' ' || char === '\t') && !inQuotes) {
      if (current.length > 0) {
        args.push(current);
        current = '';
      }
    } else {
      current += char;
    }
  }
  if (current.length > 0) {
    args.push(current);
  }
  return args;
}

export const RunConfigModal: React.FC<RunConfigModalProps> = ({
  isOpen,
  tool,
  target,
  onClose,
  onExecute,
}) => {
  const [customArgs, setCustomArgs] = useState('');
  const [customHeaders, setCustomHeaders] = useState('');

  if (!isOpen || !tool || !target) return null;

  const handleRun = () => {
    let finalArgs = customArgs.trim() ? parseCommandLine(customArgs.trim()) : [];

    // Ffuf Custom Header Injection
    if (tool.name === 'ffuf' && customHeaders.trim()) {
      const headerLines = customHeaders.split('\n').map((h) => h.trim()).filter(Boolean);
      for (const h of headerLines) {
        finalArgs.push('-H', h);
      }
    }

    onExecute(tool.name, finalArgs);
    onClose();
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm animate-in fade-in duration-150">
      <div className="bg-[#0e1420] border border-white/[0.08] rounded-2xl max-w-lg w-full p-6 relative shadow-2xl font-mono space-y-4">
        <button
          onClick={onClose}
          className="absolute top-5 right-5 text-slate-400 hover:text-white transition-colors"
        >
          <X className="w-4 h-4" />
        </button>

        <div>
          <h3 className="text-sm font-bold text-slate-100">Configure & Execute {tool.name}</h3>
          <p className="text-xs text-slate-400 mt-0.5">
            Target Scope: <span className="text-slate-200 font-semibold">{target.domain}</span>
          </p>
        </div>

        <div className="space-y-3.5">
          <div className="p-3 bg-white/[0.02] border border-white/[0.04] rounded-xl text-xs text-slate-300">
            <span className="text-slate-500 block text-[10px] uppercase font-semibold mb-1">Methodology Objective</span>
            {tool.description}
          </div>

          {tool.name === 'ffuf' && (
            <div>
              <label className="block text-xs font-medium text-slate-300 mb-1">
                Custom Headers & Cookies
              </label>
              <textarea
                rows={2}
                placeholder="Cookie: session=xyz&#10;Authorization: Bearer token123"
                value={customHeaders}
                onChange={(e) => setCustomHeaders(e.target.value)}
                className="w-full bg-white/[0.03] border border-white/[0.08] rounded-lg px-3 py-2 text-xs text-slate-100 placeholder-slate-600 focus:outline-none focus:border-indigo-500"
              />
              <p className="text-[11px] text-slate-500 mt-1">
                Each line will be injected via -H into the fuzzing requests.
              </p>
            </div>
          )}

          <div>
            <label className="block text-xs font-medium text-slate-300 mb-1">
              Custom CLI Arguments (Optional Override)
            </label>
            <input
              type="text"
              placeholder="Leave blank to use pre-tuned defaults"
              value={customArgs}
              onChange={(e) => setCustomArgs(e.target.value)}
              className="w-full bg-white/[0.03] border border-white/[0.08] rounded-lg px-3 py-2 text-xs text-slate-100 placeholder-slate-600 focus:outline-none focus:border-indigo-500"
            />
            <p className="text-[11px] text-slate-500 mt-1">
              Pre-tuned high-speed non-disruptive flags are supplied automatically.
            </p>
          </div>

          <div className="flex justify-end space-x-2 pt-3 border-t border-white/[0.04]">
            <Button variant="ghost" size="sm" onClick={onClose}>
              Cancel
            </Button>
            <Button variant="primary" size="sm" onClick={handleRun}>
              <Play className="w-3.5 h-3.5 mr-1 fill-current" />
              Launch Tool
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
};
