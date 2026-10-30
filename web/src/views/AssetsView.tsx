import React, { useState } from 'react';
import { LayoutList, LayoutGrid, Network } from 'lucide-react';
import { Target } from '../types/target';
import { PaginatedAssets } from '../types/asset';
import { AssetTable, FilterChip } from '../components/AssetTable';
import { VisualGallery } from '../components/VisualGallery';
import { SitemapTree } from '../components/SitemapTree';

interface AssetsViewProps {
  activeTarget: Target | null;
  assetsData: PaginatedAssets;
  currentPage: number;
  searchQuery: string;
  activeFilter: FilterChip;
  sortField: string;
  sortOrder: 'asc' | 'desc';
  onPageChange: (page: number) => void;
  onSearchChange: (query: string) => void;
  onFilterChange: (filter: FilterChip) => void;
  onSortChange: (field: string) => void;
  onRefresh: () => void;
  onLaunchScan?: (toolName: string, customArgs: string[]) => void;
  isLoading?: boolean;
}

export const AssetsView: React.FC<AssetsViewProps> = ({
  activeTarget,
  assetsData,
  currentPage,
  searchQuery,
  activeFilter,
  sortField,
  sortOrder,
  onPageChange,
  onSearchChange,
  onFilterChange,
  onSortChange,
  onRefresh,
  onLaunchScan,
  isLoading = false,
}) => {
  const [viewMode, setViewMode] = useState<'table' | 'gallery' | 'sitemap'>('table');

  if (!activeTarget) {
    return (
      <div className="p-8 text-center bg-[#161b22] rounded-md border border-[#30363d] max-w-lg mx-auto mt-12 font-mono">
        <h3 className="text-sm font-semibold text-zinc-200">No Target Selected</h3>
        <p className="text-xs text-zinc-400 mt-1">Please select a target scope to view assets.</p>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {/* View Switcher Header */}
      <div className="flex items-center justify-between font-mono">
        <div className="flex items-center space-x-1 bg-[#161b22] p-1 rounded-md border border-[#30363d] text-xs">
          <button
            type="button"
            onClick={() => setViewMode('table')}
            className={`flex items-center space-x-1.5 px-3 py-1 rounded transition ${
              viewMode === 'table'
                ? 'bg-[#21262d] text-zinc-100 font-semibold border border-[#30363d]'
                : 'text-zinc-400 hover:text-zinc-200'
            }`}
          >
            <LayoutList className="w-3.5 h-3.5" />
            <span>Table View</span>
          </button>

          <button
            type="button"
            onClick={() => setViewMode('gallery')}
            className={`flex items-center space-x-1.5 px-3 py-1 rounded transition ${
              viewMode === 'gallery'
                ? 'bg-[#21262d] text-zinc-100 font-semibold border border-[#30363d]'
                : 'text-zinc-400 hover:text-zinc-200'
            }`}
          >
            <LayoutGrid className="w-3.5 h-3.5" />
            <span>Visual Gallery</span>
          </button>

          <button
            type="button"
            onClick={() => setViewMode('sitemap')}
            className={`flex items-center space-x-1.5 px-3 py-1 rounded transition ${
              viewMode === 'sitemap'
                ? 'bg-[#21262d] text-zinc-100 font-semibold border border-[#30363d]'
                : 'text-zinc-400 hover:text-zinc-200'
            }`}
          >
            <Network className="w-3.5 h-3.5" />
            <span>Sitemap Tree</span>
          </button>
        </div>

        <div className="text-xs text-zinc-500 hidden sm:block">
          Active Scope: <strong className="text-zinc-300">{activeTarget.domain}</strong>
        </div>
      </div>

      {/* Render Active View Container */}
      {viewMode === 'table' && (
        <AssetTable
          data={assetsData}
          currentPage={currentPage}
          searchQuery={searchQuery}
          activeFilter={activeFilter}
          sortField={sortField}
          sortOrder={sortOrder}
          onPageChange={onPageChange}
          onSearchChange={onSearchChange}
          onFilterChange={onFilterChange}
          onSortChange={onSortChange}
          onRefresh={onRefresh}
          isLoading={isLoading}
        />
      )}

      {viewMode === 'gallery' && (
        <VisualGallery
          assets={assetsData.assets}
          onLaunchScan={onLaunchScan}
        />
      )}

      {viewMode === 'sitemap' && (
        <SitemapTree
          targetId={activeTarget.id}
          targetDomain={activeTarget.domain}
          fallbackAssets={assetsData.assets}
        />
      )}
    </div>
  );
};
