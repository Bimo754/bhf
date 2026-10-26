import React from 'react';
import { Target } from '../types/target';
import { Button } from './Button';
import { ScopeDropdown } from './ScopeDropdown';

interface NavbarProps {
  targets: Target[];
  activeTarget: Target | null;
  activeTab: 'tools' | 'assets' | 'findings' | 'terminal';
  findingsCount?: number;
  isRunning: boolean;
  onSelectTarget: (target: Target) => void;
  onOpenCreateTarget: () => void;
  onOpenDeleteTarget?: () => void;
  onSelectTab: (tab: 'tools' | 'assets' | 'findings' | 'terminal') => void;
}

export const Navbar: React.FC<NavbarProps> = ({
  targets,
  activeTarget,
  activeTab,
  findingsCount = 0,
  isRunning,
  onSelectTarget,
  onOpenCreateTarget,
  onOpenDeleteTarget,
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
              onClick={() => onSelectTab('findings')}
              className={`flex items-center space-x-1.5 px-3 py-1.5 rounded transition-colors ${
                activeTab === 'findings'
                  ? 'bg-[#21262d] text-zinc-100 font-medium'
                  : 'text-zinc-400 hover:text-zinc-200 hover:bg-[#21262d]/50'
              }`}
            >
              <span>Findings</span>
              {findingsCount > 0 && (
                <span className="px-1.5 py-0.2 rounded-full text-[10px] bg-red-950/80 text-red-400 border border-red-800 font-bold ml-1">
                  {findingsCount}
                </span>
              )}
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
          <ScopeDropdown
            targets={targets}
            activeTarget={activeTarget}
            onSelectTarget={onSelectTarget}
            onOpenCreateTarget={onOpenCreateTarget}
            onOpenDeleteTarget={onOpenDeleteTarget}
          />

          <Button variant="secondary" size="sm" onClick={onOpenCreateTarget} className="text-xs font-mono">
            + Scope
          </Button>
        </div>
      </div>
    </header>
  );
};
