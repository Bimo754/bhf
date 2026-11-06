import React, { useState } from 'react';
import { Terminal } from '../components/Terminal';
import { LogEntry } from '../types/log';
import { Scan } from '../types/scan';
import { Copy, Check } from 'lucide-react';

interface TerminalViewProps {
  scans: Scan[];
  activeScanId: string | null;
  onSelectScan: (scanId: string) => void;
  logs: LogEntry[];
  isRunning: boolean;
  onStop: () => void;
  onSkipTool?: () => void;
  onClear: () => void;
}

export const TerminalView: React.FC<TerminalViewProps> = ({
  scans,
  activeScanId,
  onSelectScan,
  logs,
  isRunning,
  onStop,
  onSkipTool,
  onClear,
}) => {
  const [copiedCmd, setCopiedCmd] = useState(false);
  const activeScan = scans.find((s) => s.id === activeScanId) || (scans.length > 0 ? scans[0] : null);

  const handleCopyCommand = () => {
    if (!activeScan?.command) return;
    navigator.clipboard.writeText(activeScan.command);
    setCopiedCmd(true);
    setTimeout(() => setCopiedCmd(false), 2000);
  };

  const getStatusDot = (status: string, isCurrentRunning: boolean) => {
    if (isCurrentRunning) {
      return <span className="w-1.5 h-1.5 rounded-full bg-sky-400 animate-pulse shrink-0" />;
    }
    if (status === 'COMPLETED') {
      return <span className="w-1.5 h-1.5 rounded-full bg-emerald-400 shrink-0" />;
    }
    if (status === 'FAILED') {
      return <span className="w-1.5 h-1.5 rounded-full bg-rose-400 shrink-0" />;
    }
    return <span className="w-1.5 h-1.5 rounded-full bg-slate-600 shrink-0" />;
  };

  return (
    <div className="h-[calc(100vh-7.5rem)] flex flex-col space-y-2.5">
      {/* Session Tab Strip */}
      <div className="flex items-center space-x-1.5 overflow-x-auto pb-1">
        {scans.length === 0 ? (
          <div className="text-slate-500 text-xs font-mono py-1 px-2">
            No active execution sessions recorded yet.
          </div>
        ) : (
          scans.map((s) => {
            const isSelected = s.id === activeScan?.id;
            const isThisRunning = isRunning && isSelected;

            return (
              <button
                key={s.id}
                onClick={() => onSelectScan(s.id)}
                className={`flex items-center space-x-2 px-3 py-1.5 rounded-lg text-xs font-mono transition-all shrink-0 ${
                  isSelected
                    ? 'bg-white/[0.08] text-white shadow-sm ring-1 ring-white/10 font-medium'
                    : 'bg-[#0e1420] text-slate-400 hover:text-slate-200 hover:bg-white/[0.04] border border-white/[0.05]'
                }`}
              >
                {getStatusDot(s.status, isThisRunning)}
                <span>{s.tool_name}</span>
                <span className="text-[10px] text-slate-500 font-normal">
                  {new Date(s.started_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                </span>
              </button>
            );
          })
        )}
      </div>

      {/* Active Command Bar */}
      {activeScan && (
        <div className="flex items-center justify-between bg-[#0e1420] border border-white/[0.06] px-3.5 py-2 rounded-xl text-xs font-mono">
          <div className="flex items-center space-x-2 truncate">
            <span className="text-slate-600 select-none">$</span>
            <span className="text-slate-200 truncate">{activeScan.command}</span>
          </div>

          <button
            onClick={handleCopyCommand}
            title="Copy command to clipboard"
            className="p-1 rounded text-slate-400 hover:text-white hover:bg-white/[0.04] transition ml-2 shrink-0"
          >
            {copiedCmd ? <Check className="w-3.5 h-3.5 text-emerald-400" /> : <Copy className="w-3.5 h-3.5" />}
          </button>
        </div>
      )}

      {/* Terminal Container */}
      <div className="flex-1 min-h-0">
        <Terminal
          logs={logs}
          title={activeScan ? `${activeScan.tool_name} Console` : 'Session Output'}
          isRunning={isRunning}
          onStop={onStop}
          onSkip={onSkipTool}
          onClear={onClear}
        />
      </div>
    </div>
  );
};
