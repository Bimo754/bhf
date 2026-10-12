import React from 'react';
import { Shield, Target as TargetIcon, Plus, Layers, Terminal, Database } from 'lucide-react';
import { Target } from '../types/target';
import { Button } from './Button';

interface NavbarProps {
  targets: Target[];
  activeTarget: Target | null;
  activeTab: 'tools' | 'assets' | 'terminal';
  onSelectTarget: (target: Target) => void;
  onOpenCreateTarget: () => void;
  onSelectTab: (tab: 'tools' | 'assets' | 'terminal') => void;
}

export const Navbar: React.FC<NavbarProps> = ({
  targets,
  activeTarget,
  activeTab,
  onSelectTarget,
  onOpenCreateTarget,
  onSelectTab,
}) => {
  return (
    <header className="bg-dark-800/90 border-b border-dark-600/70 backdrop-blur sticky top-0 z-40">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
        {/* Brand */}
        <div className="flex items-center space-x-6">
          <div className="flex items-center space-x-3 cursor-pointer" onClick={() => onSelectTab('tools')}>
            <div className="p-2 rounded-xl bg-cyber-cyan/10 border border-cyber-cyan/30 text-cyber-cyan shadow-sm shadow-cyber-cyan/20">
              <Shield className="w-5 h-5" />
            </div>
            <div>
              <span className="text-lg font-black tracking-wider text-gray-100 font-mono">BHF</span>
              <span className="text-[10px] text-cyber-cyan block font-mono font-medium -mt-1 tracking-widest uppercase">
                Bug Hunter
              </span>
            </div>
          </div>

          {/* Navigation Links */}
          <nav className="hidden md:flex items-center space-x-1 pl-4 border-l border-dark-600/60 text-xs">
            <button
              onClick={() => onSelectTab('tools')}
              className={`flex items-center space-x-2 px-3 py-2 rounded-lg font-medium transition ${
                activeTab === 'tools'
                  ? 'bg-dark-700 text-cyber-cyan'
                  : 'text-gray-400 hover:text-gray-200 hover:bg-dark-700/50'
              }`}
            >
              <Layers className="w-4 h-4" />
              <span>Methodology & Tools</span>
            </button>

            <button
              onClick={() => onSelectTab('assets')}
              className={`flex items-center space-x-2 px-3 py-2 rounded-lg font-medium transition ${
                activeTab === 'assets'
                  ? 'bg-dark-700 text-cyber-cyan'
                  : 'text-gray-400 hover:text-gray-200 hover:bg-dark-700/50'
              }`}
            >
              <Database className="w-4 h-4" />
              <span>Attack Surface Assets</span>
            </button>

            <button
              onClick={() => onSelectTab('terminal')}
              className={`flex items-center space-x-2 px-3 py-2 rounded-lg font-medium transition ${
                activeTab === 'terminal'
                  ? 'bg-dark-700 text-cyber-cyan'
                  : 'text-gray-400 hover:text-gray-200 hover:bg-dark-700/50'
              }`}
            >
              <Terminal className="w-4 h-4" />
              <span>Live Console</span>
            </button>
          </nav>
        </div>

        {/* Target Scope Switcher */}
        <div className="flex items-center space-x-3">
          <div className="flex items-center space-x-2 bg-dark-900 border border-dark-600/80 rounded-lg px-2.5 py-1.5 text-xs">
            <TargetIcon className="w-4 h-4 text-cyber-cyan shrink-0" />
            <select
              value={activeTarget?.id || ''}
              onChange={(e) => {
                const found = targets.find((t) => t.id === e.target.value);
                if (found) onSelectTarget(found);
              }}
              className="bg-transparent text-gray-100 font-mono text-xs focus:outline-none cursor-pointer pr-2"
            >
              {targets.length === 0 ? (
                <option value="">No Targets Defined</option>
              ) : (
                targets.map((t) => (
                  <option key={t.id} value={t.id} className="bg-dark-800 text-gray-100">
                    {t.name} ({t.domain})
                  </option>
                ))
              )}
            </select>
          </div>

          <Button variant="primary" size="sm" onClick={onOpenCreateTarget}>
            <Plus className="w-3.5 h-3.5 mr-1" />
            New Scope
          </Button>
        </div>
      </div>
    </header>
  );
};
