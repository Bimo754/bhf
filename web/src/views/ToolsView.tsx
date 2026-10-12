import React, { useState } from 'react';
import { ToolInfo } from '../types/tool';
import { Scan } from '../types/scan';
import { Target } from '../types/target';
import { ToolCard } from '../components/ToolCard';
import { Terminal } from '../components/Terminal';
import { LogEntry } from '../types/log';
import { RunConfigModal } from '../components/RunConfigModal';

interface ToolsViewProps {
  tools: ToolInfo[];
  scans: Scan[];
  activeTarget: Target | null;
  activeScanLogs: LogEntry[];
  currentRunningScanId: string | null;
  onExecuteTool: (toolName: string, customArgs: string[]) => void;
  onViewLogs: (scanId: string) => void;
  onStopScan: (scanId: string) => void;
  onClearLogs: () => void;
}

export const ToolsView: React.FC<ToolsViewProps> = ({
  tools,
  scans,
  activeTarget,
  activeScanLogs,
  currentRunningScanId,
  onExecuteTool,
  onViewLogs,
  onStopScan,
  onClearLogs,
}) => {
  const [selectedToolForConfig, setSelectedToolForConfig] = useState<ToolInfo | null>(null);

  // Group tools by category
  const categories = Array.from(new Set(tools.map((t) => t.category)));

  const getLatestScanForTool = (toolName: string): Scan | undefined => {
    return scans.find((s) => s.tool_name === toolName);
  };

  return (
    <div className="grid grid-cols-1 lg:grid-cols-12 gap-6 h-[calc(100vh-6.5rem)]">
      {/* Left Column: Tool Methodology Grid (7 cols) */}
      <div className="lg:col-span-7 overflow-y-auto pr-2 space-y-6">
        {!activeTarget ? (
          <div className="p-8 text-center bg-dark-800 rounded-xl border border-dark-600/60">
            <h3 className="text-base font-bold text-gray-200">No Target Scope Selected</h3>
            <p className="text-xs text-gray-400 mt-1">Please add or select a target domain above to begin hunting.</p>
          </div>
        ) : (
          categories.map((cat) => {
            const catTools = tools.filter((t) => t.category === cat);
            return (
              <div key={cat} className="space-y-3">
                <div className="flex items-center space-x-2 border-b border-dark-700 pb-2">
                  <div className="w-2 h-2 rounded-full bg-cyber-cyan" />
                  <h3 className="text-sm font-semibold text-gray-200 uppercase tracking-wider font-mono">
                    {cat}
                  </h3>
                  <span className="text-xs text-gray-500 font-mono">({catTools.length})</span>
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  {catTools.map((tool) => {
                    const latestScan = getLatestScanForTool(tool.name);
                    const isRunning = currentRunningScanId === latestScan?.id && latestScan?.status === 'RUNNING';

                    return (
                      <ToolCard
                        key={tool.name}
                        tool={tool}
                        latestScan={latestScan}
                        isRunning={isRunning}
                        onRunClick={(t) => setSelectedToolForConfig(t)}
                        onViewLogs={(scanId) => onViewLogs(scanId)}
                        onStop={(scanId) => onStopScan(scanId)}
                      />
                    );
                  })}
                </div>
              </div>
            );
          })
        )}
      </div>

      {/* Right Column: Embedded Real-time Streaming Terminal (5 cols) */}
      <div className="lg:col-span-5 h-full">
        <Terminal
          logs={activeScanLogs}
          isRunning={!!currentRunningScanId}
          onStop={() => currentRunningScanId && onStopScan(currentRunningScanId)}
          onClear={onClearLogs}
        />
      </div>

      {/* Configuration modal before launch */}
      <RunConfigModal
        isOpen={!!selectedToolForConfig}
        tool={selectedToolForConfig}
        target={activeTarget}
        onClose={() => setSelectedToolForConfig(null)}
        onExecute={onExecuteTool}
      />
    </div>
  );
};
