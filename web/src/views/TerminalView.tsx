import React from 'react';
import { Terminal } from '../components/Terminal';
import { LogEntry } from '../types/log';
import { Scan } from '../types/scan';

interface TerminalViewProps {
  scans: Scan[];
  activeScanId: string | null;
  onSelectScan: (scanId: string) => void;
  logs: LogEntry[];
  isRunning: boolean;
  onStop: () => void;
  onClear: () => void;
}

export const TerminalView: React.FC<TerminalViewProps> = ({
  scans,
  activeScanId,
  onSelectScan,
  logs,
  isRunning,
  onStop,
  onClear,
}) => {
  const activeScan = scans.find((s) => s.id === activeScanId);

  return (
    <div className="h-[calc(100vh-7.5rem)] flex flex-col space-y-3">
      {/* Session Header / Selector */}
      <div className="flex items-center justify-between bg-[#161b22] border border-[#30363d] px-4 py-2 rounded-md text-xs">
        <div className="flex items-center space-x-3">
          <span className="text-zinc-400 font-mono">Execution Session:</span>
          <select
            value={activeScanId || ''}
            onChange={(e) => onSelectScan(e.target.value)}
            className="bg-[#0d1117] border border-[#30363d] rounded px-2 py-1 text-zinc-100 font-mono text-xs focus:outline-none"
          >
            {scans.length === 0 ? (
              <option value="">No executions recorded yet</option>
            ) : (
              scans.map((s) => (
                <option key={s.id} value={s.id}>
                  [{s.status}] {s.tool_name} — {new Date(s.started_at).toLocaleTimeString()}
                </option>
              ))
            )}
          </select>
        </div>

        {activeScan && (
          <div className="text-zinc-500 font-mono text-[11px] hidden sm:block truncate max-w-md">
            cmd: <span className="text-zinc-300">{activeScan.command}</span>
          </div>
        )}
      </div>

      {/* Terminal View Container */}
      <div className="flex-1 overflow-hidden">
        <Terminal
          logs={logs}
          title={activeScan ? `${activeScan.tool_name} Console` : 'Execution Output'}
          isRunning={isRunning}
          onStop={onStop}
          onClear={onClear}
        />
      </div>
    </div>
  );
};
