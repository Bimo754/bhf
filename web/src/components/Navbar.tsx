import React from 'react';
import { Target } from '../types/target';
import { ScopeDropdown } from './ScopeDropdown';
import { Globe, Server, ShieldCheck, Terminal, AlertCircle } from 'lucide-react';

export type MainTab = 'web' | 'network' | 'ad' | 'findings' | 'terminal';

interface NavbarProps {
  targets: Target[];
  activeTarget: Target | null;
  activeTab: MainTab;
  webCount?: number;
  networkCount?: number;
  adCount?: number;
  findingsCount?: number;
  isRunning: boolean;
  onSelectTarget: (target: Target) => void;
  onOpenCreateTarget: () => void;
  onOpenDeleteTarget?: () => void;
  onSelectTab: (tab: MainTab) => void;
}

export const Navbar: React.FC<NavbarProps> = ({
  targets,
  activeTarget,
  activeTab,
  webCount = 0,
  networkCount = 0,
  adCount = 0,
  findingsCount = 0,
  isRunning,
  onSelectTarget,
  onOpenCreateTarget,
  onOpenDeleteTarget,
  onSelectTab,
}) => {
  return (
    <header className="bg-[#080c14]/90 backdrop-blur-md border-b border-white/[0.06] sticky top-0 z-40">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 h-14 flex items-center justify-between">
        {/* Brand & Navigation Disciplines */}
        <div className="flex items-center space-x-6">
          <div
            className="flex items-center space-x-2 cursor-pointer select-none"
            onClick={() => onSelectTab('web')}
          >
            <span className="text-sm font-bold tracking-wider text-white font-mono flex items-center gap-1.5">
              <span className="w-2 h-2 rounded-full bg-indigo-500" />
              BHF
            </span>
          </div>

          {/* Segmented Discipline Tabs */}
          <nav className="flex items-center p-0.5 bg-white/[0.02] border border-white/[0.05] rounded-xl text-xs">
            {/* 1. Web Security */}
            <button
              onClick={() => onSelectTab('web')}
              className={`flex items-center space-x-1.5 px-3 py-1.5 rounded-lg transition-all ${
                activeTab === 'web'
                  ? 'bg-white/[0.08] text-white shadow-sm ring-1 ring-white/10 font-medium'
                  : 'text-slate-400 hover:text-slate-200 hover:bg-white/[0.03]'
              }`}
            >
              <Globe className="w-3.5 h-3.5 text-slate-400" />
              <span>Web</span>
              {webCount > 0 && (
                <span className="px-1.5 py-0.2 rounded-full text-[10px] bg-white/[0.06] text-slate-400 font-mono font-normal ml-0.5">
                  {webCount}
                </span>
              )}
            </button>

            {/* 2. Network Security */}
            <button
              onClick={() => onSelectTab('network')}
              className={`flex items-center space-x-1.5 px-3 py-1.5 rounded-lg transition-all ${
                activeTab === 'network'
                  ? 'bg-white/[0.08] text-white shadow-sm ring-1 ring-white/10 font-medium'
                  : 'text-slate-400 hover:text-slate-200 hover:bg-white/[0.03]'
              }`}
            >
              <Server className="w-3.5 h-3.5 text-slate-400" />
              <span>Network</span>
              {networkCount > 0 && (
                <span className="px-1.5 py-0.2 rounded-full text-[10px] bg-white/[0.06] text-slate-400 font-mono font-normal ml-0.5">
                  {networkCount}
                </span>
              )}
            </button>

            {/* 3. Active Directory */}
            <button
              onClick={() => onSelectTab('ad')}
              className={`flex items-center space-x-1.5 px-3 py-1.5 rounded-lg transition-all ${
                activeTab === 'ad'
                  ? 'bg-white/[0.08] text-white shadow-sm ring-1 ring-white/10 font-medium'
                  : 'text-slate-400 hover:text-slate-200 hover:bg-white/[0.03]'
              }`}
            >
              <ShieldCheck className="w-3.5 h-3.5 text-slate-400" />
              <span>Active Directory</span>
              {adCount > 0 && (
                <span className="px-1.5 py-0.2 rounded-full text-[10px] bg-white/[0.06] text-slate-400 font-mono font-normal ml-0.5">
                  {adCount}
                </span>
              )}
            </button>

            {/* 4. Findings */}
            <button
              onClick={() => onSelectTab('findings')}
              className={`flex items-center space-x-1.5 px-3 py-1.5 rounded-lg transition-all ${
                activeTab === 'findings'
                  ? 'bg-white/[0.08] text-white shadow-sm ring-1 ring-white/10 font-medium'
                  : 'text-slate-400 hover:text-slate-200 hover:bg-white/[0.03]'
              }`}
            >
              <AlertCircle className="w-3.5 h-3.5 text-slate-400" />
              <span>Findings</span>
              {findingsCount > 0 && (
                <span className="px-1.5 py-0.2 rounded-full text-[10px] bg-rose-500/10 text-rose-300 font-mono font-medium ml-0.5 border border-rose-500/20">
                  {findingsCount}
                </span>
              )}
            </button>

            {/* 5. Terminal Console */}
            <button
              onClick={() => onSelectTab('terminal')}
              className={`flex items-center space-x-1.5 px-3 py-1.5 rounded-lg transition-all ${
                activeTab === 'terminal'
                  ? 'bg-white/[0.08] text-white shadow-sm ring-1 ring-white/10 font-medium'
                  : 'text-slate-400 hover:text-slate-200 hover:bg-white/[0.03]'
              }`}
            >
              <Terminal className="w-3.5 h-3.5 text-slate-400" />
              <span>Console</span>
              {isRunning && (
                <span className="w-1.5 h-1.5 rounded-full bg-sky-400 animate-pulse ml-1" />
              )}
            </button>
          </nav>
        </div>

        {/* Scope Selector */}
        <div className="flex items-center space-x-3">
          <ScopeDropdown
            targets={targets}
            activeTarget={activeTarget}
            onSelectTarget={onSelectTarget}
            onOpenCreateTarget={onOpenCreateTarget}
            onOpenDeleteTarget={onOpenDeleteTarget}
          />
        </div>
      </div>
    </header>
  );
};
