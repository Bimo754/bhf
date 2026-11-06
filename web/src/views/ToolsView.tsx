import React, { useState, useMemo } from 'react';
import { Play, Square, SkipForward, Layers } from 'lucide-react';
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
  onSkipTool: () => void;
  onStopQueue: () => void;
  onViewLogs: (scanId: string) => void;
  onStopScan: (scanId: string) => void;
}

interface MethodologyStage {
  id: string;
  title: string;
  description: string;
  toolNames: string[];
}

const METHODOLOGY_STAGES: MethodologyStage[] = [
  {
    id: 's1',
    title: 'Subdomain & OSINT Enumeration',
    description: 'Map external perimeter, discover subdomains, and actively resolve DNS records.',
    toolNames: ['subfinder', 'assetfinder', 'theharvester', 'dnsx'],
  },
  {
    id: 's2',
    title: 'Port & Network Scanning',
    description: 'Uncover non-standard open ports and exposed services on live resolving hosts.',
    toolNames: ['naabu'],
  },
  {
    id: 's3',
    title: 'Live HTTP/S Probing & Fingerprinting',
    description: 'Probe web ports, detect tech stack, extract TLS SANs, and capture screenshots.',
    toolNames: ['httpx', 'tlsx', 'wafw00f'],
  },
  {
    id: 's4',
    title: 'Crawling & Endpoint Discovery',
    description: 'Spider dynamic routes, extract archive URLs, and mine JavaScript bundles.',
    toolNames: ['katana', 'gau', 'uro', 'jsluice'],
  },
  {
    id: 's5',
    title: 'Content Discovery & Fuzzing',
    description: 'Brute-force hidden paths and discover user-controlled query parameters.',
    toolNames: ['ffuf', 'arjun'],
  },
  {
    id: 's6',
    title: 'Specialized Vulnerability Assessment',
    description: 'Audits for subdomain takeovers, CORS origin reflection, and cloud bucket exposures.',
    toolNames: ['subzy', 'corsy', 's3scanner'],
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
  onSkipTool,
  onStopQueue,
  onViewLogs,
  onStopScan,
}) => {
  const [selectedToolForConfig, setSelectedToolForConfig] = useState<ToolInfo | null>(null);
  const [enrichAfterCrawl, setEnrichAfterCrawl] = useState(true);
  const [stageFilter, setStageFilter] = useState<string>('all');

  const getLatestScanForTool = (toolName: string): Scan | undefined => {
    return scans.find((s) => s.tool_name === toolName);
  };

  const fullPipelineTools = useMemo(() => {
    const s1 = ['subfinder', 'assetfinder', 'theharvester', 'dnsx'];
    const s2 = ['naabu'];
    const s3 = ['httpx', 'tlsx', 'wafw00f'];
    const s4 = ['katana', 'gau', 'uro', 'jsluice'];
    const reProbe = enrichAfterCrawl ? ['httpx'] : [];
    const s5 = ['ffuf', 'arjun'];
    const s6 = ['subzy', 'corsy', 's3scanner'];

    const raw = [...s1, ...s2, ...s3, ...s4, ...reProbe, ...s5, ...s6];
    return raw.filter((name) => tools.some((t) => t.name === name));
  }, [tools, enrichAfterCrawl]);

  const handleRunFullPipeline = () => {
    if (fullPipelineTools.length === 0) return;
    onExecuteQueue(fullPipelineTools, 'Full Recon Pipeline');
  };

  const visibleStages = useMemo(() => {
    if (stageFilter === 'all') return METHODOLOGY_STAGES;
    return METHODOLOGY_STAGES.filter((s) => s.id === stageFilter);
  }, [stageFilter]);

  if (!activeTarget) {
    return (
      <div className="p-12 text-center bg-[#0e1420] rounded-xl border border-white/[0.06] max-w-md mx-auto my-12">
        <Layers className="w-8 h-8 text-slate-500 mx-auto mb-3" />
        <h3 className="text-sm font-medium text-slate-200">No Target Scope Selected</h3>
        <p className="text-xs text-slate-400 mt-1">Please select or register a target domain above to begin assessment.</p>
      </div>
    );
  }

  return (
    <div className="space-y-5 pb-12">
      {/* Sleek, Eye-Comfortable Pipeline Command Bar */}
      <div className="bg-[#0e1420] border border-white/[0.06] rounded-xl p-4 flex flex-col md:flex-row md:items-center justify-between gap-4">
        {/* Left: Pipeline Scope & Stage Filter */}
        <div className="flex flex-wrap items-center gap-3">
          <div className="flex items-center space-x-2 text-xs">
            <span className="font-semibold text-slate-200">Recon Pipeline</span>
            <span className="text-slate-600">•</span>
            <span className="text-slate-400 font-mono text-[11px]">{fullPipelineTools.length} Tools</span>
          </div>

          <div className="hidden sm:block h-4 w-px bg-white/[0.08]" />

          {/* Discreet Stage Selector */}
          <div className="flex items-center space-x-1.5 text-xs text-slate-400">
            <span>Filter:</span>
            <select
              value={stageFilter}
              onChange={(e) => setStageFilter(e.target.value)}
              className="bg-white/[0.04] border border-white/[0.08] rounded-lg px-2.5 py-1 text-xs text-slate-200 focus:outline-none focus:border-indigo-500 cursor-pointer"
            >
              <option value="all" className="bg-[#0e1420] text-slate-200">All 6 Stages (Complete Pipeline)</option>
              {METHODOLOGY_STAGES.map((st) => (
                <option key={st.id} value={st.id} className="bg-[#0e1420] text-slate-200">
                  {st.title}
                </option>
              ))}
            </select>
          </div>

          <div className="hidden sm:block h-4 w-px bg-white/[0.08]" />

          <label className="flex items-center space-x-2 text-xs text-slate-400 hover:text-slate-300 cursor-pointer select-none">
            <input
              type="checkbox"
              checked={enrichAfterCrawl}
              onChange={(e) => setEnrichAfterCrawl(e.target.checked)}
              className="rounded bg-white/[0.04] border-white/[0.1] text-indigo-500 focus:ring-0 focus:ring-offset-0 cursor-pointer h-3.5 w-3.5"
            >
            </input>
            <span className="text-[11px]">Auto-enrich with httpx</span>
          </label>
        </div>

        {/* Right: Pipeline Controls */}
        <div className="flex items-center space-x-2">
          {activeQueue ? (
            <div className="flex items-center space-x-2">
              <div className="flex items-center space-x-2 bg-indigo-500/10 border border-indigo-500/20 px-3 py-1 rounded-lg text-xs font-mono text-indigo-300">
                <span className="w-1.5 h-1.5 rounded-full bg-indigo-400 animate-pulse" />
                <span>
                  {activeQueue.currentIndex + 1}/{activeQueue.tools.length}: <strong className="text-white">{activeQueue.tools[activeQueue.currentIndex]}</strong>
                </span>
              </div>

              <Button
                variant="secondary"
                size="sm"
                onClick={onSkipTool}
                className="text-xs font-mono"
                title={`Skip ${activeQueue.tools[activeQueue.currentIndex]} and advance`}
              >
                <SkipForward className="w-3 h-3 mr-1 text-amber-400" />
                <span>Skip</span>
              </Button>

              <Button
                variant="danger"
                size="sm"
                onClick={onStopQueue}
                className="text-xs font-mono"
              >
                <Square className="w-3 h-3 mr-1 fill-current" />
                <span>Stop</span>
              </Button>
            </div>
          ) : (
            <Button
              variant="primary"
              size="sm"
              disabled={!!currentRunningScanId}
              onClick={handleRunFullPipeline}
              className="text-xs font-medium"
            >
              <Play className="w-3 h-3 mr-1.5 fill-current" />
              <span>Run Full Recon Sequence</span>
            </Button>
          )}
        </div>
      </div>

      {/* Methodology Stages */}
      <div className="space-y-4">
        {visibleStages.map((stage) => {
          const stageTools = stage.toolNames
            .map((name) => tools.find((t) => t.name === name))
            .filter((t): t is ToolInfo => !!t);

          if (stageTools.length === 0) return null;

          const isStageRunning = activeQueue?.title === stage.title;

          return (
            <div
              key={stage.id}
              className="bg-[#0e1420] border border-white/[0.06] rounded-xl p-4 space-y-3"
            >
              {/* Stage Header */}
              <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 border-b border-white/[0.04] pb-2.5">
                <div>
                  <h3 className="text-xs font-semibold text-slate-200">
                    {stage.title}
                  </h3>
                  <p className="text-[11px] text-slate-400 mt-0.5">
                    {stage.description}
                  </p>
                </div>

                <div className="flex items-center space-x-2 shrink-0">
                  {isStageRunning ? (
                    <div className="flex items-center space-x-1.5 text-xs text-indigo-400 font-mono">
                      <span className="w-1.5 h-1.5 rounded-full bg-indigo-400 animate-pulse" />
                      <span>Stage running...</span>
                    </div>
                  ) : (
                    <Button
                      variant="ghost"
                      size="xs"
                      disabled={!!currentRunningScanId || !!activeQueue}
                      onClick={() => onExecuteQueue(stageTools.map((t) => t.name), stage.title)}
                      className="text-slate-400 hover:text-slate-200 text-xs font-mono"
                      title={`Run only the ${stageTools.length} tools in this stage`}
                    >
                      <Play className="w-2.5 h-2.5 mr-1 fill-current" />
                      <span>Run Stage ({stageTools.length})</span>
                    </Button>
                  )}
                </div>
              </div>

              {/* Clean Tool Rows */}
              <div className="grid grid-cols-1 md:grid-cols-2 gap-2">
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
            </div>
          );
        })}
      </div>

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
