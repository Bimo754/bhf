import React, { useState, useMemo } from 'react';
import { ExternalLink, Search, ChevronLeft, ChevronRight, RefreshCw, Copy, Check, Download, Globe } from 'lucide-react';
import { Asset, PaginatedAssets } from '../types/asset';
import { Button } from './Button';

interface AssetTableProps {
  data: PaginatedAssets;
  currentPage: number;
  searchQuery: string;
  onPageChange: (page: number) => void;
  onSearchChange: (search: string) => void;
  onRefresh: () => void;
  isLoading?: boolean;
}

type FilterChip = 'all' | 'priority' | 'live' | '2xx' | '3xx' | 'errors';

export const AssetTable: React.FC<AssetTableProps> = ({
  data,
  currentPage,
  searchQuery,
  onPageChange,
  onSearchChange,
  onRefresh,
  isLoading = false,
}) => {
  const [localSearch, setLocalSearch] = useState(searchQuery);
  const [activeFilter, setActiveFilter] = useState<FilterChip>('all');
  const [copiedAction, setCopiedAction] = useState<'hosts' | 'urls' | null>(null);

  const totalPages = Math.ceil(data.total / data.limit) || 1;

  const handleSearchSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSearchChange(localSearch);
  };

  const filteredAssets = useMemo(() => {
    if (activeFilter === 'all') return data.assets;
    if (activeFilter === 'priority') {
      return data.assets.filter((a) => (a.priority_score && a.priority_score >= 30));
    }
    if (activeFilter === 'live') {
      return data.assets.filter((a) => (a.status_code && a.status_code > 0) || a.url);
    }
    if (activeFilter === '2xx') {
      return data.assets.filter((a) => a.status_code && a.status_code >= 200 && a.status_code < 300);
    }
    if (activeFilter === '3xx') {
      return data.assets.filter((a) => a.status_code && a.status_code >= 300 && a.status_code < 400);
    }
    if (activeFilter === 'errors') {
      return data.assets.filter((a) => a.status_code && a.status_code >= 400);
    }
    return data.assets;
  }, [data.assets, activeFilter]);

  const handleCopyHosts = () => {
    const hosts = Array.from(new Set(data.assets.map((a) => a.host).filter(Boolean))).join('\n');
    navigator.clipboard.writeText(hosts);
    setCopiedAction('hosts');
    setTimeout(() => setCopiedAction(null), 2000);
  };

  const handleCopyURLs = () => {
    const urls = Array.from(
      new Set(
        data.assets
          .map((a) => a.url || (a.status_code ? `https://${a.host}` : null))
          .filter(Boolean)
      )
    ).join('\n');
    navigator.clipboard.writeText(urls);
    setCopiedAction('urls');
    setTimeout(() => setCopiedAction(null), 2000);
  };

  const handleExportJSON = () => {
    const blob = new Blob([JSON.stringify(data.assets, null, 2)], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `bhf_assets_export_p${currentPage}.json`;
    a.click();
    URL.revokeObjectURL(url);
  };

  const getStatusBadge = (code?: number) => {
    if (!code || code === 0) {
      return <span className="text-zinc-600 font-mono text-[11px]">-</span>;
    }
    let style = 'text-zinc-400 bg-zinc-800/80 border-zinc-700';
    if (code >= 200 && code < 300) style = 'text-emerald-400 bg-emerald-950/60 border-emerald-800/80';
    else if (code >= 300 && code < 400) style = 'text-blue-400 bg-blue-950/60 border-blue-800/80';
    else if (code >= 400 && code < 500) style = 'text-amber-400 bg-amber-950/60 border-amber-800/80';
    else if (code >= 500) style = 'text-red-400 bg-red-950/60 border-red-800/80';

    return (
      <span className={`px-1.5 py-0.5 rounded border text-[11px] font-mono font-bold ${style}`}>
        {code}
      </span>
    );
  };

  return (
    <div className="bg-[#161b22] border border-[#30363d] rounded-md overflow-hidden flex flex-col">
      {/* Top Toolbar */}
      <div className="p-3 bg-[#161b22] border-b border-[#30363d] flex flex-col md:flex-row items-stretch md:items-center justify-between gap-3">
        {/* Search Input */}
        <form onSubmit={handleSearchSubmit} className="relative flex-1 max-w-md">
          <Search className="w-3.5 h-3.5 text-zinc-500 absolute left-2.5 top-2.5" />
          <input
            type="text"
            placeholder="Search host, URL, page title, or IP..."
            value={localSearch}
            onChange={(e) => setLocalSearch(e.target.value)}
            className="w-full bg-[#0d1117] border border-[#30363d] rounded pl-8 pr-3 py-1.5 text-xs text-zinc-200 placeholder-zinc-600 focus:outline-none focus:border-zinc-500 font-mono"
          />
        </form>

        {/* Action Buttons */}
        <div className="flex items-center space-x-1.5 text-xs font-mono self-end md:self-auto">
          <button
            onClick={onRefresh}
            title="Refresh assets inventory"
            className="flex items-center space-x-1 px-2.5 py-1.5 rounded bg-[#0d1117] border border-[#30363d] text-zinc-300 hover:text-zinc-100 hover:bg-[#21262d] transition"
          >
            <RefreshCw className={`w-3.5 h-3.5 ${isLoading ? 'animate-spin text-blue-400' : ''}`} />
            <span>Refresh</span>
          </button>

          <button
            onClick={handleCopyHosts}
            title="Copy all distinct hosts to clipboard"
            className="flex items-center space-x-1 px-2.5 py-1.5 rounded bg-[#0d1117] border border-[#30363d] text-zinc-300 hover:text-zinc-100 hover:bg-[#21262d] transition"
          >
            {copiedAction === 'hosts' ? <Check className="w-3.5 h-3.5 text-emerald-400" /> : <Copy className="w-3.5 h-3.5" />}
            <span>Hosts</span>
          </button>

          <button
            onClick={handleCopyURLs}
            title="Copy live URLs to clipboard"
            className="flex items-center space-x-1 px-2.5 py-1.5 rounded bg-[#0d1117] border border-[#30363d] text-zinc-300 hover:text-zinc-100 hover:bg-[#21262d] transition"
          >
            {copiedAction === 'urls' ? <Check className="w-3.5 h-3.5 text-emerald-400" /> : <Globe className="w-3.5 h-3.5" />}
            <span>URLs</span>
          </button>

          <button
            onClick={handleExportJSON}
            title="Export page data as JSON"
            className="p-1.5 rounded bg-[#0d1117] border border-[#30363d] text-zinc-300 hover:text-zinc-100 hover:bg-[#21262d] transition"
          >
            <Download className="w-3.5 h-3.5" />
          </button>
        </div>
      </div>

      {/* Filter Chips Bar */}
      <div className="px-3 py-2 bg-[#0d1117] border-b border-[#21262d] flex items-center justify-between text-xs font-mono">
        <div className="flex items-center space-x-1 overflow-x-auto">
          {(() => {
            const stats = data.stats || {
              total: data.total,
              live: 0,
              code_2xx: 0,
              code_3xx: 0,
              code_4xx_5xx: 0,
            };

            const chips = [
              { key: 'all' as const, label: `All (${stats.total})` },
              { key: 'priority' as const, label: `Priority Targets (${stats.high_priority || 0})` },
              { key: 'live' as const, label: `Live Web (${stats.live})` },
              { key: '2xx' as const, label: `2xx OK (${stats.code_2xx})` },
              { key: '3xx' as const, label: `3xx Redirect (${stats.code_3xx})` },
              { key: 'errors' as const, label: `4xx / 5xx (${stats.code_4xx_5xx})` },
            ];

            return chips.map((chip) => (
              <button
                key={chip.key}
                onClick={() => setActiveFilter(chip.key)}
                className={`px-2 py-0.5 rounded text-[11px] transition-colors shrink-0 ${
                  activeFilter === chip.key
                    ? 'bg-[#21262d] text-zinc-100 font-semibold border border-[#30363d]'
                    : 'text-zinc-400 hover:text-zinc-200 hover:bg-[#161b22]'
                }`}
              >
                {chip.label}
              </button>
            ));
          })()}
        </div>

        <div className="text-zinc-500 text-[11px] hidden sm:block">
          Showing <span className="text-zinc-300 font-semibold">{filteredAssets.length}</span> of{' '}
          <span className="text-zinc-300 font-semibold">{data.total}</span> assets
        </div>
      </div>

      {/* Data Table */}
      <div className="overflow-x-auto flex-1">
        <table className="w-full text-left border-collapse text-xs">
          <thead>
            <tr className="border-b border-[#30363d] bg-[#0d1117] text-zinc-400 uppercase tracking-wider font-mono text-[10px]">
              <th className="py-2.5 px-3">Host & Endpoint</th>
              <th className="py-2.5 px-3">Status</th>
              <th className="py-2.5 px-3">Port / IP</th>
              <th className="py-2.5 px-3">Page Title</th>
              <th className="py-2.5 px-3">Technologies</th>
              <th className="py-2.5 px-3 text-right">Browse</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-[#21262d] font-mono">
            {filteredAssets.length === 0 ? (
              <tr>
                <td colSpan={6} className="py-10 text-center text-zinc-500 italic">
                  {data.total === 0
                    ? 'No assets discovered for this scope yet. Run tools from the Methodology pipeline to populate findings.'
                    : 'No assets match the active filter criteria.'}
                </td>
              </tr>
            ) : (
              filteredAssets.map((asset: Asset) => {
                const liveURL = asset.url || (asset.status_code ? `https://${asset.host}` : null);

                return (
                  <tr key={asset.id} className="hover:bg-[#21262d]/40 transition-colors">
                    {/* Host & URL */}
                    <td className="py-2 px-3 font-semibold text-zinc-200">
                      <div className="flex flex-col max-w-md">
                        <div className="flex items-center space-x-1.5 truncate">
                          {liveURL ? (
                            <a
                              href={liveURL}
                              target="_blank"
                              rel="noopener noreferrer"
                              className="text-zinc-100 hover:text-blue-400 truncate font-semibold"
                            >
                              {asset.host}
                            </a>
                          ) : (
                            <span className="truncate">{asset.host}</span>
                          )}
                          {asset.priority_score && asset.priority_score >= 30 ? (
                            <span className="px-1.5 py-0.2 rounded text-[10px] font-mono font-bold bg-amber-950/70 text-amber-300 border border-amber-700/70 shrink-0">
                              ★ {asset.priority_score}
                            </span>
                          ) : null}
                        </div>
                        {asset.url && asset.url !== `https://${asset.host}` && asset.url !== `http://${asset.host}` && (
                          <span className="text-[11px] text-zinc-500 truncate font-normal">
                            {asset.url}
                          </span>
                        )}
                      </div>
                    </td>

                    {/* Status Code */}
                    <td className="py-2 px-3">{getStatusBadge(asset.status_code)}</td>

                    {/* Port & IP */}
                    <td className="py-2 px-3 text-zinc-400 text-[11px]">
                      <span>{asset.port ? asset.port : 80}</span>
                      {asset.ip && <span className="text-zinc-500 ml-1.5">({asset.ip})</span>}
                    </td>

                    {/* Title */}
                    <td className="py-2 px-3 text-zinc-300 max-w-xs truncate font-sans text-[11px]">
                      {asset.title || <span className="text-zinc-600">-</span>}
                    </td>

                    {/* Technologies */}
                    <td className="py-2 px-3">
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

                    {/* Link Action */}
                    <td className="py-2 px-3 text-right">
                      {liveURL ? (
                        <a
                          href={liveURL}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="inline-flex items-center p-1 rounded hover:bg-[#30363d] text-zinc-400 hover:text-blue-400"
                          title="Open in new browser tab"
                        >
                          <ExternalLink className="w-3.5 h-3.5" />
                        </a>
                      ) : (
                        <span className="text-zinc-700">-</span>
                      )}
                    </td>
                  </tr>
                );
              })
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
