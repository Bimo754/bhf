import React from 'react';
import { Play, Terminal, Octagon, CheckCircle2, AlertCircle } from 'lucide-react';
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
    <div className="bg-dark-800/80 border border-dark-600/70 hover:border-dark-600 transition-all rounded-xl p-5 flex flex-col justify-between shadow-lg group">
      <div>
        {/* Top Header */}
        <div className="flex items-start justify-between mb-3">
          <div>
            <span className="text-[10px] font-mono uppercase tracking-wider text-cyber-cyan/80 bg-cyber-cyan/10 px-2 py-0.5 rounded border border-cyber-cyan/20">
              {tool.category}
            </span>
            <h4 className="text-base font-bold text-gray-100 mt-2 font-mono flex items-center space-x-2">
              <span>{tool.name}</span>
            </h4>
          </div>

          <div className="flex items-center space-x-1.5 text-xs">
            {tool.is_installed ? (
              <span className="flex items-center text-cyber-green text-[11px] space-x-1" title="Binary installed on system">
                <CheckCircle2 className="w-3.5 h-3.5" />
                <span className="hidden sm:inline">Installed</span>
              </span>
            ) : (
              <span className="flex items-center text-cyber-amber text-[11px] space-x-1" title="Binary not detected in PATH">
                <AlertCircle className="w-3.5 h-3.5" />
                <span className="hidden sm:inline">Missing in PATH</span>
              </span>
            )}
          </div>
        </div>

        {/* Description */}
        <p className="text-xs text-gray-400 line-clamp-2 leading-relaxed mb-4">
          {tool.description}
        </p>

        {/* Status of latest run */}
        {latestScan && (
          <div className="flex items-center justify-between text-xs py-2 px-3 rounded-lg bg-dark-900 border border-dark-700/60 mb-4">
            <span className="text-gray-400 font-mono text-[11px]">Last Run:</span>
            <Badge status={latestScan.status} />
          </div>
        )}
      </div>

      {/* Action Controls */}
      <div className="flex items-center space-x-2 pt-2 border-t border-dark-700/50">
        {isRunning && latestScan ? (
          <>
            <Button
              variant="danger"
              size="sm"
              onClick={() => onStop(latestScan.id)}
              className="flex-1"
            >
              <Octagon className="w-3.5 h-3.5 mr-1.5" />
              Stop Scan
            </Button>
            <Button
              variant="secondary"
              size="sm"
              onClick={() => onViewLogs(latestScan.id)}
              className="px-3"
              title="View live log stream"
            >
              <Terminal className="w-3.5 h-3.5 text-cyber-cyan" />
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
              <Play className="w-3.5 h-3.5 mr-1.5" />
              Run Tool
            </Button>

            {latestScan && (
              <Button
                variant="secondary"
                size="sm"
                onClick={() => onViewLogs(latestScan.id)}
                className="px-3"
                title="View previous execution logs"
              >
                <Terminal className="w-3.5 h-3.5 text-gray-400 group-hover:text-cyber-cyan transition" />
              </Button>
            )}
          </>
        )}
      </div>
    </div>
  );
};
