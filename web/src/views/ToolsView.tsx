import React, { useState, useMemo } from 'react';
import { Play, Square } from 'lucide-react';
import { ToolInfo } from '../types/tool';
import { Scan } from '../types/scan';
import { Target } from '../types/target';
import { ToolCard } from '../components/ToolCard';
import { RunConfigModal } from '../components/RunConfigModal';
import { Button } from '../components/Button';

export interface ActiveQueueState {
  title: string;
  tools: string[];
  currentIndex: number;
}

interface ToolsViewProps {
  tools: ToolInfo[];
  scans: Scan[];
  activeTarget: Target | null;
  currentRunningScanId: string | null;
  activeQueue: ActiveQueueState | null;
  onExecuteTool: (toolName: string, customArgs: string[]) => void;
  onExecuteQueue: (toolNames: string[], title: string) => void;
  onStopQueue: () => void;
  onViewLogs: (scanId: string) => void;
  onStopScan: (scanId: string) => void;
}

interface MethodologyStage {
  title: string;
  description: string;
  toolNames: string[];
}

const METHODOLOGY_STAGES: MethodologyStage[] = [
  {
    title: 'Subdomain & OSINT Enumeration',
    description: 'Map external perimeter, discover subdomains, and actively resolve DNS records.',
    toolNames: ['subfinder', 'assetfinder', 'theharvester', 'dnsx'],
  },
  {
    title: 'Port & Network Scanning',
    description: 'Uncover non-standard open ports and exposed services on live resolving hosts.',
    toolNames: ['naabu'],
  },
  {
    title: 'Live HTTP/S Probing & Fingerprinting',
    description: 'Probe web ports, detect tech stack, extract TLS SANs, and capture screenshots.',
    toolNames: ['httpx', 'tlsx', 'wafw00f', 'gowitness'],
  },
  {
    title: 'Crawling & Endpoint Discovery',
    description: 'Spider dynamic routes, extract archive URLs, and mine JavaScript bundles.',
    toolNames: ['katana', 'gau', 'uro', 'jsluice'],
  },
  {
    title: 'Content Discovery & Fuzzing',
    description: 'Brute-force hidden paths, discover query parameters, and test 401/403 bypasses.',
    toolNames: ['ffuf', 'arjun', 'nomore403'],
  },
  {
    title: 'Vulnerability Assessment & Verification',
    description: 'High-signal vulnerability testing with verified, executable cURL reproductions.',
    toolNames: ['subzy', 'corsy', 's3scanner', 'nuclei'],
  },
];

export const ToolsView: React.FC<ToolsViewProps> = ({
  tools,
  scans,
  activeTarget,
  currentRunningScanId,
  activeQueue,
  onExecuteTool,
  onExecuteQueue,
  onStopQueue,
  onViewLogs,
  onStopScan,
}) => {
  const [selectedToolForConfig, setSelectedToolForConfig] = useState<ToolInfo | null>(null);
  const [enrichAfterCrawl, setEnrichAfterCrawl] = useState(true);

  const getLatestScanForTool = (toolName: string): Scan | undefined => {
    return scans.find((s) => s.tool_name === toolName);
  };

  const fullPipelineTools = useMemo(() => {
    const s1 = ['subfinder', 'assetfinder', 'theharvester', 'dnsx'];
    const s2 = ['naabu'];
    const s3 = ['httpx', 'tlsx', 'wafw00f', 'gowitness'];
    const s4 = ['katana', 'gau', 'uro', 'jsluice'];
    const reProbe = enrichAfterCrawl ? ['httpx'] : [];
    const s5 = ['ffuf', 'arjun', 'nomore403'];
    const s6 = ['subzy', 'corsy', 's3scanner', 'nuclei'];

    const raw = [...s1, ...s2, ...s3, ...s4, ...reProbe, ...s5, ...s6];
    return raw.filter((name) => tools.some((t) => t.name === name));
  }, [tools, enrichAfterCrawl]);

  const isFullPipelineRunning = activeQueue?.title === 'Full Recon Pipeline';

  const handleRunFullPipeline = () => {
    if (fullPipelineTools.length === 0) return;
    onExecuteQueue(fullPipelineTools, 'Full Recon Pipeline');
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
      {/* Sleek Minimal Pipeline Action Bar */}
      <div className="flex flex-wrap items-center justify-between gap-3 px-3.5 py-2.5 bg-[#161b22] border border-[#30363d] rounded-md font-mono text-xs">
        <div className="flex items-center space-x-3">
          <div className="flex items-center space-x-2 text-zinc-400">
            <span className="text-zinc-200 font-bold uppercase tracking-wider text-[11px]">Pipeline</span>
            <span className="text-zinc-600">/</span>
            <span>6 Stages</span>
            <span className="text-zinc-600">&bull;</span>
            <span className="text-zinc-300">{fullPipelineTools.length} Tools</span>
          </div>

          <span className="text-zinc-700">|</span>

          <label className="flex items-center space-x-1.5 cursor-pointer text-zinc-400 hover:text-zinc-200 select-none text-[11px]">
            <input
              type="checkbox"
              checked={enrichAfterCrawl}
              onChange={(e) => setEnrichAfterCrawl(e.target.checked)}
              className="rounded bg-[#0d1117] border-[#30363d] text-emerald-500 focus:ring-0 focus:ring-offset-0 cursor-pointer h-3.5 w-3.5"
            />
            <span>Auto-enrich with httpx</span>
          </label>
        </div>

        <div className="flex items-center space-x-2">
          {isFullPipelineRunning ? (
            <div className="flex items-center space-x-2">
              <span className="inline-flex items-center space-x-1.5 px-2 py-0.5 rounded text-[11px] font-mono bg-blue-500/10 text-blue-400 border border-blue-500/30">
                <span className="w-1.5 h-1.5 rounded-full bg-blue-400 animate-pulse" />
                <span>
                  {activeQueue!.currentIndex + 1}/{activeQueue!.tools.length}: <strong className="text-white">{activeQueue!.tools[activeQueue!.currentIndex]}</strong>
                </span>
              </span>
              <Button
                variant="danger"
                size="sm"
                onClick={onStopQueue}
                className="!py-0.5 !px-2.5 text-xs font-mono"
              >
                <Square className="w-3 h-3 mr-1 fill-current" />
                Stop Pipeline
              </Button>
            </div>
          ) : (
            <Button
              variant="primary"
              size="sm"
              disabled={!!currentRunningScanId || !!activeQueue}
              onClick={handleRunFullPipeline}
              className="!py-1 !px-3 text-xs font-mono font-medium shadow-sm"
            >
              <Play className="w-3 h-3 mr-1.5 fill-current text-white" />
              Run Full Pipeline
            </Button>
          )}
        </div>
      </div>

      {METHODOLOGY_STAGES.map((stage) => {
        // Deterministic ordering: iterate over stage.toolNames so tools display in exact progressive order
        const stageTools = stage.toolNames
          .map((name) => tools.find((t) => t.name === name))
          .filter((t): t is ToolInfo => !!t);

        if (stageTools.length === 0) return null;

        const isStageRunning = activeQueue?.title === stage.title;

        return (
          <section key={stage.title} className="space-y-2.5">
            <div className="flex items-center justify-between pl-1 pr-1">
              <div className="flex items-center space-x-2">
                <h3 className="text-xs font-mono font-bold uppercase tracking-wider text-zinc-300">
                  {stage.title}
                </h3>
                <span className="text-[11px] font-mono text-zinc-500">
                  ({stageTools.length} {stageTools.length === 1 ? 'tool' : 'tools'})
                </span>
              </div>

              <div className="flex items-center space-x-2">
                {isStageRunning ? (
                  <div className="flex items-center space-x-2">
                    <span className="inline-flex items-center space-x-1.5 px-2 py-0.5 rounded text-[11px] font-mono bg-blue-500/10 text-blue-400 border border-blue-500/30">
                      <span className="w-1.5 h-1.5 rounded-full bg-blue-400 animate-pulse" />
                      <span>
                        Stage Running ({activeQueue!.currentIndex + 1}/{activeQueue!.tools.length}):{' '}
                        <strong className="text-white">{activeQueue!.tools[activeQueue!.currentIndex]}</strong>
                      </span>
                    </span>
                    <Button
                      variant="danger"
                      size="sm"
                      onClick={onStopQueue}
                      className="!py-0.5 !px-2 text-xs font-mono"
                    >
                      <Square className="w-3 h-3 mr-1 fill-current" />
                      Stop Stage
                    </Button>
                  </div>
                ) : (
                  <Button
                    variant="secondary"
                    size="sm"
                    disabled={!!currentRunningScanId || !!activeQueue}
                    onClick={() => onExecuteQueue(stageTools.map((t) => t.name), stage.title)}
                    className="!py-0.5 !px-2.5 text-xs font-mono text-zinc-300 hover:text-white"
                    title={`Execute all ${stageTools.length} tools in ${stage.title} sequentially in order`}
                  >
                    <Play className="w-3 h-3 mr-1 fill-current text-emerald-400" />
                    Run Stage
                  </Button>
                )}
              </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-2 xl:grid-cols-3 gap-3">
              {stageTools.map((tool, idx) => {
                const latestScan = getLatestScanForTool(tool.name);
                const isRunning = currentRunningScanId === latestScan?.id && latestScan?.status === 'RUNNING';

                return (
                  <ToolCard
                    key={tool.name}
                    tool={tool}
                    latestScan={latestScan}
                    isRunning={isRunning}
                    stepNumber={idx + 1}
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
