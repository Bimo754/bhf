import React, { useState } from 'react';
import { Search, ChevronDown, ChevronUp, Copy, Check, ExternalLink, RefreshCw, ShieldCheck } from 'lucide-react';
import { Finding, PaginatedFindings, FindingSeverity } from '../types/finding';
import { Button } from '../components/Button';

interface FindingsViewProps {
  data: PaginatedFindings;
  currentPage: number;
  searchQuery: string;
  severityFilter: string;
  onPageChange: (page: number) => void;
  onSearchChange: (search: string) => void;
  onSeverityFilterChange: (severity: string) => void;
  onRefresh: () => void;
  isLoading?: boolean;
}

export const FindingsView: React.FC<FindingsViewProps> = ({
  data,
  currentPage,
  searchQuery,
  severityFilter,
  onPageChange,
  onSearchChange,
  onSeverityFilterChange,
  onRefresh,
  isLoading = false,
}) => {
  const [localSearch, setLocalSearch] = useState(searchQuery);
  const [expandedId, setExpandedId] = useState<string | null>(null);
  const [copiedId, setCopiedId] = useState<string | null>(null);

  const totalPages = Math.ceil(data.total / data.limit) || 1;

  const handleSearchSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSearchChange(localSearch);
  };

  const getSeverityBadge = (severity: FindingSeverity) => {
    switch (severity.toLowerCase()) {
      case 'critical':
        return (
          <span className="inline-flex items-center space-x-1.5 px-2 py-0.5 rounded text-[11px] font-mono bg-rose-500/10 text-rose-300 border border-rose-500/20 font-medium">
            <span className="w-1.5 h-1.5 rounded-full bg-rose-400" />
            <span>Critical</span>
          </span>
        );
      case 'high':
        return (
          <span className="inline-flex items-center space-x-1.5 px-2 py-0.5 rounded text-[11px] font-mono bg-orange-500/10 text-orange-300 border border-orange-500/20 font-medium">
            <span className="w-1.5 h-1.5 rounded-full bg-orange-400" />
            <span>High</span>
          </span>
        );
      case 'medium':
        return (
          <span className="inline-flex items-center space-x-1.5 px-2 py-0.5 rounded text-[11px] font-mono bg-amber-500/10 text-amber-300 border border-amber-500/20 font-medium">
            <span className="w-1.5 h-1.5 rounded-full bg-amber-400" />
            <span>Medium</span>
          </span>
        );
      case 'low':
        return (
          <span className="inline-flex items-center space-x-1.5 px-2 py-0.5 rounded text-[11px] font-mono bg-blue-500/10 text-blue-300 border border-blue-500/20 font-medium">
            <span className="w-1.5 h-1.5 rounded-full bg-blue-400" />
            <span>Low</span>
          </span>
        );
      default:
        return (
          <span className="inline-flex items-center space-x-1.5 px-2 py-0.5 rounded text-[11px] font-mono bg-white/[0.04] text-slate-400 border border-white/[0.06]">
            <span className="w-1.5 h-1.5 rounded-full bg-slate-500" />
            <span>Info</span>
          </span>
        );
    }
  };

  const handleCopyCurl = (id: string, curl: string) => {
    navigator.clipboard.writeText(curl);
    setCopiedId(id);
    setTimeout(() => setCopiedId(null), 2000);
  };

  const handleCopyReport = (f: Finding) => {
    const report = `## Summary
**Vulnerability**: ${f.name}
**Severity**: ${f.severity.toUpperCase()}
**Asset**: ${f.matched_at}
**Tool**: ${f.tool_name}

## Description
${f.description || 'Vulnerability detected during automated reconnaissance assessment.'}

## Steps to Reproduce
Run the following command in terminal:
\`\`\`bash
${f.curl_command || `curl -i '${f.matched_at}'`}
\`\`\`

## Impact
Exposed vulnerability could allow unauthorized access or security policy violation.
`;
    navigator.clipboard.writeText(report);
    setCopiedId(`report-${f.id}`);
    setTimeout(() => setCopiedId(null), 2000);
  };

  return (
    <div className="space-y-4">
      {/* Top Bar: Severity Filters & Search */}
      <div className="bg-[#0e1420] border border-white/[0.06] rounded-xl p-3 flex flex-col md:flex-row md:items-center justify-between gap-3">
        {/* Severity Filter Chips */}
        <div className="flex flex-wrap items-center gap-1.5 text-xs font-mono">
          <button
            onClick={() => onSeverityFilterChange('all')}
            className={`px-2.5 py-1 rounded-lg transition-colors ${
              severityFilter === 'all' || !severityFilter
                ? 'bg-white/[0.08] text-white font-medium'
                : 'text-slate-400 hover:text-slate-200'
            }`}
          >
            All ({data.stats?.total || data.total || 0})
          </button>
          <button
            onClick={() => onSeverityFilterChange('critical')}
            className={`px-2.5 py-1 rounded-lg transition-colors ${
              severityFilter === 'critical'
                ? 'bg-rose-500/15 text-rose-300 font-medium'
                : 'text-slate-400 hover:text-slate-200'
            }`}
          >
            Critical ({data.stats?.critical || 0})
          </button>
          <button
            onClick={() => onSeverityFilterChange('high')}
            className={`px-2.5 py-1 rounded-lg transition-colors ${
              severityFilter === 'high'
                ? 'bg-orange-500/15 text-orange-300 font-medium'
                : 'text-slate-400 hover:text-slate-200'
            }`}
          >
            High ({data.stats?.high || 0})
          </button>
          <button
            onClick={() => onSeverityFilterChange('medium')}
            className={`px-2.5 py-1 rounded-lg transition-colors ${
              severityFilter === 'medium'
                ? 'bg-amber-500/15 text-amber-300 font-medium'
                : 'text-slate-400 hover:text-slate-200'
            }`}
          >
            Medium ({data.stats?.medium || 0})
          </button>
          <button
            onClick={() => onSeverityFilterChange('low')}
            className={`px-2.5 py-1 rounded-lg transition-colors ${
              severityFilter === 'low'
                ? 'bg-blue-500/15 text-blue-300 font-medium'
                : 'text-slate-400 hover:text-slate-200'
            }`}
          >
            Low ({data.stats?.low || 0})
          </button>
          <button
            onClick={() => onSeverityFilterChange('info')}
            className={`px-2.5 py-1 rounded-lg transition-colors ${
              severityFilter === 'info'
                ? 'bg-white/[0.08] text-white font-medium'
                : 'text-slate-400 hover:text-slate-200'
            }`}
          >
            Info ({data.stats?.info || 0})
          </button>
        </div>

        {/* Search & Refresh */}
        <div className="flex items-center space-x-2">
          <form onSubmit={handleSearchSubmit} className="relative">
            <input
              type="text"
              placeholder="Search findings..."
              value={localSearch}
              onChange={(e) => setLocalSearch(e.target.value)}
              className="bg-white/[0.03] border border-white/[0.08] text-slate-100 text-xs rounded-lg pl-8 pr-3 py-1.5 focus:outline-none focus:border-indigo-500 font-mono w-48 sm:w-64 placeholder-slate-500"
            />
            <Search className="w-3.5 h-3.5 text-slate-500 absolute left-2.5 top-2.5" />
          </form>

          <button
            onClick={onRefresh}
            disabled={isLoading}
            className="p-1.5 rounded-lg text-slate-400 hover:text-white transition-colors"
          >
            <RefreshCw className={`w-3.5 h-3.5 ${isLoading ? 'animate-spin' : ''}`} />
          </button>
        </div>
      </div>

      {/* Findings Table */}
      <div className="bg-[#0e1420] border border-white/[0.06] rounded-xl overflow-hidden">
        {data.findings.length === 0 ? (
          <div className="p-12 text-center text-slate-500 text-xs">
            <ShieldCheck className="w-8 h-8 text-slate-600 mx-auto mb-2 opacity-60" />
            No vulnerability findings recorded yet for this target scope.
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left text-xs font-mono border-collapse">
              <thead>
                <tr className="border-b border-white/[0.04] bg-white/[0.02] text-slate-400 text-[10px] uppercase">
                  <th className="py-2.5 px-4 w-28">Severity</th>
                  <th className="py-2.5 px-4">Finding / Vulnerability</th>
                  <th className="py-2.5 px-4">Matched Endpoint</th>
                  <th className="py-2.5 px-4 w-24">Tool</th>
                  <th className="py-2.5 px-4 w-28 text-right">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-white/[0.03]">
                {data.findings.map((f) => {
                  const isExpanded = expandedId === f.id;

                  return (
                    <React.Fragment key={f.id}>
                      <tr className="hover:bg-white/[0.02] transition-colors">
                        <td className="py-3 px-4">{getSeverityBadge(f.severity)}</td>
                        <td className="py-3 px-4 font-semibold text-slate-200">
                          {f.name}
                        </td>
                        <td className="py-3 px-4 text-slate-300 truncate max-w-xs sm:max-w-md">
                          <a
                            href={f.matched_at}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="hover:text-indigo-400 flex items-center space-x-1"
                          >
                            <span className="truncate">{f.matched_at}</span>
                            <ExternalLink className="w-3 h-3 text-slate-500 shrink-0" />
                          </a>
                        </td>
                        <td className="py-3 px-4 text-slate-400 font-mono text-[11px]">
                          {f.tool_name}
                        </td>
                        <td className="py-3 px-4 text-right">
                          <button
                            onClick={() => setExpandedId(isExpanded ? null : f.id)}
                            className="p-1 rounded text-slate-400 hover:text-white hover:bg-white/[0.04] transition-colors"
                          >
                            {isExpanded ? <ChevronUp className="w-4 h-4" /> : <ChevronDown className="w-4 h-4" />}
                          </button>
                        </td>
                      </tr>

                      {/* Expanded Evidence View */}
                      {isExpanded && (
                        <tr className="bg-[#080c14]">
                          <td colSpan={5} className="p-4 space-y-3 border-t border-white/[0.04]">
                            <div className="flex items-center justify-between">
                              <span className="text-xs font-bold text-slate-300">Reproducible Evidence & Proof</span>
                              <div className="flex items-center space-x-2">
                                {f.curl_command && (
                                  <Button
                                    variant="secondary"
                                    size="xs"
                                    onClick={() => handleCopyCurl(f.id, f.curl_command!)}
                                  >
                                    {copiedId === f.id ? <Check className="w-3 h-3 text-emerald-400 mr-1" /> : <Copy className="w-3 h-3 mr-1" />}
                                    <span>Copy cURL</span>
                                  </Button>
                                )}
                                <Button
                                  variant="secondary"
                                  size="xs"
                                  onClick={() => handleCopyReport(f)}
                                >
                                  {copiedId === `report-${f.id}` ? <Check className="w-3 h-3 text-emerald-400 mr-1" /> : <Copy className="w-3 h-3 mr-1" />}
                                  <span>Copy Report</span>
                                </Button>
                              </div>
                            </div>

                            {f.description && (
                              <p className="text-xs text-slate-300 bg-white/[0.02] border border-white/[0.04] p-2.5 rounded-lg">
                                {f.description}
                              </p>
                            )}

                            {f.curl_command && (
                              <div>
                                <span className="text-[10px] uppercase text-slate-500 font-semibold block mb-1">Reproduction Command</span>
                                <pre className="bg-black/60 p-2.5 rounded-lg border border-white/[0.05] text-[11px] text-slate-300 overflow-x-auto whitespace-pre-wrap">
                                  {f.curl_command}
                                </pre>
                              </div>
                            )}

                            {f.extracted_results && (
                              <div>
                                <span className="text-[10px] uppercase text-slate-500 font-semibold block mb-1">Raw Evidence Transaction</span>
                                <pre className="bg-black/60 p-2.5 rounded-lg border border-white/[0.05] text-[11px] text-slate-300 overflow-x-auto max-h-56">
                                  {f.extracted_results}
                                </pre>
                              </div>
                            )}
                          </td>
                        </tr>
                      )}
                    </React.Fragment>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}

        {/* Pagination */}
        {data.total > data.limit && (
          <div className="px-4 py-3 bg-white/[0.01] border-t border-white/[0.04] flex items-center justify-between text-xs font-mono text-slate-400">
            <span>Total: {data.total} findings</span>
            <div className="flex items-center space-x-2">
              <span>Page {currentPage} of {totalPages}</span>
              <button
                onClick={() => onPageChange(currentPage - 1)}
                disabled={currentPage <= 1 || isLoading}
                className="p-1 rounded text-slate-400 hover:text-white disabled:opacity-30"
              >
                Prev
              </button>
              <button
                onClick={() => onPageChange(currentPage + 1)}
                disabled={currentPage >= totalPages || isLoading}
                className="p-1 rounded text-slate-400 hover:text-white disabled:opacity-30"
              >
                Next
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};
