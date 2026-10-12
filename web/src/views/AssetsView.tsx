import React from 'react';
import { Target } from '../types/target';
import { PaginatedAssets } from '../types/asset';
import { AssetTable } from '../components/AssetTable';

interface AssetsViewProps {
  activeTarget: Target | null;
  assetsData: PaginatedAssets;
  currentPage: number;
  searchQuery: string;
  onPageChange: (page: number) => void;
  onSearchChange: (query: string) => void;
}

export const AssetsView: React.FC<AssetsViewProps> = ({
  activeTarget,
  assetsData,
  currentPage,
  searchQuery,
  onPageChange,
  onSearchChange,
}) => {
  if (!activeTarget) {
    return (
      <div className="p-8 text-center bg-[#161b22] rounded-md border border-[#30363d] max-w-lg mx-auto mt-12">
        <h3 className="text-sm font-semibold text-zinc-200">No Target Selected</h3>
        <p className="text-xs text-zinc-400 mt-1">Please select a target scope to view assets.</p>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between border-b border-[#30363d] pb-3">
        <div>
          <h2 className="text-base font-semibold text-zinc-100 font-mono">Discovered Assets Inventory</h2>
          <p className="text-xs text-zinc-400 mt-0.5">
            Scope: <span className="font-mono text-zinc-200 font-semibold">{activeTarget.domain}</span>
          </p>
        </div>
      </div>

      <AssetTable
        data={assetsData}
        currentPage={currentPage}
        searchQuery={searchQuery}
        onPageChange={onPageChange}
        onSearchChange={onSearchChange}
      />
    </div>
  );
};
