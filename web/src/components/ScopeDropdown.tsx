import React, { useState, useRef, useEffect } from 'react';
import { ChevronDown, Check, Plus, Trash2, Globe } from 'lucide-react';
import { Target } from '../types/target';

interface ScopeDropdownProps {
  targets: Target[];
  activeTarget: Target | null;
  onSelectTarget: (target: Target) => void;
  onOpenCreateTarget: () => void;
  onOpenDeleteTarget?: () => void;
}

export const ScopeDropdown: React.FC<ScopeDropdownProps> = ({
  targets,
  activeTarget,
  onSelectTarget,
  onOpenCreateTarget,
  onOpenDeleteTarget,
}) => {
  const [isOpen, setIsOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(e.target as Node)) {
        setIsOpen(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  return (
    <div className="relative font-mono" ref={dropdownRef}>
      {/* Trigger Button */}
      <div className="flex items-center bg-white/[0.03] border border-white/[0.08] rounded-lg overflow-hidden hover:border-white/[0.15] transition-all">
        <button
          type="button"
          onClick={() => setIsOpen(!isOpen)}
          className="flex items-center space-x-2 px-3 py-1.5 text-xs text-slate-200 hover:text-white transition-colors"
        >
          <Globe className="w-3.5 h-3.5 text-slate-400" />
          <span className="font-medium max-w-[170px] truncate">
            {activeTarget ? activeTarget.domain : 'Select Scope'}
          </span>
          <ChevronDown className={`w-3.5 h-3.5 text-slate-400 transition-transform duration-150 ${isOpen ? 'rotate-180' : ''}`} />
        </button>

        {activeTarget && onOpenDeleteTarget && (
          <button
            type="button"
            onClick={onOpenDeleteTarget}
            title={`Delete ${activeTarget.domain}`}
            className="p-1.5 text-slate-500 hover:text-rose-400 border-l border-white/[0.08] hover:bg-white/[0.04] transition-colors"
          >
            <Trash2 className="w-3.5 h-3.5" />
          </button>
        )}
      </div>

      {/* Popover Menu */}
      {isOpen && (
        <div className="absolute right-0 mt-1.5 w-64 bg-[#0e1420] border border-white/[0.08] rounded-xl shadow-2xl z-50 py-1 divide-y divide-white/[0.04] animate-in fade-in slide-in-from-top-1 duration-150">
          <div className="px-3 py-1.5 text-[10px] uppercase font-semibold text-slate-500 tracking-wider">
            Target Scopes ({targets.length})
          </div>

          <div className="max-h-60 overflow-y-auto py-1">
            {targets.length === 0 ? (
              <div className="px-3 py-2 text-xs text-slate-500 italic">No targets registered</div>
            ) : (
              targets.map((target) => {
                const isSelected = activeTarget?.id === target.id;
                return (
                  <button
                    key={target.id}
                    onClick={() => {
                      onSelectTarget(target);
                      setIsOpen(false);
                    }}
                    className={`w-full flex items-center justify-between px-3 py-2 text-xs text-left transition-colors ${
                      isSelected
                        ? 'bg-white/[0.06] text-white font-medium'
                        : 'text-slate-300 hover:bg-white/[0.03] hover:text-white'
                    }`}
                  >
                    <span className="truncate">{target.domain}</span>
                    {isSelected && <Check className="w-3.5 h-3.5 text-emerald-400 ml-2 shrink-0" />}
                  </button>
                );
              })
            )}
          </div>

          <div className="p-1.5">
            <button
              onClick={() => {
                setIsOpen(false);
                onOpenCreateTarget();
              }}
              className="w-full flex items-center justify-center space-x-1.5 py-1.5 text-xs text-slate-300 hover:text-white hover:bg-white/[0.04] rounded-lg transition-colors font-medium"
            >
              <Plus className="w-3.5 h-3.5 text-slate-400" />
              <span>Add New Target Scope</span>
            </button>
          </div>
        </div>
      )}
    </div>
  );
};
