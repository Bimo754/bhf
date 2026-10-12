import React from 'react';
import { Play, Terminal, Square, Check, AlertTriangle } from 'lucide-react';
import { ToolInfo } from '../types/tool';
import { Scan } from '../types/scan';
import { Badge } from './Badge';
import { Button } from './Button';

interface ToolCardProps {
  tool: ToolInfo;
  latestScan?: Scan;
  isRunning: boolean;
  onRunClick: (tool: ToolInfo) => void;
  onViewLogs: (scanId: string) => void;
  onStop: (scanId: string) => void;
}

export const ToolCard: React.FC<ToolCardProps> = ({
  tool,
  latestScan,
  isRunning,
  onRunClick,
  onViewLogs,
  onStop,
}) => {
  return (
    <div className="bg-[#161b22] border border-[#30363d] rounded-md p-4 flex flex-col justify-between hover:border-zinc-500 transition-colors">
      <div>
        {/* Header */}
        <div className="flex items-start justify-between mb-2">
          <div>
            <h4 className="text-sm font-semibold text-zinc-100 font-mono flex items-center space-x-2">
              <span>{tool.name}</span>
            </h4>
            <span className="text-[11px] text-zinc-500 font-mono">bin: {tool.binary_name}</span>
          </div>

          <div className="flex items-center space-x-1.5 text-xs font-mono">
            {tool.is_installed ? (
              <span className="inline-flex items-center text-emerald-400 text-[11px] space-x-1" title="Binary available in PATH">
                <Check className="w-3.5 h-3.5" />
                <span>Installed</span>
              </span>
            ) : (
              <span className="inline-flex items-center text-amber-400 text-[11px] space-x-1" title="Binary not found in PATH">
                <AlertTriangle className="w-3.5 h-3.5" />
                <span>Missing</span>
              </span>
            )}
          </div>
        </div>

        {/* Description */}
        <p className="text-xs text-zinc-400 leading-relaxed mb-3">
          {tool.description}
        </p>

        {/* Previous Run Status */}
        {latestScan && (
          <div className="flex items-center justify-between text-xs py-1.5 px-2 rounded bg-[#0d1117] border border-[#21262d] mb-3">
            <span className="text-zinc-500 font-mono text-[11px]">Last run:</span>
            <Badge status={latestScan.status} />
          </div>
        )}
      </div>

      {/* Action Footer */}
      <div className="flex items-center space-x-2 pt-2 border-t border-[#21262d]">
        {isRunning && latestScan ? (
          <>
            <Button
              variant="danger"
              size="sm"
              onClick={() => onStop(latestScan.id)}
              className="flex-1"
            >
              <Square className="w-3 h-3 mr-1.5 fill-current" />
              Stop Execution
            </Button>
            <Button
              variant="secondary"
              size="sm"
              onClick={() => onViewLogs(latestScan.id)}
              className="px-2.5"
              title="Inspect live logs"
            >
              <Terminal className="w-3.5 h-3.5 text-zinc-300" />
            </Button>
          </>
        ) : (
          <>
            <Button
              variant="primary"
              size="sm"
              onClick={() => onRunClick(tool)}
              className="flex-1"
            >
              <Play className="w-3 h-3 mr-1.5 fill-current" />
              Run Tool
            </Button>

            {latestScan && (
              <Button
                variant="secondary"
                size="sm"
                onClick={() => onViewLogs(latestScan.id)}
                className="px-2.5"
                title="View logs"
              >
                <Terminal className="w-3.5 h-3.5 text-zinc-400 hover:text-zinc-200" />
              </Button>
            )}
          </>
        )}
      </div>
    </div>
  );
};
