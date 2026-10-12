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

interface MethodologyStep {
  stepNumber: number;
  phase: string;
  category: string;
  objective: string;
  toolNames: string[];
}

const METHODOLOGY_WORKFLOW: MethodologyStep[] = [
  {
    stepNumber: 1,
    phase: 'Subdomain Enumeration',
    category: 'Passive OSINT & Subdomain Discovery',
    objective: 'Systematically discover all subdomains from passive online sources, certificate transparency logs, and threat intelligence feeds.',
    toolNames: ['subfinder', 'assetfinder'],
  },
  {
    stepNumber: 2,
    phase: 'Live Host Probing',
    category: 'HTTP Verification & Technology Detection',
    objective: 'Probe discovered subdomains to identify active web servers, HTTP status codes, titles, and underlying technology stacks.',
    toolNames: ['httpx'],
  },
  {
    stepNumber: 3,
    phase: 'Crawling & URL Discovery',
    category: 'Endpoint & Route Discovery',
    objective: 'Spider live targets and fetch historical archive records to discover dynamic endpoints, API routes, and hidden parameters.',
    toolNames: ['katana', 'gau'],
  },
  {
    stepNumber: 4,
    phase: 'Content Discovery',
    category: 'Directory & Parameter Fuzzing',
    objective: 'Fuzz live endpoints to identify unlinked administrative panels, backup files, hidden directories, and exposed assets.',
    toolNames: ['ffuf'],
  },
  {
    stepNumber: 5,
    phase: 'Vulnerability Assessment',
    category: 'Exposure & Template Scanning',
    objective: 'Run non-intrusive template-based vulnerability scans across live assets to identify misconfigurations and known CVEs.',
    toolNames: ['nuclei'],
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
    <div className="space-y-8 pb-12">
      {/* Scope Header */}
      <div className="border-b border-[#30363d] pb-4">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-base font-semibold text-zinc-100 font-mono">Bug Hunting Methodology Workflow</h2>
            <p className="text-xs text-zinc-400 mt-0.5">
              Target Scope: <span className="font-mono text-zinc-200 font-semibold">{activeTarget.domain}</span>
            </p>
          </div>
        </div>
      </div>

      {/* Sequential Methodology Pipeline */}
      <div className="space-y-8">
        {METHODOLOGY_WORKFLOW.map((step) => {
          const stepTools = tools.filter((t) => step.toolNames.includes(t.name));

          return (
            <section key={step.stepNumber} className="border border-[#30363d] bg-[#0d1117] rounded-md p-5">
              {/* Step Title Header */}
              <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 border-b border-[#21262d] pb-3 mb-4">
                <div>
                  <div className="flex items-center space-x-2">
                    <span className="font-mono text-xs px-2 py-0.5 rounded bg-[#21262d] text-zinc-300 font-semibold border border-[#30363d]">
                      Step {step.stepNumber}
                    </span>
                    <h3 className="text-sm font-bold text-zinc-100 font-mono">
                      {step.phase}
                    </h3>
                    <span className="text-xs text-zinc-500 hidden sm:inline">— {step.category}</span>
                  </div>
                  <p className="text-xs text-zinc-400 mt-1 max-w-3xl leading-relaxed">
                    {step.objective}
                  </p>
                </div>
              </div>

              {/* Tools Grid for this Step */}
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                {stepTools.map((tool) => {
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
            </section>
          );
        })}
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
