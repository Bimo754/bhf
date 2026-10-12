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
      <div className="flex items-center bg-[#0d1117] border border-[#30363d] rounded-md overflow-hidden hover:border-zinc-500 transition-colors">
        <button
          type="button"
          onClick={() => setIsOpen(!isOpen)}
          className="flex items-center space-x-2 px-3 py-1.5 text-xs text-zinc-200 hover:text-white transition-colors"
        >
          <Globe className="w-3.5 h-3.5 text-zinc-400" />
          <span className="font-semibold max-w-[180px] truncate">
            {activeTarget ? activeTarget.domain : 'Select Target'}
          </span>
          <ChevronDown className={`w-3.5 h-3.5 text-zinc-400 transition-transform ${isOpen ? 'rotate-180' : ''}`} />
        </button>

        {activeTarget && onOpenDeleteTarget && (
          <button
            type="button"
            onClick={onOpenDeleteTarget}
            title={`Delete ${activeTarget.domain}`}
            className="p-1.5 text-zinc-500 hover:text-red-400 border-l border-[#30363d] hover:bg-[#21262d] transition-colors"
          >
            <Trash2 className="w-3.5 h-3.5" />
          </button>
        )}
      </div>

      {/* Popover Menu */}
      {isOpen && (
        <div className="absolute right-0 mt-1.5 w-64 bg-[#161b22] border border-[#30363d] rounded-md shadow-2xl z-50 py-1 divide-y divide-[#21262d] animate-in fade-in slide-in-from-top-1 duration-150">
          <div className="px-3 py-1.5 text-[10px] uppercase font-bold text-zinc-500 tracking-wider">
            Target Scopes ({targets.length})
          </div>

          <div className="max-h-60 overflow-y-auto py-1">
            {targets.length === 0 ? (
              <div className="px-3 py-2 text-xs text-zinc-500 italic">No targets defined yet</div>
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
                    className={`w-full flex items-center justify-between px-3 py-1.5 text-xs text-left transition-colors ${
                      isSelected
                        ? 'bg-[#21262d] text-zinc-100 font-semibold'
                        : 'text-zinc-300 hover:bg-[#1c2128] hover:text-white'
                    }`}
                  >
                    <span className="truncate">{target.domain}</span>
                    {isSelected && <Check className="w-3.5 h-3.5 text-emerald-400 ml-2 shrink-0" />}
                  </button>
                );
              })
            )}
          </div>

          <div className="p-1">
            <button
              onClick={() => {
                setIsOpen(false);
                onOpenCreateTarget();
              }}
              className="w-full flex items-center space-x-2 px-2.5 py-1.5 text-xs text-zinc-300 hover:text-white hover:bg-[#21262d] rounded transition-colors"
            >
              <Plus className="w-3.5 h-3.5 text-blue-400" />
              <span>Add New Target Scope</span>
            </button>
          </div>
        </div>
      )}
    </div>
  );
};
