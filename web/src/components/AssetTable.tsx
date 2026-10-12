import React, { useState } from 'react';
import { ExternalLink, Search, ChevronLeft, ChevronRight } from 'lucide-react';
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

  const getStatusBadge = (code?: number) => {
    if (!code) return <span className="text-zinc-600 font-mono">-</span>;
    let style = 'text-zinc-400 bg-zinc-800 border-zinc-700';
    if (code >= 200 && code < 300) style = 'text-emerald-400 bg-emerald-950/60 border-emerald-800';
    else if (code >= 300 && code < 400) style = 'text-blue-400 bg-blue-950/60 border-blue-800';
    else if (code >= 400 && code < 500) style = 'text-amber-400 bg-amber-950/60 border-amber-800';
    else if (code >= 500) style = 'text-red-400 bg-red-950/60 border-red-800';

    return (
      <span className={`px-1.5 py-0.5 rounded border text-[11px] font-mono font-bold ${style}`}>
        {code}
      </span>
    );
  };

  return (
    <div className="bg-[#161b22] border border-[#30363d] rounded-md overflow-hidden flex flex-col">
      {/* Table Toolbar */}
      <div className="p-3 bg-[#161b22] border-b border-[#30363d] flex flex-col sm:flex-row items-center justify-between gap-3">
        <form onSubmit={handleSearchSubmit} className="relative w-full sm:w-80">
          <Search className="w-3.5 h-3.5 text-zinc-500 absolute left-2.5 top-2.5" />
          <input
            type="text"
            placeholder="Search host, title, or IP..."
            value={localSearch}
            onChange={(e) => setLocalSearch(e.target.value)}
            className="w-full bg-[#0d1117] border border-[#30363d] rounded pl-8 pr-3 py-1.5 text-xs text-zinc-200 placeholder-zinc-600 focus:outline-none focus:border-zinc-500 font-mono"
          />
        </form>

        <div className="text-xs text-zinc-400 font-mono">
          Total Discovered: <strong className="text-zinc-100">{data.total}</strong>
        </div>
      </div>

      {/* Data Table */}
      <div className="overflow-x-auto flex-1">
        <table className="w-full text-left border-collapse text-xs">
          <thead>
            <tr className="border-b border-[#30363d] bg-[#0d1117] text-zinc-400 uppercase tracking-wider font-mono text-[10px]">
              <th className="py-2.5 px-3">Host / Subdomain</th>
              <th className="py-2.5 px-3">Status</th>
              <th className="py-2.5 px-3">Port / IP</th>
              <th className="py-2.5 px-3">Page Title</th>
              <th className="py-2.5 px-3">Technologies</th>
              <th className="py-2.5 px-3 text-right">Link</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-[#21262d] font-mono">
            {data.assets.length === 0 ? (
              <tr>
                <td colSpan={6} className="py-8 text-center text-zinc-500 italic">
                  No assets discovered for this scope yet. Run Subdomain Discovery or HTTP Probing in Methodology to begin.
                </td>
              </tr>
            ) : (
              data.assets.map((asset: Asset) => (
                <tr key={asset.id} className="hover:bg-[#21262d]/40 transition-colors">
                  {/* Host */}
                  <td className="py-2.5 px-3 font-semibold text-zinc-200">
                    <span className="truncate max-w-xs block">{asset.host}</span>
                  </td>

                  {/* Status Code */}
                  <td className="py-2.5 px-3">{getStatusBadge(asset.status_code)}</td>

                  {/* Port & IP */}
                  <td className="py-2.5 px-3 text-zinc-400">
                    <span>{asset.port ? asset.port : 80}</span>
                    {asset.ip && <span className="text-zinc-500 ml-1.5">({asset.ip})</span>}
                  </td>

                  {/* Title */}
                  <td className="py-2.5 px-3 text-zinc-300 max-w-xs truncate font-sans">
                    {asset.title || <span className="text-zinc-600">-</span>}
                  </td>

                  {/* Technologies */}
                  <td className="py-2.5 px-3">
                    <div className="flex flex-wrap gap-1 max-w-xs">
                      {asset.technologies && asset.technologies.length > 0 ? (
                        asset.technologies.map((t, idx) => (
                          <span
                            key={idx}
                            className="bg-[#0d1117] border border-[#30363d] text-zinc-300 text-[10px] px-1.5 py-0.5 rounded"
                          >
                            {t}
                          </span>
                        ))
                      ) : (
                        <span className="text-zinc-600">-</span>
                      )}
                    </div>
                  </td>

                  {/* Link */}
                  <td className="py-2.5 px-3 text-right">
                    <a
                      href={asset.url || `https://${asset.host}`}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="inline-flex items-center p-1 rounded hover:bg-[#30363d] text-zinc-400 hover:text-zinc-200"
                      title="Open in browser"
                    >
                      <ExternalLink className="w-3.5 h-3.5" />
                    </a>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {/* Pagination Bar */}
      <div className="p-2.5 bg-[#0d1117] border-t border-[#30363d] flex items-center justify-between text-xs text-zinc-400 font-mono">
        <div>
          Page <strong className="text-zinc-200">{currentPage}</strong> of <strong className="text-zinc-200">{totalPages}</strong>
        </div>

        <div className="flex items-center space-x-1.5">
          <Button
            variant="secondary"
            size="sm"
            onClick={() => onPageChange(currentPage - 1)}
            disabled={currentPage <= 1}
            className="!p-1"
          >
            <ChevronLeft className="w-3.5 h-3.5" />
          </Button>
          <Button
            variant="secondary"
            size="sm"
            onClick={() => onPageChange(currentPage + 1)}
            disabled={currentPage >= totalPages}
            className="!p-1"
          >
            <ChevronRight className="w-3.5 h-3.5" />
          </Button>
        </div>
      </div>
    </div>
  );
};
