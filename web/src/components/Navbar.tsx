import React from 'react';
import { Target } from '../types/target';
import { Button } from './Button';

interface NavbarProps {
  targets: Target[];
  activeTarget: Target | null;
  activeTab: 'tools' | 'assets' | 'terminal';
  isRunning: boolean;
  onSelectTarget: (target: Target) => void;
  onOpenCreateTarget: () => void;
  onSelectTab: (tab: 'tools' | 'assets' | 'terminal') => void;
}

export const Navbar: React.FC<NavbarProps> = ({
  targets,
  activeTarget,
  activeTab,
  isRunning,
  onSelectTarget,
  onOpenCreateTarget,
  onSelectTab,
}) => {
  return (
    <header className="bg-[#161b22] border-b border-[#30363d] sticky top-0 z-40">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 h-14 flex items-center justify-between">
        {/* Brand & Tabs */}
        <div className="flex items-center space-x-6">
          <div
            className="flex items-center space-x-2.5 cursor-pointer select-none"
            onClick={() => onSelectTab('tools')}
          >
            <span className="text-base font-black tracking-wider text-zinc-100 font-mono">BHF</span>
            <span className="text-xs text-zinc-500 font-mono hidden sm:inline">| Bug Hunting Framework</span>
          </div>

          {/* Navigation Links */}
          <nav className="flex items-center space-x-1 text-xs font-mono">
            <button
              onClick={() => onSelectTab('tools')}
              className={`px-3 py-1.5 rounded transition-colors ${
                activeTab === 'tools'
                  ? 'bg-[#21262d] text-zinc-100 font-medium'
                  : 'text-zinc-400 hover:text-zinc-200 hover:bg-[#21262d]/50'
              }`}
            >
              Methodology
            </button>

            <button
              onClick={() => onSelectTab('assets')}
              className={`px-3 py-1.5 rounded transition-colors ${
                activeTab === 'assets'
                  ? 'bg-[#21262d] text-zinc-100 font-medium'
                  : 'text-zinc-400 hover:text-zinc-200 hover:bg-[#21262d]/50'
              }`}
            >
              Discovered Assets
            </button>

            <button
              onClick={() => onSelectTab('terminal')}
              className={`flex items-center space-x-1.5 px-3 py-1.5 rounded transition-colors ${
                activeTab === 'terminal'
                  ? 'bg-[#21262d] text-zinc-100 font-medium'
                  : 'text-zinc-400 hover:text-zinc-200 hover:bg-[#21262d]/50'
              }`}
            >
              <span>Console</span>
              {isRunning && (
                <span className="w-2 h-2 rounded-full bg-blue-500 animate-pulse ml-1" />
              )}
            </button>
          </nav>
        </div>

        {/* Target Scope Switcher */}
        <div className="flex items-center space-x-2">
          <div className="flex items-center bg-[#0d1117] border border-[#30363d] rounded px-2.5 py-1 text-xs">
            <span className="text-zinc-500 font-mono mr-2 text-[11px]">Scope:</span>
            <select
              value={activeTarget?.id || ''}
              onChange={(e) => {
                const found = targets.find((t) => t.id === e.target.value);
                if (found) onSelectTarget(found);
              }}
              className="bg-transparent text-zinc-200 font-mono text-xs focus:outline-none cursor-pointer pr-1"
            >
              {targets.length === 0 ? (
                <option value="">No Targets Defined</option>
              ) : (
                targets.map((t) => (
                  <option key={t.id} value={t.id} className="bg-[#161b22] text-zinc-200">
                    {t.domain}
                  </option>
                ))
              )}
            </select>
          </div>

          <Button variant="secondary" size="sm" onClick={onOpenCreateTarget}>
            + Add Target
          </Button>
        </div>
      </div>
    </header>
  );
};
