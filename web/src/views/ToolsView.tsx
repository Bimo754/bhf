import React, { useState } from 'react';
import { ToolInfo } from '../types/tool';
import { Scan } from '../types/scan';
import { Target } from '../types/target';
import { ToolCard } from '../components/ToolCard';
import { RunConfigModal } from '../components/RunConfigModal';

interface ToolsViewProps {
  tools: ToolInfo[];
  scans: Scan[];
  activeTarget: Target | null;
  currentRunningScanId: string | null;
  onExecuteTool: (toolName: string, customArgs: string[]) => void;
  onViewLogs: (scanId: string) => void;
  onStopScan: (scanId: string) => void;
}

interface MethodologyStage {
  title: string;
  toolNames: string[];
}

const METHODOLOGY_STAGES: MethodologyStage[] = [
  {
    title: 'Subdomain & OSINT Enumeration',
    toolNames: ['subfinder', 'assetfinder', 'theharvester', 'alterx', 'asnmap', 'dnsx'],
  },
  {
    title: 'Port & Network Scanning',
    toolNames: ['naabu'],
  },
  {
    title: 'Live Host Probing',
    toolNames: ['httpx', 'tlsx', 'wafw00f'],
  },
  {
    title: 'Crawling & URL Discovery',
    toolNames: ['katana', 'gau'],
  },
  {
    title: 'Content Discovery',
    toolNames: ['ffuf'],
  },
  {
    title: 'Vulnerability Assessment',
    toolNames: ['subzy', 'nuclei'],
  },
];

export const ToolsView: React.FC<ToolsViewProps> = ({
  tools,
  scans,
  activeTarget,
  currentRunningScanId,
  onExecuteTool,
  onViewLogs,
  onStopScan,
}) => {
  const [selectedToolForConfig, setSelectedToolForConfig] = useState<ToolInfo | null>(null);

  const getLatestScanForTool = (toolName: string): Scan | undefined => {
    return scans.find((s) => s.tool_name === toolName);
  };

  if (!activeTarget) {
    return (
      <div className="p-8 text-center bg-[#161b22] rounded-md border border-[#30363d] max-w-lg mx-auto mt-12">
        <h3 className="text-sm font-semibold text-zinc-200">No Target Selected</h3>
        <p className="text-xs text-zinc-400 mt-1">Please select or add a target domain in the top bar to begin.</p>
      </div>
    );
  }

  return (
    <div className="space-y-6 pb-12">
      {METHODOLOGY_STAGES.map((stage) => {
        const stageTools = tools.filter((t) => stage.toolNames.includes(t.name));
        if (stageTools.length === 0) return null;

        return (
          <section key={stage.title} className="space-y-2.5">
            <h3 className="text-xs font-mono font-bold uppercase tracking-wider text-zinc-400 pl-1">
              {stage.title}
            </h3>

            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-2 xl:grid-cols-3 gap-3">
              {stageTools.map((tool) => {
                const latestScan = getLatestScanForTool(tool.name);
                const isRunning = currentRunningScanId === latestScan?.id && latestScan?.status === 'RUNNING';

                return (
                  <ToolCard
                    key={tool.name}
                    tool={tool}
                    latestScan={latestScan}
                    isRunning={isRunning}
                    onRunClick={(toolName) => onExecuteTool(toolName, [])}
                    onConfigClick={(t) => setSelectedToolForConfig(t)}
                    onViewLogs={(scanId) => onViewLogs(scanId)}
                    onStop={(scanId) => onStopScan(scanId)}
                  />
                );
              })}
            </div>
          </section>
        );
      })}

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
