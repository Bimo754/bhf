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
      return <span className="w-2 h-2 rounded-full bg-blue-400 animate-pulse shrink-0" />;
    }
    if (status === 'COMPLETED') {
      return <span className="w-2 h-2 rounded-full bg-emerald-400 shrink-0" />;
    }
    if (status === 'FAILED') {
      return <span className="w-2 h-2 rounded-full bg-red-400 shrink-0" />;
    }
    return <span className="w-2 h-2 rounded-full bg-zinc-500 shrink-0" />;
  };

  return (
    <div className="h-[calc(100vh-7.5rem)] flex flex-col space-y-2">
      {/* Session Tab Strip */}
      <div className="flex items-center space-x-1.5 overflow-x-auto pb-1 scrollbar-thin">
        {scans.length === 0 ? (
          <div className="text-zinc-500 text-xs font-mono py-1 px-2">
            No execution sessions recorded yet.
          </div>
        ) : (
          scans.map((s) => {
            const isSelected = s.id === activeScan?.id;
            const isThisRunning = isRunning && isSelected;

            return (
              <button
                key={s.id}
                onClick={() => onSelectScan(s.id)}
                className={`flex items-center space-x-2 px-3 py-1.5 rounded-md text-xs font-mono border transition-all shrink-0 ${
                  isSelected
                    ? 'bg-[#1c2128] border-zinc-500 text-zinc-100 font-semibold shadow-sm'
                    : 'bg-[#161b22] border-[#30363d] text-zinc-400 hover:text-zinc-200 hover:border-zinc-600'
                }`}
              >
                {getStatusDot(s.status, isThisRunning)}
                <span>{s.tool_name}</span>
                <span className="text-[10px] text-zinc-500">
                  {new Date(s.started_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' })}
                </span>
              </button>
            );
          })
        )}
      </div>

      {/* Active Command Bar */}
      {activeScan && (
        <div className="flex items-center justify-between bg-[#161b22] border border-[#30363d] px-3 py-1.5 rounded-md text-xs font-mono">
          <div className="flex items-center space-x-2 truncate">
            <span className="text-zinc-500 select-none">$</span>
            <span className="text-zinc-200 truncate">{activeScan.command}</span>
          </div>

          <button
            onClick={handleCopyCommand}
            title="Copy command to clipboard"
            className="p-1 rounded text-zinc-400 hover:text-zinc-200 hover:bg-[#21262d] transition ml-2 shrink-0"
          >
            {copiedCmd ? <Check className="w-3.5 h-3.5 text-emerald-400" /> : <Copy className="w-3.5 h-3.5" />}
          </button>
        </div>
      )}

      {/* Terminal Output */}
      <div className="flex-1 overflow-hidden">
        <Terminal
          logs={logs}
          title={activeScan ? `${activeScan.tool_name}` : 'Console'}
          isRunning={isRunning && activeScan?.status === 'RUNNING'}
          onStop={onStop}
          onSkip={onSkipTool}
          onClear={onClear}
        />
      </div>
    </div>
  );
};
