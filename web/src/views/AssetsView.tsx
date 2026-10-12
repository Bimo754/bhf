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
  onRefresh: () => void;
  isLoading?: boolean;
}

export const AssetsView: React.FC<AssetsViewProps> = ({
  activeTarget,
  assetsData,
  currentPage,
  searchQuery,
  onPageChange,
  onSearchChange,
  onRefresh,
  isLoading = false,
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
      <AssetTable
        data={assetsData}
        currentPage={currentPage}
        searchQuery={searchQuery}
        onPageChange={onPageChange}
        onSearchChange={onSearchChange}
        onRefresh={onRefresh}
        isLoading={isLoading}
      />
    </div>
  );
};
