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
      <div className="p-8 text-center bg-dark-800 rounded-xl border border-dark-600/60 max-w-lg mx-auto mt-12">
        <h3 className="text-base font-bold text-gray-200">No Target Scope Selected</h3>
        <p className="text-xs text-gray-400 mt-1">Please select or create a target to view discovered assets.</p>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-lg font-bold text-gray-100 font-mono">Discovered Attack Surface</h2>
          <p className="text-xs text-gray-400">
            Scope target: <span className="text-cyber-cyan font-mono font-medium">{activeTarget.domain}</span>
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
