import React, { useState, useEffect, useRef } from 'react';
import { Terminal as TerminalIcon, Copy, Trash2, ArrowDownCircle, Check, Octagon } from 'lucide-react';
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
  title = 'Live Execution Console',
  isRunning = false,
  onStop,
  onClear,
}) => {
  const [autoScroll, setAutoScroll] = useState(true);
  const [copied, setCopied] = useState(false);
  const [filter, setFilter] = useState('');
  const bottomRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (autoScroll && bottomRef.current) {
      bottomRef.current.scrollIntoView({ behavior: 'smooth' });
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
    <div className="flex flex-col h-full bg-[#0d1117] rounded-xl border border-dark-600/60 overflow-hidden shadow-2xl">
      {/* Terminal Titlebar */}
      <div className="flex items-center justify-between px-4 py-2.5 bg-dark-800/90 border-b border-dark-600/40 text-xs">
        <div className="flex items-center space-x-3">
          <div className="flex space-x-1.5">
            <span className="w-3 h-3 rounded-full bg-red-500/80 inline-block" />
            <span className="w-3 h-3 rounded-full bg-yellow-500/80 inline-block" />
            <span className="w-3 h-3 rounded-full bg-green-500/80 inline-block" />
          </div>
          <div className="flex items-center space-x-2 text-gray-300 font-mono font-medium">
            <TerminalIcon className="w-3.5 h-3.5 text-cyber-cyan" />
            <span>{title}</span>
            {isRunning && (
              <span className="flex items-center text-cyber-cyan text-[10px] space-x-1">
                <span className="w-1.5 h-1.5 rounded-full bg-cyber-cyan animate-ping" />
                <span>ACTIVE</span>
              </span>
            )}
          </div>
        </div>

        {/* Action Controls */}
        <div className="flex items-center space-x-2">
          <input
            type="text"
            placeholder="Filter output..."
            value={filter}
            onChange={(e) => setFilter(e.target.value)}
            className="bg-dark-900 border border-dark-600 rounded px-2 py-1 text-gray-200 text-xs focus:outline-none focus:border-cyber-cyan w-32 md:w-44 font-mono"
          />

          <button
            onClick={() => setAutoScroll(!autoScroll)}
            title="Toggle autoscroll"
            className={`p-1.5 rounded hover:bg-dark-700 transition ${
              autoScroll ? 'text-cyber-cyan' : 'text-gray-500'
            }`}
          >
            <ArrowDownCircle className="w-3.5 h-3.5" />
          </button>

          <button
            onClick={handleCopy}
            title="Copy logs"
            className="p-1.5 rounded text-gray-400 hover:text-gray-200 hover:bg-dark-700 transition"
          >
            {copied ? <Check className="w-3.5 h-3.5 text-cyber-green" /> : <Copy className="w-3.5 h-3.5" />}
          </button>

          {onClear && (
            <button
              onClick={onClear}
              title="Clear terminal"
              className="p-1.5 rounded text-gray-400 hover:text-gray-200 hover:bg-dark-700 transition"
            >
              <Trash2 className="w-3.5 h-3.5" />
            </button>
          )}

          {isRunning && onStop && (
            <Button variant="danger" size="sm" onClick={onStop} className="!py-1 !px-2 text-xs">
              <Octagon className="w-3 h-3 mr-1" />
              Stop
            </Button>
          )}
        </div>
      </div>

      {/* Terminal Output Body */}
      <div className="flex-1 p-4 font-mono text-xs overflow-y-auto space-y-1 selection:bg-cyan-900">
        {filteredLogs.length === 0 ? (
          <div className="flex items-center justify-center h-full text-gray-600 italic select-none">
            Waiting for tool execution to stream output...
          </div>
        ) : (
          filteredLogs.map((log, index) => (
            <div key={index} className="flex space-x-3 leading-relaxed hover:bg-white/[0.02] px-1 rounded">
              <span className="text-gray-600 select-none w-8 text-right shrink-0">{index + 1}</span>
              <span
                className={`w-12 shrink-0 uppercase text-[10px] px-1 py-0.5 rounded text-center select-none ${
                  log.stream === 'stderr'
                    ? 'text-red-400 bg-red-950/40'
                    : log.stream === 'system'
                    ? 'text-yellow-400 bg-yellow-950/40'
                    : 'text-gray-400 bg-gray-800/40'
                }`}
              >
                {log.stream}
              </span>
              <span
                className={`break-all ${
                  log.stream === 'stderr'
                    ? 'text-red-300'
                    : log.stream === 'system'
                    ? 'text-yellow-300'
                    : 'text-gray-200'
                }`}
              >
                {log.line}
              </span>
            </div>
          ))
        )}
        <div ref={bottomRef} />
      </div>
    </div>
  );
};
