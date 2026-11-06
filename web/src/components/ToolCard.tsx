import React from 'react';
import { Play, Sliders, Square, Terminal } from 'lucide-react';
import { ToolInfo } from '../types/tool';
import { Scan } from '../types/scan';
import { Button } from './Button';

interface ToolCardProps {
  tool: ToolInfo;
  latestScan?: Scan;
  isRunning: boolean;
  stepNumber?: number;
  onRunClick: (toolName: string) => void;
  onConfigClick: (tool: ToolInfo) => void;
  onViewLogs: (scanId: string) => void;
  onStop: (scanId: string) => void;
}

export const ToolCard: React.FC<ToolCardProps> = ({
  tool,
  latestScan,
  isRunning,
  stepNumber,
  onRunClick,
  onConfigClick,
  onViewLogs,
  onStop,
}) => {
  const getStatusIndicator = () => {
    if (isRunning) {
      return (
        <span className="flex items-center space-x-1.5 text-xs text-sky-400 font-mono">
          <span className="w-2 h-2 rounded-full bg-sky-400 animate-pulse" />
          <span>running</span>
        </span>
      );
    }
    if (!latestScan) {
      return (
        <span className="flex items-center space-x-1.5 text-xs text-slate-500 font-mono">
          <span className="w-1.5 h-1.5 rounded-full bg-slate-600" />
          <span>ready</span>
        </span>
      );
    }
    if (latestScan.status === 'COMPLETED') {
      return (
        <span className="flex items-center space-x-1.5 text-xs text-emerald-400 font-mono">
          <span className="w-1.5 h-1.5 rounded-full bg-emerald-400" />
          <span>completed</span>
        </span>
      );
    }
    if (latestScan.status === 'FAILED') {
      return (
        <span className="flex items-center space-x-1.5 text-xs text-rose-400 font-mono">
          <span className="w-1.5 h-1.5 rounded-full bg-rose-400" />
          <span>failed</span>
        </span>
      );
    }
    return (
      <span className="flex items-center space-x-1.5 text-xs text-slate-500 font-mono">
        <span className="w-1.5 h-1.5 rounded-full bg-slate-600" />
        <span>{latestScan.status.toLowerCase()}</span>
      </span>
    );
  };

  return (
    <div className="group bg-[#0e1420]/80 hover:bg-[#141b2b]/90 border border-white/[0.05] hover:border-white/[0.12] rounded-xl p-3 flex items-center justify-between transition-all duration-150">
      <div className="flex items-center space-x-3 min-w-0">
        {stepNumber !== undefined && (
          <span className="text-[11px] font-mono text-slate-500 w-5 text-right shrink-0 select-none">
            {stepNumber < 10 ? `0${stepNumber}` : stepNumber}
          </span>
        )}

        <div className="min-w-0">
          <div className="flex items-center space-x-2.5">
            <h4 className="text-xs font-semibold text-slate-100 font-mono tracking-wide truncate">
              {tool.name}
            </h4>
            {getStatusIndicator()}
          </div>
          <p className="text-[11px] text-slate-400 truncate max-w-sm mt-0.5">
            {tool.description}
          </p>
        </div>
      </div>

      <div className="flex items-center space-x-1.5 shrink-0 ml-3">
        {isRunning && latestScan ? (
          <>
            <Button
              variant="danger"
              size="xs"
              onClick={() => onStop(latestScan.id)}
              className="flex items-center space-x-1 font-mono"
            >
              <Square className="w-2.5 h-2.5 fill-current" />
              <span>Stop</span>
            </Button>
            <button
              onClick={() => onViewLogs(latestScan.id)}
              title="Inspect logs"
              className="p-1.5 rounded-lg text-slate-400 hover:text-white hover:bg-white/[0.06] transition-colors"
            >
              <Terminal className="w-3.5 h-3.5" />
            </button>
          </>
        ) : (
          <>
            <Button
              variant="secondary"
              size="xs"
              onClick={() => onRunClick(tool.name)}
              className="font-mono text-slate-300 hover:text-white"
              title="Run tool with defaults"
            >
              <Play className="w-2.5 h-2.5 mr-1 fill-current text-slate-400 group-hover:text-indigo-400 transition-colors" />
              <span>Run</span>
            </Button>

            <button
              onClick={() => onConfigClick(tool)}
              title="Configure arguments"
              className="p-1.5 rounded-lg text-slate-500 hover:text-slate-200 hover:bg-white/[0.06] transition-colors"
            >
              <Sliders className="w-3.5 h-3.5" />
            </button>

            {latestScan && (
              <button
                onClick={() => onViewLogs(latestScan.id)}
                title="View output"
                className="p-1.5 rounded-lg text-slate-500 hover:text-slate-200 hover:bg-white/[0.06] transition-colors"
              >
                <Terminal className="w-3.5 h-3.5" />
              </button>
            )}
          </>
        )}
      </div>
    </div>
  );
};
