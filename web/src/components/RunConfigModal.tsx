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

// Safely parses command-line string respecting single and double quotes
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
  const [nucleiSeverity, setNucleiSeverity] = useState('high,critical');
  const [nucleiTagPreset, setNucleiTagPreset] = useState('cve,exposure,vuln');

  if (!isOpen || !tool || !target) return null;

  const handleRun = () => {
    let finalArgs = customArgs.trim() ? parseCommandLine(customArgs.trim()) : [];

    // Ffuf Custom Header Injection (FEAT-002)
    if (tool.name === 'ffuf' && customHeaders.trim()) {
      const headerLines = customHeaders.split('\n').map((h) => h.trim()).filter(Boolean);
      for (const h of headerLines) {
        finalArgs.push('-H', h);
      }
    }

    // Nuclei Preset Scoping (BUG-002 Optimization)
    if (tool.name === 'nuclei' && !customArgs.trim()) {
      finalArgs = [
        '-s', nucleiSeverity,
        '-tags', nucleiTagPreset,
        '-c', '50',
        '-rl', '300',
        '-timeout', '5',
        '-retries', '1',
        '-mhe', '3',
        '-stats',
        '-si', '10',
      ];
    }

    onExecute(tool.name, finalArgs);
    onClose();
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80">
      <div className="bg-[#161b22] border border-[#30363d] rounded-md max-w-lg w-full p-5 relative shadow-xl font-mono">
        <button
          onClick={onClose}
          className="absolute top-4 right-4 text-zinc-400 hover:text-zinc-100"
        >
          <X className="w-4 h-4" />
        </button>

        <div className="mb-4">
          <h3 className="text-sm font-bold text-zinc-100">Execute {tool.name}</h3>
          <p className="text-xs text-zinc-400 mt-0.5">
            Target: <span className="text-zinc-200 font-semibold">{target.domain}</span>
          </p>
        </div>

        <div className="space-y-3">
          <div className="p-2.5 bg-[#0d1117] border border-[#21262d] rounded text-xs text-zinc-400">
            <span className="text-zinc-500 block text-[10px] uppercase font-semibold mb-1">Methodology Objective</span>
            {tool.description}
          </div>

          {/* Ffuf Custom Header Injection (FEAT-002) */}
          {tool.name === 'ffuf' && (
            <div>
              <label className="block text-xs font-semibold text-zinc-300 mb-1">
                Custom Headers & Cookies (Authentication)
              </label>
              <textarea
                rows={2}
                placeholder="Cookie: session=xyz&#10;Authorization: Bearer token123"
                value={customHeaders}
                onChange={(e) => setCustomHeaders(e.target.value)}
                className="w-full bg-[#0d1117] border border-[#30363d] rounded px-3 py-1.5 text-xs text-zinc-100 placeholder-zinc-600 focus:outline-none focus:border-zinc-500"
              />
              <p className="text-[11px] text-zinc-500 mt-0.5">
                Each line will be injected via -H into the fuzzing requests.
              </p>
            </div>
          )}

          {/* Nuclei Template & Severity Presets (BUG-002) */}
          {tool.name === 'nuclei' && (
            <div className="space-y-2 p-2.5 bg-[#0d1117] border border-[#21262d] rounded text-xs">
              <div>
                <span className="text-[10px] uppercase text-zinc-400 font-semibold block mb-1">Severity Preset</span>
                <div className="flex flex-wrap gap-1.5">
                  {[
                    { id: 'critical,high', label: 'High & Critical' },
                    { id: 'medium,high,critical', label: 'Medium+' },
                    { id: 'low,medium,high,critical', label: 'Low+' },
                  ].map((p) => (
                    <button
                      key={p.id}
                      type="button"
                      onClick={() => setNucleiSeverity(p.id)}
                      className={`px-2 py-0.5 rounded text-[11px] border transition ${
                        nucleiSeverity === p.id
                          ? 'bg-[#21262d] text-zinc-100 border-[#30363d] font-semibold'
                          : 'bg-[#161b22] text-zinc-400 border-transparent hover:text-zinc-200'
                      }`}
                    >
                      {p.label}
                    </button>
                  ))}
                </div>
              </div>

              <div>
                <span className="text-[10px] uppercase text-zinc-400 font-semibold block mb-1">Category Preset</span>
                <div className="flex flex-wrap gap-1.5">
                  {[
                    { id: 'cve,exposure,vuln', label: 'CVEs & Exposures' },
                    { id: 'misconfig,takeover', label: 'Misconfigurations' },
                    { id: 'cve', label: 'CVEs Only' },
                    { id: 'default-login,exposure', label: 'Exposed Logins' },
                  ].map((p) => (
                    <button
                      key={p.id}
                      type="button"
                      onClick={() => setNucleiTagPreset(p.id)}
                      className={`px-2 py-0.5 rounded text-[11px] border transition ${
                        nucleiTagPreset === p.id
                          ? 'bg-[#21262d] text-zinc-100 border-[#30363d] font-semibold'
                          : 'bg-[#161b22] text-zinc-400 border-transparent hover:text-zinc-200'
                      }`}
                    >
                      {p.label}
                    </button>
                  ))}
                </div>
              </div>
            </div>
          )}

          <div>
            <label className="block text-xs text-zinc-300 mb-1">
              Custom CLI Arguments (Optional Override)
            </label>
            <input
              type="text"
              placeholder="Leave blank to use pre-tuned defaults"
              value={customArgs}
              onChange={(e) => setCustomArgs(e.target.value)}
              className="w-full bg-[#0d1117] border border-[#30363d] rounded px-3 py-1.5 text-xs text-zinc-100 placeholder-zinc-600 focus:outline-none focus:border-zinc-500"
            />
            <p className="text-[11px] text-zinc-500 mt-1">
              Pre-tuned high-speed non-disruptive flags are supplied automatically.
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
