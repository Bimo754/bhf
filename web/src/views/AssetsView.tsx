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
      <div className="p-12 text-center bg-[#0e1420] rounded-xl border border-white/[0.06] max-w-md mx-auto my-12">
        <h3 className="text-sm font-semibold text-slate-200">No Target Selected</h3>
        <p className="text-xs text-slate-400 mt-1">Please select a target scope to view assets.</p>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {/* View Switcher Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center p-0.5 bg-white/[0.02] border border-white/[0.05] rounded-xl text-xs">
          <button
            type="button"
            onClick={() => setViewMode('table')}
            className={`flex items-center space-x-1.5 px-3 py-1.5 rounded-lg transition-all ${
              viewMode === 'table'
                ? 'bg-white/[0.08] text-white shadow-sm ring-1 ring-white/10 font-medium'
                : 'text-slate-400 hover:text-slate-200 hover:bg-white/[0.03]'
            }`}
          >
            <LayoutList className="w-3.5 h-3.5" />
            <span>Table View</span>
          </button>

          <button
            type="button"
            onClick={() => setViewMode('gallery')}
            className={`flex items-center space-x-1.5 px-3 py-1.5 rounded-lg transition-all ${
              viewMode === 'gallery'
                ? 'bg-white/[0.08] text-white shadow-sm ring-1 ring-white/10 font-medium'
                : 'text-slate-400 hover:text-slate-200 hover:bg-white/[0.03]'
            }`}
          >
            <LayoutGrid className="w-3.5 h-3.5" />
            <span>Visual Gallery</span>
          </button>

          <button
            type="button"
            onClick={() => setViewMode('sitemap')}
            className={`flex items-center space-x-1.5 px-3 py-1.5 rounded-lg transition-all ${
              viewMode === 'sitemap'
                ? 'bg-white/[0.08] text-white shadow-sm ring-1 ring-white/10 font-medium'
                : 'text-slate-400 hover:text-slate-200 hover:bg-white/[0.03]'
            }`}
          >
            <Network className="w-3.5 h-3.5" />
            <span>Sitemap Tree</span>
          </button>
        </div>

        <div className="text-xs text-slate-500 font-mono hidden sm:block">
          Scope: <strong className="text-slate-300 font-medium">{activeTarget.domain}</strong>
        </div>
      </div>

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
