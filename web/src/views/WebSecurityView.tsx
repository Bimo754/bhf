import React from 'react';
import { Target } from '../types/target';
import { ToolInfo } from '../types/tool';
import { Scan } from '../types/scan';
import { PaginatedAssets } from '../types/asset';
import { PaginatedTraffic, TrafficStats } from '../types/traffic';
import { FilterChip } from '../components/AssetTable';
import { ActiveQueueState } from './ToolsView';
import { ToolsView } from './ToolsView';
import { AssetsView } from './AssetsView';
import { TrafficView } from './TrafficView';

export type WebSecuritySubTab = 'methodology' | 'assets' | 'appintel';

interface WebSecurityViewProps {
  activeSubTab: WebSecuritySubTab;
  onSelectSubTab: (subTab: WebSecuritySubTab) => void;
  activeTarget: Target | null;
  tools: ToolInfo[];
  scans: Scan[];
  currentRunningScanId: string | null;
  activeQueue: ActiveQueueState | null;
  onExecuteTool: (toolName: string, customArgs: string[]) => void;
  onExecuteQueue: (toolNames: string[], title: string) => void;
  onSkipTool: () => void;
  onStopQueue: () => void;
  onViewLogs: (scanId: string) => void;
  onStopScan: (scanId: string) => void;

  // Assets Props
  assetsData: PaginatedAssets;
  assetSearch: string;
  assetFilter: FilterChip;
  assetSort: string;
  assetOrder: 'asc' | 'desc';
  isAssetsLoading: boolean;
  onPageChangeAssets: (page: number) => void;
  onSearchChangeAssets: (search: string) => void;
  onFilterChangeAssets: (filter: FilterChip) => void;
  onSortChangeAssets: (field: string) => void;
  onRefreshAssets: () => void;

  // Traffic / App Intel Props
  trafficData: PaginatedTraffic;
  trafficStats: TrafficStats | null;
  trafficSearch: string;
  trafficCategory: string;
  trafficMethod: string;
  isTrafficLoading: boolean;
  onPageChangeTraffic: (page: number) => void;
  onSearchChangeTraffic: (search: string) => void;
  onCategoryFilterChangeTraffic: (cat: string) => void;
  onMethodFilterChangeTraffic: (method: string) => void;
  onRefreshTraffic: () => void;
  onAssessmentTriggered?: () => void;
}

export const WebSecurityView: React.FC<WebSecurityViewProps> = ({
  activeSubTab,
  onSelectSubTab,
  activeTarget,
  tools,
  scans,
  currentRunningScanId,
  activeQueue,
  onExecuteTool,
  onExecuteQueue,
  onSkipTool,
  onStopQueue,
  onViewLogs,
  onStopScan,
  assetsData,
  assetSearch,
  assetFilter,
  assetSort,
  assetOrder,
  isAssetsLoading,
  onPageChangeAssets,
  onSearchChangeAssets,
  onFilterChangeAssets,
  onSortChangeAssets,
  onRefreshAssets,
  trafficData,
  trafficStats,
  trafficSearch,
  trafficCategory,
  trafficMethod,
  isTrafficLoading,
  onPageChangeTraffic,
  onSearchChangeTraffic,
  onCategoryFilterChangeTraffic,
  onMethodFilterChangeTraffic,
  onRefreshTraffic,
  onAssessmentTriggered,
}) => {
  return (
    <div className="space-y-5">
      {/* Sleek Sub-Discipline Switcher */}
      <div className="flex items-center justify-between border-b border-white/[0.06] pb-3">
        <div className="flex items-center p-0.5 bg-white/[0.02] border border-white/[0.05] rounded-xl text-xs">
          <button
            onClick={() => onSelectSubTab('methodology')}
            className={`px-3 py-1.5 rounded-lg transition-all ${
              activeSubTab === 'methodology'
                ? 'bg-white/[0.08] text-white shadow-sm ring-1 ring-white/10 font-medium'
                : 'text-slate-400 hover:text-slate-200 hover:bg-white/[0.03]'
            }`}
          >
            Recon Pipeline & Tools
          </button>
          <button
            onClick={() => onSelectSubTab('assets')}
            className={`flex items-center space-x-1 px-3 py-1.5 rounded-lg transition-all ${
              activeSubTab === 'assets'
                ? 'bg-white/[0.08] text-white shadow-sm ring-1 ring-white/10 font-medium'
                : 'text-slate-400 hover:text-slate-200 hover:bg-white/[0.03]'
            }`}
          >
            <span>Discovered Assets</span>
            {assetsData.total > 0 && (
              <span className="px-1.5 py-0.2 rounded-full text-[10px] bg-white/[0.06] text-slate-400 font-mono ml-1">
                {assetsData.total}
              </span>
            )}
          </button>
          <button
            onClick={() => onSelectSubTab('appintel')}
            className={`flex items-center space-x-1 px-3 py-1.5 rounded-lg transition-all ${
              activeSubTab === 'appintel'
                ? 'bg-white/[0.08] text-white shadow-sm ring-1 ring-white/10 font-medium'
                : 'text-slate-400 hover:text-slate-200 hover:bg-white/[0.03]'
            }`}
          >
            <span>Traffic & DOM Intel</span>
            {trafficData.total > 0 && (
              <span className="px-1.5 py-0.2 rounded-full text-[10px] bg-white/[0.06] text-slate-400 font-mono ml-1">
                {trafficData.total}
              </span>
            )}
          </button>
        </div>

        <div className="text-xs text-slate-500 font-mono hidden md:block">
          Target: <span className="text-slate-300 font-medium">{activeTarget?.domain || 'None'}</span>
        </div>
      </div>

      {/* Active Sub-Discipline View */}
      {activeSubTab === 'methodology' && (
        <ToolsView
          tools={tools}
          scans={scans}
          activeTarget={activeTarget}
          currentRunningScanId={currentRunningScanId}
          activeQueue={activeQueue}
          onExecuteTool={onExecuteTool}
          onExecuteQueue={onExecuteQueue}
          onSkipTool={onSkipTool}
          onStopQueue={onStopQueue}
          onViewLogs={onViewLogs}
          onStopScan={onStopScan}
        />
      )}

      {activeSubTab === 'assets' && (
        <AssetsView
          activeTarget={activeTarget}
          assetsData={assetsData}
          currentPage={assetsData.page}
          searchQuery={assetSearch}
          activeFilter={assetFilter}
          sortField={assetSort}
          sortOrder={assetOrder}
          onPageChange={onPageChangeAssets}
          onSearchChange={onSearchChangeAssets}
          onFilterChange={onFilterChangeAssets}
          onSortChange={onSortChangeAssets}
          onRefresh={onRefreshAssets}
          onLaunchScan={onExecuteTool}
          isLoading={isAssetsLoading}
        />
      )}

      {activeSubTab === 'appintel' && (
        <TrafficView
          targetId={activeTarget?.id}
          data={trafficData}
          stats={trafficStats}
          currentPage={trafficData.page}
          searchQuery={trafficSearch}
          categoryFilter={trafficCategory}
          methodFilter={trafficMethod}
          onPageChange={onPageChangeTraffic}
          onSearchChange={onSearchChangeTraffic}
          onCategoryFilterChange={onCategoryFilterChangeTraffic}
          onMethodFilterChange={onMethodFilterChangeTraffic}
          onRefresh={onRefreshTraffic}
          onAssessmentTriggered={onAssessmentTriggered}
          isLoading={isTrafficLoading}
        />
      )}
    </div>
  );
};
