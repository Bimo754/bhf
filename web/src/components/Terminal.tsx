import React, { useState, useEffect, useRef } from 'react';
import { Terminal as TerminalIcon, Copy, Trash2, ArrowDownCircle, Check, Square, SkipForward } from 'lucide-react';
import { LogEntry } from '../types/log';
import { Button } from './Button';

interface TerminalProps {
  logs: LogEntry[];
  title?: string;
  isRunning?: boolean;
  onStop?: () => void;
  onSkip?: () => void;
  onClear?: () => void;
}

export const Terminal: React.FC<TerminalProps> = ({
  logs,
  title = 'Execution Console',
  isRunning = false,
  onStop,
  onSkip,
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

  const renderHighlightedLine = (text: string, query: string) => {
    if (!query) return text;
    const lowerText = text.toLowerCase();
    const lowerQuery = query.toLowerCase();
    const parts = [];
    let start = 0;
    let matchIdx = lowerText.indexOf(lowerQuery, start);

    if (matchIdx === -1) return text;

    while (matchIdx !== -1) {
      if (matchIdx > start) {
        parts.push(text.substring(start, matchIdx));
      }
      const matched = text.substring(matchIdx, matchIdx + query.length);
      parts.push(
        <mark
          key={start}
          className="bg-amber-300 text-black px-0.5 rounded-sm font-semibold select-text"
        >
          {matched}
        </mark>
      );
      start = matchIdx + query.length;
      matchIdx = lowerText.indexOf(lowerQuery, start);
    }

    if (start < text.length) {
      parts.push(text.substring(start));
    }

    return parts;
  };

  return (
    <div className="flex flex-col h-full bg-[#05080e] rounded-xl border border-white/[0.06] overflow-hidden">
      {/* Title Bar */}
      <div className="flex items-center justify-between px-3.5 py-2.5 bg-[#0e1420] border-b border-white/[0.05] text-xs">
        <div className="flex items-center space-x-2 text-slate-300 font-mono">
          <TerminalIcon className="w-3.5 h-3.5 text-slate-400" />
          <span className="font-semibold text-slate-200">{title}</span>
          {isRunning && (
            <span className="inline-flex items-center text-sky-400 text-[10px] space-x-1 ml-2 font-mono">
              <span className="w-1.5 h-1.5 rounded-full bg-sky-400 animate-pulse" />
              <span>RUNNING</span>
            </span>
          )}
        </div>

        {/* Controls */}
        <div className="flex items-center space-x-2">
          {filter && (
            <span className="text-[10px] text-slate-500 font-mono hidden sm:inline">
              {filteredLogs.length} matches
            </span>
          )}

          <input
            type="text"
            placeholder="Search console..."
            value={filter}
            onChange={(e) => setFilter(e.target.value)}
            className="bg-white/[0.03] border border-white/[0.08] rounded-lg px-2.5 py-1 text-slate-200 text-xs focus:outline-none focus:border-indigo-500 w-40 font-mono placeholder-slate-500"
          />

          <button
            onClick={() => setAutoScroll(!autoScroll)}
            title={autoScroll ? 'Autoscroll active' : 'Autoscroll paused'}
            className={`p-1.5 rounded-lg hover:bg-white/[0.04] transition ${
              autoScroll ? 'text-indigo-400' : 'text-slate-500'
            }`}
          >
            <ArrowDownCircle className="w-3.5 h-3.5" />
          </button>

          <button
            onClick={handleCopy}
            title="Copy logs"
            className="p-1.5 rounded-lg text-slate-400 hover:text-white hover:bg-white/[0.04] transition"
          >
            {copied ? <Check className="w-3.5 h-3.5 text-emerald-400" /> : <Copy className="w-3.5 h-3.5" />}
          </button>

          {onClear && (
            <button
              onClick={onClear}
              title="Clear terminal"
              className="p-1.5 rounded-lg text-slate-400 hover:text-white hover:bg-white/[0.04] transition"
            >
              <Trash2 className="w-3.5 h-3.5" />
            </button>
          )}

          {isRunning && onSkip && (
            <Button
              variant="secondary"
              size="xs"
              onClick={onSkip}
              className="font-mono text-slate-300 hover:text-white"
              title="Skip current tool and advance pipeline"
            >
              <SkipForward className="w-3 h-3 mr-1 text-amber-400" />
              <span>Skip</span>
            </Button>
          )}

          {isRunning && onStop && (
            <Button variant="danger" size="xs" onClick={onStop}>
              <Square className="w-3 h-3 mr-1 fill-current" />
              <span>Stop</span>
            </Button>
          )}
        </div>
      </div>

      {/* Terminal Body */}
      <div
        ref={terminalBodyRef}
        className="flex-1 p-3 overflow-y-auto font-mono text-xs leading-relaxed select-text space-y-0.5 text-slate-300"
      >
        {filteredLogs.length === 0 ? (
          <div className="text-slate-600 italic select-none py-4 text-center">
            {filter ? 'No log lines matching query.' : 'Awaiting command execution output...'}
          </div>
        ) : (
          filteredLogs.map((log, idx) => {
            const isStderr = log.stream === 'stderr';
            const isSystem = log.stream === 'system';

            let lineClass = 'text-slate-300';
            if (isStderr) lineClass = 'text-rose-400';
            if (isSystem) lineClass = 'text-sky-400 font-semibold';

            return (
              <div key={idx} className="flex items-start space-x-2 hover:bg-white/[0.02] px-1 rounded">
                <span className="text-[10px] text-slate-600 select-none shrink-0 w-16 pt-0.5">
                  {new Date(log.timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' })}
                </span>
                <span className={`break-all whitespace-pre-wrap flex-1 ${lineClass}`}>
                  {renderHighlightedLine(log.line, filter)}
                </span>
              </div>
            );
          })
        )}
      </div>
    </div>
  );
};
