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
  const getStatusBadge = () => {
    if (isRunning) {
      return (
        <span className="inline-flex items-center space-x-1.5 px-2 py-0.5 rounded text-[10px] font-mono font-semibold bg-blue-500/10 text-blue-400 border border-blue-500/20">
          <span className="w-1.5 h-1.5 rounded-full bg-blue-400 animate-pulse" />
          <span>RUNNING</span>
        </span>
      );
    }
    if (!latestScan) return null;
    if (latestScan.status === 'COMPLETED') {
      return (
        <span className="inline-flex items-center space-x-1.5 px-2 py-0.5 rounded text-[10px] font-mono font-semibold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
          <span className="w-1.5 h-1.5 rounded-full bg-emerald-400" />
          <span>COMPLETED</span>
        </span>
      );
    }
    if (latestScan.status === 'FAILED') {
      return (
        <span className="inline-flex items-center space-x-1.5 px-2 py-0.5 rounded text-[10px] font-mono font-semibold bg-red-500/10 text-red-400 border border-red-500/20">
          <span className="w-1.5 h-1.5 rounded-full bg-red-400" />
          <span>FAILED</span>
        </span>
      );
    }
    if (latestScan.status === 'STOPPED') {
      return (
        <span className="inline-flex items-center space-x-1.5 px-2 py-0.5 rounded text-[10px] font-mono font-semibold bg-zinc-800 text-zinc-400 border border-zinc-700">
          <span className="w-1.5 h-1.5 rounded-full bg-zinc-400" />
          <span>STOPPED</span>
        </span>
      );
    }
    return null;
  };

  return (
    <div className="bg-[#161b22] border border-[#30363d] rounded-md p-3.5 flex items-center justify-between hover:border-zinc-500 transition-colors">
      <div className="flex items-center space-x-2.5">
        {stepNumber !== undefined && (
          <span
            className="px-1.5 py-0.5 rounded text-[10px] font-mono font-bold text-zinc-400 bg-[#21262d] border border-[#30363d]"
            title={`Step ${stepNumber} in this methodology stage`}
          >
            #{stepNumber < 10 ? `0${stepNumber}` : stepNumber}
          </span>
        )}
        <h4 className="text-sm font-semibold text-zinc-100 font-mono tracking-wide">
          {tool.name}
        </h4>
        {getStatusBadge()}
      </div>

      <div className="flex items-center space-x-1.5">
        {isRunning && latestScan ? (
          <>
            <Button
              variant="danger"
              size="sm"
              onClick={() => onStop(latestScan.id)}
              className="!py-1 !px-2.5 text-xs font-mono"
            >
              <Square className="w-3 h-3 mr-1 fill-current" />
              Stop
            </Button>
            <Button
              variant="secondary"
              size="sm"
              onClick={() => onViewLogs(latestScan.id)}
              className="!p-1.5"
              title="Inspect live console"
            >
              <Terminal className="w-3.5 h-3.5 text-zinc-300" />
            </Button>
          </>
        ) : (
          <>
            <Button
              variant="primary"
              size="sm"
              onClick={() => onRunClick(tool.name)}
              className="!py-1 !px-3 text-xs font-mono"
              title="Run tool with standard flags"
            >
              <Play className="w-3 h-3 mr-1 fill-current" />
              Run
            </Button>

            <button
              onClick={() => onConfigClick(tool)}
              title="Configure custom CLI flags"
              className="p-1.5 rounded text-zinc-400 hover:text-zinc-100 hover:bg-[#21262d] border border-[#30363d] transition-colors"
            >
              <Sliders className="w-3.5 h-3.5" />
            </button>

            {latestScan && (
              <button
                onClick={() => onViewLogs(latestScan.id)}
                title="View previous output"
                className="p-1.5 rounded text-zinc-400 hover:text-zinc-100 hover:bg-[#21262d] border border-[#30363d] transition-colors"
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
