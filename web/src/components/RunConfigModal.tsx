import React, { useState } from 'react';
import { X, Play, Sliders } from 'lucide-react';
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
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/70 backdrop-blur-sm">
      <div className="bg-dark-800 border border-dark-600 rounded-xl max-w-lg w-full p-6 shadow-2xl relative">
        <button
          onClick={onClose}
          className="absolute top-4 right-4 text-gray-400 hover:text-gray-100"
        >
          <X className="w-5 h-5" />
        </button>

        <div className="flex items-center space-x-3 mb-6">
          <div className="p-2.5 rounded-lg bg-cyber-cyan/10 text-cyber-cyan border border-cyber-cyan/30">
            <Sliders className="w-6 h-6" />
          </div>
          <div>
            <h3 className="text-lg font-bold text-gray-100">Execute {tool.name}</h3>
            <p className="text-xs text-gray-400">Target: <span className="font-mono text-cyber-cyan">{target.domain}</span></p>
          </div>
        </div>

        <div className="space-y-4">
          <div className="p-3 bg-dark-900 border border-dark-700 rounded-lg text-xs">
            <span className="text-gray-400 block mb-1 font-semibold uppercase tracking-wider">Description</span>
            <p className="text-gray-200">{tool.description}</p>
          </div>

          <div>
            <label className="block text-xs font-medium text-gray-300 mb-1">
              Custom CLI Flags & Arguments (Optional Override)
            </label>
            <input
              type="text"
              placeholder="Leave blank for optimized default arguments"
              value={customArgs}
              onChange={(e) => setCustomArgs(e.target.value)}
              className="w-full bg-dark-900 border border-dark-600 rounded-lg px-3 py-2 text-sm text-gray-100 placeholder-gray-500 focus:outline-none focus:border-cyber-cyan font-mono"
            />
            <p className="text-[11px] text-gray-500 mt-1">
              If left blank, BHF uses tuned flags for high-speed, non-intrusive enumeration.
            </p>
          </div>

          <div className="flex justify-end space-x-3 pt-3">
            <Button variant="ghost" onClick={onClose}>
              Cancel
            </Button>
            <Button variant="primary" onClick={handleRun}>
              <Play className="w-4 h-4 mr-1.5" />
              Launch Tool
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
};
