import React, { useState, useEffect, useRef } from 'react';
import { Terminal as TerminalIcon, Copy, Trash2, ArrowDownCircle, Check, Square } from 'lucide-react';
import { LogEntry } from '../types/log';
import { Button } from './Button';

interface TerminalProps {
  logs: LogEntry[];
  title?: string;
  isRunning?: boolean;
  onStop?: () => void;
  onClear?: () => void;
}

export const Terminal: React.FC<TerminalProps> = ({
  logs,
  title = 'Execution Console',
  isRunning = false,
  onStop,
  onClear,
}) => {
  const [autoScroll, setAutoScroll] = useState(true);
  const [copied, setCopied] = useState(false);
  const [filter, setFilter] = useState('');
  const terminalBodyRef = useRef<HTMLDivElement>(null);

  // Directly set container scrollTop to prevent page/window scrolling
  useEffect(() => {
    if (autoScroll && terminalBodyRef.current) {
      terminalBodyRef.current.scrollTop = terminalBodyRef.current.scrollHeight;
    }
  }, [logs, autoScroll]);

  const handleCopy = () => {
    const text = logs.map((l) => `[${l.stream.toUpperCase()}] ${l.line}`).join('\n');
    navigator.clipboard.writeText(text);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const filteredLogs = filter
    ? logs.filter((l) => l.line.toLowerCase().includes(filter.toLowerCase()))
    : logs;

  return (
    <div className="flex flex-col h-full bg-[#010409] rounded-md border border-[#30363d] overflow-hidden">
      {/* Title Bar */}
      <div className="flex items-center justify-between px-3 py-2 bg-[#161b22] border-b border-[#30363d] text-xs">
        <div className="flex items-center space-x-2 text-zinc-300 font-mono">
          <TerminalIcon className="w-3.5 h-3.5 text-zinc-400" />
          <span className="font-semibold text-zinc-200">{title}</span>
          {isRunning && (
            <span className="inline-flex items-center text-blue-400 text-[10px] space-x-1 ml-2 font-mono">
              <span className="w-1.5 h-1.5 rounded-full bg-blue-500 animate-pulse" />
              <span>RUNNING</span>
            </span>
          )}
        </div>

        {/* Controls */}
        <div className="flex items-center space-x-2">
          <input
            type="text"
            placeholder="Filter logs..."
            value={filter}
            onChange={(e) => setFilter(e.target.value)}
            className="bg-[#0d1117] border border-[#30363d] rounded px-2 py-0.5 text-zinc-200 text-xs focus:outline-none focus:border-zinc-500 w-36 font-mono"
          />

          <button
            onClick={() => setAutoScroll(!autoScroll)}
            title={autoScroll ? 'Autoscroll enabled' : 'Autoscroll paused'}
            className={`p-1 rounded hover:bg-[#21262d] transition ${
              autoScroll ? 'text-blue-400' : 'text-zinc-500'
            }`}
          >
            <ArrowDownCircle className="w-3.5 h-3.5" />
          </button>

          <button
            onClick={handleCopy}
            title="Copy logs"
            className="p-1 rounded text-zinc-400 hover:text-zinc-200 hover:bg-[#21262d] transition"
          >
            {copied ? <Check className="w-3.5 h-3.5 text-emerald-400" /> : <Copy className="w-3.5 h-3.5" />}
          </button>

          {onClear && (
            <button
              onClick={onClear}
              title="Clear terminal"
              className="p-1 rounded text-zinc-400 hover:text-zinc-200 hover:bg-[#21262d] transition"
            >
              <Trash2 className="w-3.5 h-3.5" />
            </button>
          )}

          {isRunning && onStop && (
            <Button variant="danger" size="sm" onClick={onStop} className="!py-0.5 !px-2 text-xs">
              <Square className="w-3 h-3 mr-1 fill-current" />
              Stop
            </Button>
          )}
        </div>
      </div>

      {/* Terminal Lines Container (Strictly isolated scrolling) */}
      <div
        ref={terminalBodyRef}
        className="flex-1 p-3 font-mono text-[11px] leading-relaxed overflow-y-auto space-y-0.5 selection:bg-zinc-800"
      >
        {filteredLogs.length === 0 ? (
          <div className="flex items-center justify-center h-full text-zinc-600 font-mono text-xs select-none">
            No execution output to display.
          </div>
        ) : (
          filteredLogs.map((log, index) => (
            <div key={index} className="flex space-x-2.5 hover:bg-white/[0.03] px-1 py-0.5 rounded">
              <span className="text-zinc-600 select-none w-10 text-right shrink-0">{index + 1}</span>
              <span
                className={`w-12 shrink-0 uppercase text-[9px] px-1 py-0.2 rounded text-center select-none font-bold ${
                  log.stream === 'stderr'
                    ? 'text-red-400 bg-red-950/60'
                    : log.stream === 'system'
                    ? 'text-amber-400 bg-amber-950/60'
                    : 'text-zinc-400 bg-zinc-800/60'
                }`}
              >
                {log.stream}
              </span>
              <span
                className={`break-all ${
                  log.stream === 'stderr'
                    ? 'text-red-300'
                    : log.stream === 'system'
                    ? 'text-amber-300'
                    : 'text-zinc-200'
                }`}
              >
                {log.line}
              </span>
            </div>
          ))
        )}
      </div>
    </div>
  );
};
