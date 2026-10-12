import React, { useState } from 'react';
import { ExternalLink, Search, ChevronLeft, ChevronRight, Globe, Server } from 'lucide-react';
import { Asset, PaginatedAssets } from '../types/asset';
import { Button } from './Button';

interface AssetTableProps {
  data: PaginatedAssets;
  currentPage: number;
  searchQuery: string;
  onPageChange: (page: number) => void;
  onSearchChange: (search: string) => void;
}

export const AssetTable: React.FC<AssetTableProps> = ({
  data,
  currentPage,
  searchQuery,
  onPageChange,
  onSearchChange,
}) => {
  const [localSearch, setLocalSearch] = useState(searchQuery);
  const totalPages = Math.ceil(data.total / data.limit) || 1;

  const handleSearchSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSearchChange(localSearch);
  };

  const getStatusColor = (code?: number) => {
    if (!code) return 'text-gray-500';
    if (code >= 200 && code < 300) return 'text-cyber-green bg-cyber-green/10 border-cyber-green/30';
    if (code >= 300 && code < 400) return 'text-cyber-cyan bg-cyber-cyan/10 border-cyber-cyan/30';
    if (code >= 400 && code < 500) return 'text-cyber-amber bg-cyber-amber/10 border-cyber-amber/30';
    return 'text-cyber-rose bg-cyber-rose/10 border-cyber-rose/30';
  };

  return (
    <div className="bg-dark-800 border border-dark-600/70 rounded-xl overflow-hidden shadow-xl flex flex-col">
      {/* Table Controls Header */}
      <div className="p-4 bg-dark-800/90 border-b border-dark-600/50 flex flex-col sm:flex-row items-center justify-between gap-3">
        <form onSubmit={handleSearchSubmit} className="relative w-full sm:w-80">
          <Search className="w-4 h-4 text-gray-500 absolute left-3 top-2.5" />
          <input
            type="text"
            placeholder="Search host, title, or IP..."
            value={localSearch}
            onChange={(e) => setLocalSearch(e.target.value)}
            className="w-full bg-dark-900 border border-dark-600 rounded-lg pl-9 pr-3 py-1.5 text-xs text-gray-100 placeholder-gray-500 focus:outline-none focus:border-cyber-cyan font-mono"
          />
        </form>

        <div className="flex items-center space-x-2 text-xs text-gray-400">
          <span>
            Total Discovered: <strong className="text-gray-200 font-mono">{data.total}</strong>
          </span>
        </div>
      </div>

      {/* Table Data */}
      <div className="overflow-x-auto flex-1">
        <table className="w-full text-left border-collapse text-xs">
          <thead>
            <tr className="border-b border-dark-600/50 bg-dark-900/50 text-gray-400 uppercase tracking-wider font-mono text-[10px]">
              <th className="py-3 px-4">Host / Endpoint</th>
              <th className="py-3 px-4">Status</th>
              <th className="py-3 px-4">Port / IP</th>
              <th className="py-3 px-4">Page Title</th>
              <th className="py-3 px-4">Technologies</th>
              <th className="py-3 px-4 text-right">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-dark-700/50 font-mono">
            {data.assets.length === 0 ? (
              <tr>
                <td colSpan={6} className="py-8 text-center text-gray-500 italic">
                  No assets discovered yet for this scope. Run a tool above to start discovering subdomains!
                </td>
              </tr>
            ) : (
              data.assets.map((asset: Asset) => (
                <tr key={asset.id} className="hover:bg-dark-700/30 transition">
                  {/* Host */}
                  <td className="py-3 px-4 font-semibold text-gray-200">
                    <div className="flex items-center space-x-2">
                      <Globe className="w-3.5 h-3.5 text-cyber-cyan shrink-0" />
                      <span className="truncate max-w-xs">{asset.host}</span>
                    </div>
                  </td>

                  {/* Status Code */}
                  <td className="py-3 px-4">
                    {asset.status_code ? (
                      <span className={`px-2 py-0.5 rounded border text-[11px] font-bold ${getStatusColor(asset.status_code)}`}>
                        {asset.status_code}
                      </span>
                    ) : (
                      <span className="text-gray-600">-</span>
                    )}
                  </td>

                  {/* Port & IP */}
                  <td className="py-3 px-4 text-gray-400">
                    <div className="flex items-center space-x-1">
                      <Server className="w-3 h-3 text-gray-500" />
                      <span>{asset.port ? asset.port : 80}</span>
                      {asset.ip && <span className="text-gray-500">({asset.ip})</span>}
                    </div>
                  </td>

                  {/* Title */}
                  <td className="py-3 px-4 text-gray-300 max-w-xs truncate font-sans">
                    {asset.title || <span className="text-gray-600 italic">N/A</span>}
                  </td>

                  {/* Technologies */}
                  <td className="py-3 px-4">
                    <div className="flex flex-wrap gap-1 max-w-xs">
                      {asset.technologies && asset.technologies.length > 0 ? (
                        asset.technologies.map((t, idx) => (
                          <span
                            key={idx}
                            className="bg-dark-900 border border-dark-600 text-gray-300 text-[10px] px-1.5 py-0.5 rounded"
                          >
                            {t}
                          </span>
                        ))
                      ) : (
                        <span className="text-gray-600">-</span>
                      )}
                    </div>
                  </td>

                  {/* Actions */}
                  <td className="py-3 px-4 text-right">
                    {asset.url ? (
                      <a
                        href={asset.url}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="inline-flex items-center p-1.5 rounded hover:bg-dark-600 text-cyber-cyan transition"
                        title="Open in new browser tab"
                      >
                        <ExternalLink className="w-3.5 h-3.5" />
                      </a>
                    ) : (
                      <a
                        href={`https://${asset.host}`}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="inline-flex items-center p-1.5 rounded hover:bg-dark-600 text-gray-400 hover:text-cyber-cyan transition"
                        title="Open https://"
                      >
                        <ExternalLink className="w-3.5 h-3.5" />
                      </a>
                    )}
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {/* Pagination Footer */}
      <div className="p-3 bg-dark-900/80 border-t border-dark-600/50 flex items-center justify-between text-xs text-gray-400">
        <div>
          Page <strong className="text-gray-200">{currentPage}</strong> of <strong className="text-gray-200">{totalPages}</strong>
        </div>

        <div className="flex items-center space-x-2">
          <Button
            variant="secondary"
            size="sm"
            onClick={() => onPageChange(currentPage - 1)}
            disabled={currentPage <= 1}
            className="!p-1.5"
          >
            <ChevronLeft className="w-4 h-4" />
          </Button>
          <Button
            variant="secondary"
            size="sm"
            onClick={() => onPageChange(currentPage + 1)}
            disabled={currentPage >= totalPages}
            className="!p-1.5"
          >
            <ChevronRight className="w-4 h-4" />
          </Button>
        </div>
      </div>
    </div>
  );
};
