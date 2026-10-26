import React, { useState } from 'react';
import { Search, ChevronDown, ChevronUp, Copy, Check, ExternalLink, RefreshCw, ShieldAlert } from 'lucide-react';
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
        return <span className="px-2 py-0.5 rounded text-[10px] font-mono font-bold uppercase tracking-wider bg-red-950/80 text-red-400 border border-red-700/80">Critical</span>;
      case 'high':
        return <span className="px-2 py-0.5 rounded text-[10px] font-mono font-bold uppercase tracking-wider bg-orange-950/80 text-orange-400 border border-orange-700/80">High</span>;
      case 'medium':
        return <span className="px-2 py-0.5 rounded text-[10px] font-mono font-bold uppercase tracking-wider bg-amber-950/80 text-amber-400 border border-amber-700/80">Medium</span>;
      case 'low':
        return <span className="px-2 py-0.5 rounded text-[10px] font-mono font-bold uppercase tracking-wider bg-blue-950/80 text-blue-400 border border-blue-700/80">Low</span>;
      default:
        return <span className="px-2 py-0.5 rounded text-[10px] font-mono font-bold uppercase tracking-wider bg-zinc-800 text-zinc-400 border border-zinc-700">Info</span>;
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
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-3 bg-[#161b22] p-3 rounded-lg border border-[#30363d]">
        {/* Severity Filter Chips */}
        <div className="flex flex-wrap items-center gap-1.5 text-xs font-mono">
          <button
            onClick={() => onSeverityFilterChange('all')}
            className={`px-2.5 py-1 rounded transition-colors ${
              severityFilter === 'all' || !severityFilter
                ? 'bg-[#21262d] text-zinc-100 font-bold border border-[#30363d]'
                : 'text-zinc-400 hover:text-zinc-200'
            }`}
          >
            All ({data.stats?.total || data.total || 0})
          </button>
          <button
            onClick={() => onSeverityFilterChange('critical')}
            className={`px-2.5 py-1 rounded transition-colors ${
              severityFilter === 'critical'
                ? 'bg-red-950/80 text-red-300 font-bold border border-red-800'
                : 'text-red-400/80 hover:text-red-300'
            }`}
          >
            Critical ({data.stats?.critical || 0})
          </button>
          <button
            onClick={() => onSeverityFilterChange('high')}
            className={`px-2.5 py-1 rounded transition-colors ${
              severityFilter === 'high'
                ? 'bg-orange-950/80 text-orange-300 font-bold border border-orange-800'
                : 'text-orange-400/80 hover:text-orange-300'
            }`}
          >
            High ({data.stats?.high || 0})
          </button>
          <button
            onClick={() => onSeverityFilterChange('medium')}
            className={`px-2.5 py-1 rounded transition-colors ${
              severityFilter === 'medium'
                ? 'bg-amber-950/80 text-amber-300 font-bold border border-amber-800'
                : 'text-amber-400/80 hover:text-amber-300'
            }`}
          >
            Medium ({data.stats?.medium || 0})
          </button>
          <button
            onClick={() => onSeverityFilterChange('low')}
            className={`px-2.5 py-1 rounded transition-colors ${
              severityFilter === 'low'
                ? 'bg-blue-950/80 text-blue-300 font-bold border border-blue-800'
                : 'text-blue-400/80 hover:text-blue-300'
            }`}
          >
            Low ({data.stats?.low || 0})
          </button>
          <button
            onClick={() => onSeverityFilterChange('info')}
            className={`px-2.5 py-1 rounded transition-colors ${
              severityFilter === 'info'
                ? 'bg-zinc-800 text-zinc-200 font-bold border border-zinc-700'
                : 'text-zinc-400 hover:text-zinc-300'
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
              className="bg-[#0d1117] border border-[#30363d] text-zinc-100 text-xs rounded pl-8 pr-3 py-1.5 focus:outline-none focus:border-[#58a6ff] font-mono w-48 sm:w-64"
            />
            <Search className="w-3.5 h-3.5 text-zinc-500 absolute left-2.5 top-2" />
          </form>

          <Button variant="secondary" size="sm" onClick={onRefresh} disabled={isLoading}>
            <RefreshCw className={`w-3.5 h-3.5 ${isLoading ? 'animate-spin' : ''}`} />
          </Button>
        </div>
      </div>

      {/* Findings Table */}
      <div className="bg-[#161b22] border border-[#30363d] rounded-lg overflow-hidden">
        {data.findings.length === 0 ? (
          <div className="p-12 text-center text-zinc-500 font-mono text-xs">
            <ShieldAlert className="w-8 h-8 text-zinc-600 mx-auto mb-2 opacity-60" />
            No vulnerability findings recorded yet for this target scope.
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left text-xs font-mono border-collapse">
              <thead>
                <tr className="border-b border-[#30363d] bg-[#0d1117] text-zinc-400">
                  <th className="py-2.5 px-4 w-24">Severity</th>
                  <th className="py-2.5 px-4">Finding / Vulnerability</th>
                  <th className="py-2.5 px-4">Matched Endpoint</th>
                  <th className="py-2.5 px-4 w-24">Tool</th>
                  <th className="py-2.5 px-4 w-28 text-right">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-[#30363d]/50">
                {data.findings.map((f) => {
                  const isExpanded = expandedId === f.id;

                  return (
                    <React.Fragment key={f.id}>
                      <tr className="hover:bg-[#21262d]/40 transition-colors">
                        <td className="py-3 px-4">{getSeverityBadge(f.severity)}</td>
                        <td className="py-3 px-4 font-semibold text-zinc-200">
                          {f.name}
                        </td>
                        <td className="py-3 px-4 text-zinc-300 truncate max-w-xs sm:max-w-md">
                          <a
                            href={f.matched_at}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="hover:text-blue-400 flex items-center space-x-1"
                          >
                            <span className="truncate">{f.matched_at}</span>
                            <ExternalLink className="w-3 h-3 text-zinc-500 shrink-0" />
                          </a>
                        </td>
                        <td className="py-3 px-4 text-zinc-400 font-mono text-[11px]">
                          {f.tool_name}
                        </td>
                        <td className="py-3 px-4 text-right">
                          <button
                            onClick={() => setExpandedId(isExpanded ? null : f.id)}
                            className="p-1 rounded text-zinc-400 hover:text-zinc-200 hover:bg-[#21262d] transition-colors"
                          >
                            {isExpanded ? <ChevronUp className="w-4 h-4" /> : <ChevronDown className="w-4 h-4" />}
                          </button>
                        </td>
                      </tr>

                      {/* Expandable Details Row */}
                      {isExpanded && (
                        <tr className="bg-[#0d1117]/60 border-t border-b border-[#30363d]">
                          <td colSpan={5} className="p-4 space-y-3">
                            {f.description && (
                              <div>
                                <h4 className="text-[11px] font-bold text-zinc-400 uppercase tracking-wider">Description</h4>
                                <p className="text-zinc-300 text-xs mt-1 leading-relaxed">{f.description}</p>
                              </div>
                            )}

                            {f.curl_command && (
                              <div>
                                <div className="flex items-center justify-between">
                                  <h4 className="text-[11px] font-bold text-zinc-400 uppercase tracking-wider">Reproduction Command</h4>
                                  <button
                                    onClick={() => handleCopyCurl(f.id, f.curl_command!)}
                                    className="text-[11px] text-zinc-400 hover:text-zinc-200 flex items-center space-x-1"
                                  >
                                    {copiedId === f.id ? <Check className="w-3 h-3 text-emerald-400" /> : <Copy className="w-3 h-3" />}
                                    <span>{copiedId === f.id ? 'Copied' : 'Copy cURL'}</span>
                                  </button>
                                </div>
                                <pre className="bg-[#161b22] p-2.5 rounded border border-[#30363d] text-[11px] text-zinc-300 mt-1 overflow-x-auto font-mono">
                                  {f.curl_command}
                                </pre>
                              </div>
                            )}

                            <div className="flex items-center justify-end space-x-2 pt-2 border-t border-[#30363d]/40">
                              <Button
                                variant="secondary"
                                size="sm"
                                onClick={() => handleCopyReport(f)}
                                className="text-xs font-mono"
                              >
                                {copiedId === `report-${f.id}` ? (
                                  <>
                                    <Check className="w-3.5 h-3.5 text-emerald-400 mr-1" />
                                    Report Copied!
                                  </>
                                ) : (
                                  <>
                                    <Copy className="w-3.5 h-3.5 mr-1" />
                                    Copy HackerOne Report
                                  </>
                                )}
                              </Button>
                            </div>
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

        {/* Pagination Bar */}
        {totalPages > 1 && (
          <div className="flex items-center justify-between p-3 border-t border-[#30363d] bg-[#0d1117] text-xs font-mono text-zinc-400">
            <span>Page {currentPage} of {totalPages} ({data.total} findings)</span>
            <div className="flex items-center space-x-1">
              <button
                disabled={currentPage <= 1}
                onClick={() => onPageChange(currentPage - 1)}
                className="px-2 py-1 rounded bg-[#21262d] text-zinc-200 disabled:opacity-40"
              >
                Prev
              </button>
              <button
                disabled={currentPage >= totalPages}
                onClick={() => onPageChange(currentPage + 1)}
                className="px-2 py-1 rounded bg-[#21262d] text-zinc-200 disabled:opacity-40"
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
